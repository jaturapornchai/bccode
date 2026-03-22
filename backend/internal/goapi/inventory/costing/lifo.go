package costing

import (
	"context"
	"database/sql"
	"fmt"

	inv "smlcloudplatform/internal/goapi/inventory/models"
)

// LIFOEngine — Last In, First Out
// สินค้าที่ซื้อล่าสุดตัดออกก่อน — ใช้ cost layer เดียวกับ FIFO แต่ตัดจาก layer ใหม่สุด
// หมายเหตุ: IFRS/TFRS ห้ามใช้ LIFO → แจ้งเตือนผู้ใช้เมื่อเลือก
type LIFOEngine struct{}

func (e *LIFOEngine) Method() string { return inv.CostingMethodLIFO }

func (e *LIFOEngine) ProcessReceipt(ctx context.Context, tx *sql.Tx, params inv.ReceiptParams) (*inv.CostTransactionResult, error) {
	// Receipt เหมือน FIFO ทุกประการ (สร้าง layer ใหม่)
	fifo := &FIFOEngine{}
	result, err := fifo.ProcessReceipt(ctx, tx, params)
	if err != nil {
		return nil, err
	}
	result.Transaction.CostingMethodUsed = inv.CostingMethodLIFO
	return result, nil
}

func (e *LIFOEngine) ProcessIssue(ctx context.Context, tx *sql.Tx, params inv.IssueParams) (*inv.CostTransactionResult, error) {
	balance, err := getOrCreateBalance(ctx, tx, params.ShopID, params.ItemCode, params.Barcode, params.WhCode, params.LocationCode)
	if err != nil {
		return nil, err
	}

	if balance.CurrentQty < params.Qty {
		return nil, fmt.Errorf("สต็อกไม่พอ: คงเหลือ %.4f ต้องการ %.4f", balance.CurrentQty, params.Qty)
	}

	// ตัด layers จากใหม่สุดก่อน (LIFO)
	totalCost, err := e.consumeLayersLIFO(ctx, tx, params.ShopID, params.ItemCode, params.WhCode, params.LocationCode, params.Qty)
	if err != nil {
		return nil, err
	}

	unitCost := totalCost / params.Qty
	newQty := balance.CurrentQty - params.Qty
	newTotalValue := balance.CurrentTotalValue - totalCost
	if newQty <= 0 {
		newQty = 0
		newTotalValue = 0
	}
	newAvgCost := averageCostCalc(newTotalValue, newQty)

	balance.CurrentQty = newQty
	balance.CurrentAvgCost = newAvgCost
	balance.CurrentTotalValue = newTotalValue

	if err := updateBalance(ctx, tx, balance); err != nil {
		return nil, err
	}

	ct := &inv.InventoryCostTransaction{
		ShopID: params.ShopID, ItemCode: params.ItemCode, Barcode: params.Barcode,
		WhCode: params.WhCode, LocationCode: params.LocationCode,
		TransactionType: inv.TxTypeSalesIssue, TransFlag: params.TransFlag,
		RefDocType: params.RefDocType, RefDocNo: params.RefDocNo,
		Qty: -params.Qty, UnitCost: unitCost, TotalCost: -totalCost,
		LotNumber: params.LotNumber,
		BalanceQty: newQty, BalanceAvgCost: newAvgCost, BalanceTotalValue: newTotalValue,
		CostingMethodUsed: inv.CostingMethodLIFO,
		TransactionDate: params.TransactionDate, CreatedBy: params.CreatedBy,
	}
	if err := insertCostTransaction(ctx, tx, ct); err != nil {
		return nil, err
	}

	return &inv.CostTransactionResult{
		Transaction: ct, BalanceQty: newQty, BalanceAvgCost: newAvgCost, TotalValue: newTotalValue,
	}, nil
}

// consumeLayersLIFO — ตัด layers จากใหม่สุดก่อน (DESC)
func (e *LIFOEngine) consumeLayersLIFO(ctx context.Context, tx *sql.Tx, shopID, itemCode, whCode, locationCode string, qtyNeeded float64) (float64, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT id, remainingqty, totalunitcost
		 FROM inventory_cost_layers
		 WHERE shopid = $1 AND itemcode = $2 AND whcode = $3 AND locationcode = $4 AND remainingqty > 0
		 ORDER BY receiveddate DESC, id DESC
		 FOR UPDATE`,
		shopID, itemCode, whCode, locationCode,
	)
	if err != nil {
		return 0, fmt.Errorf("ดึง cost layers ไม่สำเร็จ: %w", err)
	}
	defer rows.Close()

	var totalCost float64
	remaining := qtyNeeded

	for rows.Next() && remaining > 0 {
		var layerID int64
		var layerQty, unitCost float64
		if err := rows.Scan(&layerID, &layerQty, &unitCost); err != nil {
			return 0, err
		}
		consume := remaining
		if consume > layerQty {
			consume = layerQty
		}
		totalCost += consume * unitCost
		remaining -= consume

		_, err := tx.ExecContext(ctx,
			`UPDATE inventory_cost_layers SET remainingqty = $1, updatedat = NOW() WHERE id = $2`,
			layerQty-consume, layerID,
		)
		if err != nil {
			return 0, err
		}
	}
	if remaining > 0 {
		return 0, fmt.Errorf("cost layers ไม่พอ: ยังขาดอีก %.4f หน่วย", remaining)
	}
	return totalCost, nil
}

func (e *LIFOEngine) ProcessSalesReturn(ctx context.Context, tx *sql.Tx, params inv.SalesReturnParams) (*inv.CostTransactionResult, error) {
	fifo := &FIFOEngine{}
	result, err := fifo.ProcessSalesReturn(ctx, tx, params)
	if err != nil {
		return nil, err
	}
	result.Transaction.CostingMethodUsed = inv.CostingMethodLIFO
	return result, nil
}

func (e *LIFOEngine) ProcessPurchaseReturn(ctx context.Context, tx *sql.Tx, params inv.PurchaseReturnParams) (*inv.CostTransactionResult, error) {
	// Purchase return สำหรับ LIFO ตัด layer ใหม่สุด
	balance, err := getOrCreateBalance(ctx, tx, params.ShopID, params.ItemCode, params.Barcode, params.WhCode, params.LocationCode)
	if err != nil {
		return nil, err
	}

	totalCost, err := e.consumeLayersLIFO(ctx, tx, params.ShopID, params.ItemCode, params.WhCode, params.LocationCode, params.Qty)
	if err != nil {
		return nil, err
	}
	unitCost := totalCost / params.Qty

	newQty := balance.CurrentQty - params.Qty
	newTotalValue := balance.CurrentTotalValue - totalCost
	if newQty <= 0 {
		newQty = 0
		newTotalValue = 0
	}
	newAvgCost := averageCostCalc(newTotalValue, newQty)
	balance.CurrentQty = newQty
	balance.CurrentAvgCost = newAvgCost
	balance.CurrentTotalValue = newTotalValue

	if err := updateBalance(ctx, tx, balance); err != nil {
		return nil, err
	}

	ct := &inv.InventoryCostTransaction{
		ShopID: params.ShopID, ItemCode: params.ItemCode, Barcode: params.Barcode,
		WhCode: params.WhCode, LocationCode: params.LocationCode,
		TransactionType: inv.TxTypePurchaseReturn, TransFlag: params.TransFlag,
		RefDocType: params.RefDocType, RefDocNo: params.RefDocNo,
		Qty: -params.Qty, UnitCost: unitCost, TotalCost: -totalCost,
		BalanceQty: newQty, BalanceAvgCost: newAvgCost, BalanceTotalValue: newTotalValue,
		CostingMethodUsed: inv.CostingMethodLIFO,
		TransactionDate: params.TransactionDate, CreatedBy: params.CreatedBy,
	}
	if err := insertCostTransaction(ctx, tx, ct); err != nil {
		return nil, err
	}

	return &inv.CostTransactionResult{Transaction: ct, BalanceQty: newQty, BalanceAvgCost: newAvgCost, TotalValue: newTotalValue}, nil
}

func (e *LIFOEngine) ProcessAdjustment(ctx context.Context, tx *sql.Tx, params inv.AdjustmentParams) (*inv.CostTransactionResult, error) {
	fifo := &FIFOEngine{}
	result, err := fifo.ProcessAdjustment(ctx, tx, params)
	if err != nil {
		return nil, err
	}
	result.Transaction.CostingMethodUsed = inv.CostingMethodLIFO
	return result, nil
}

func (e *LIFOEngine) GetCurrentValuation(ctx context.Context, tx *sql.Tx, shopID, itemCode, whCode string) (*inv.StockValuation, error) {
	fifo := &FIFOEngine{}
	val, err := fifo.GetCurrentValuation(ctx, tx, shopID, itemCode, whCode)
	if err != nil {
		return nil, err
	}
	val.CostingMethod = inv.CostingMethodLIFO
	return val, nil
}
