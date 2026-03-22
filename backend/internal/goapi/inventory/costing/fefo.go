package costing

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	inv "smlcloudplatform/internal/goapi/inventory/models"
)

// FEFOEngine — First Expired, First Out
// ตัดสินค้าที่หมดอายุเร็วที่สุดก่อน — เรียงตาม expiry_date
// บังคับ: ต้องมี expiry_date + lot_number ทุก layer
type FEFOEngine struct{}

func (e *FEFOEngine) Method() string { return inv.CostingMethodFEFO }

func (e *FEFOEngine) ProcessReceipt(ctx context.Context, tx *sql.Tx, params inv.ReceiptParams) (*inv.CostTransactionResult, error) {
	// validate: FEFO ต้องมี expiry_date
	if params.ExpiryDate == nil {
		return nil, fmt.Errorf("FEFO ต้องระบุ expiry_date ทุกครั้งที่รับสินค้าเข้า")
	}
	if params.LotNumber == "" {
		return nil, fmt.Errorf("FEFO ต้องระบุ lot_number ทุกครั้งที่รับสินค้าเข้า")
	}

	balance, err := getOrCreateBalance(ctx, tx, params.ShopID, params.ItemCode, params.Barcode, params.WhCode, params.LocationCode)
	if err != nil {
		return nil, err
	}

	totalCost := params.Qty * params.UnitCost

	layer := &inv.InventoryCostLayer{
		ShopID: params.ShopID, ItemCode: params.ItemCode, Barcode: params.Barcode,
		WhCode: params.WhCode, LocationCode: params.LocationCode,
		LayerType: inv.LayerTypePurchase, RefDocType: params.RefDocType, RefDocNo: params.RefDocNo,
		OriginalQty: params.Qty, RemainingQty: params.Qty,
		UnitCost: params.UnitCost, TotalUnitCost: params.UnitCost,
		LotNumber: params.LotNumber, ExpiryDate: params.ExpiryDate,
		QualityStatus: inv.QualityStatusApproved, ReceivedDate: params.ReceivedDate,
	}
	if err := insertCostLayer(ctx, tx, layer); err != nil {
		return nil, err
	}

	newQty := balance.CurrentQty + params.Qty
	newTotalValue := balance.CurrentTotalValue + totalCost
	newAvgCost := averageCostCalc(newTotalValue, newQty)

	balance.CurrentQty = newQty
	balance.CurrentAvgCost = newAvgCost
	balance.CurrentTotalValue = newTotalValue
	balance.LastPurchaseCost = params.UnitCost
	now := time.Now()
	balance.LastPurchaseDate = &now

	if err := updateBalance(ctx, tx, balance); err != nil {
		return nil, err
	}

	ct := &inv.InventoryCostTransaction{
		ShopID: params.ShopID, ItemCode: params.ItemCode, Barcode: params.Barcode,
		WhCode: params.WhCode, LocationCode: params.LocationCode,
		TransactionType: inv.TxTypePurchaseReceipt, TransFlag: params.TransFlag,
		RefDocType: params.RefDocType, RefDocNo: params.RefDocNo,
		Qty: params.Qty, UnitCost: params.UnitCost, TotalCost: totalCost,
		LotNumber: params.LotNumber, ExpiryDate: params.ExpiryDate, CostLayerID: &layer.ID,
		BalanceQty: newQty, BalanceAvgCost: newAvgCost, BalanceTotalValue: newTotalValue,
		CostingMethodUsed: inv.CostingMethodFEFO,
		TransactionDate: params.ReceivedDate, CreatedBy: params.CreatedBy,
	}
	if err := insertCostTransaction(ctx, tx, ct); err != nil {
		return nil, err
	}

	return &inv.CostTransactionResult{
		Transaction: ct, CostLayer: layer,
		BalanceQty: newQty, BalanceAvgCost: newAvgCost, TotalValue: newTotalValue,
	}, nil
}

func (e *FEFOEngine) ProcessIssue(ctx context.Context, tx *sql.Tx, params inv.IssueParams) (*inv.CostTransactionResult, error) {
	balance, err := getOrCreateBalance(ctx, tx, params.ShopID, params.ItemCode, params.Barcode, params.WhCode, params.LocationCode)
	if err != nil {
		return nil, err
	}
	if balance.CurrentQty < params.Qty {
		return nil, fmt.Errorf("สต็อกไม่พอ: คงเหลือ %.4f ต้องการ %.4f", balance.CurrentQty, params.Qty)
	}

	// ตัด layers เรียงตาม expiry_date (หมดอายุเร็วสุดก่อน)
	totalCost, err := e.consumeLayersFEFO(ctx, tx, params.ShopID, params.ItemCode, params.WhCode, params.LocationCode, params.Qty)
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
		CostingMethodUsed: inv.CostingMethodFEFO,
		TransactionDate: params.TransactionDate, CreatedBy: params.CreatedBy,
	}
	if err := insertCostTransaction(ctx, tx, ct); err != nil {
		return nil, err
	}

	return &inv.CostTransactionResult{
		Transaction: ct, BalanceQty: newQty, BalanceAvgCost: newAvgCost, TotalValue: newTotalValue,
	}, nil
}

// consumeLayersFEFO — ตัด layers เรียงตาม expiry_date ASC (หมดอายุเร็วสุดก่อน)
func (e *FEFOEngine) consumeLayersFEFO(ctx context.Context, tx *sql.Tx, shopID, itemCode, whCode, locationCode string, qtyNeeded float64) (float64, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT id, remainingqty, totalunitcost, expirydate
		 FROM inventory_cost_layers
		 WHERE shopid = $1 AND itemcode = $2 AND whcode = $3 AND locationcode = $4 AND remainingqty > 0
		 ORDER BY expirydate ASC, receiveddate ASC, id ASC
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
		var expiryDate *time.Time
		if err := rows.Scan(&layerID, &layerQty, &unitCost, &expiryDate); err != nil {
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

func (e *FEFOEngine) ProcessSalesReturn(ctx context.Context, tx *sql.Tx, params inv.SalesReturnParams) (*inv.CostTransactionResult, error) {
	fifo := &FIFOEngine{}
	result, err := fifo.ProcessSalesReturn(ctx, tx, params)
	if err != nil {
		return nil, err
	}
	result.Transaction.CostingMethodUsed = inv.CostingMethodFEFO
	return result, nil
}

func (e *FEFOEngine) ProcessPurchaseReturn(ctx context.Context, tx *sql.Tx, params inv.PurchaseReturnParams) (*inv.CostTransactionResult, error) {
	fifo := &FIFOEngine{}
	result, err := fifo.ProcessPurchaseReturn(ctx, tx, params)
	if err != nil {
		return nil, err
	}
	result.Transaction.CostingMethodUsed = inv.CostingMethodFEFO
	return result, nil
}

func (e *FEFOEngine) ProcessAdjustment(ctx context.Context, tx *sql.Tx, params inv.AdjustmentParams) (*inv.CostTransactionResult, error) {
	if params.IsIncrease && params.ExpiryDate == nil {
		return nil, fmt.Errorf("FEFO ต้องระบุ expiry_date เมื่อเพิ่มสต็อก")
	}
	fifo := &FIFOEngine{}
	result, err := fifo.ProcessAdjustment(ctx, tx, params)
	if err != nil {
		return nil, err
	}
	result.Transaction.CostingMethodUsed = inv.CostingMethodFEFO
	return result, nil
}

func (e *FEFOEngine) GetCurrentValuation(ctx context.Context, tx *sql.Tx, shopID, itemCode, whCode string) (*inv.StockValuation, error) {
	fifo := &FIFOEngine{}
	val, err := fifo.GetCurrentValuation(ctx, tx, shopID, itemCode, whCode)
	if err != nil {
		return nil, err
	}
	val.CostingMethod = inv.CostingMethodFEFO
	return val, nil
}
