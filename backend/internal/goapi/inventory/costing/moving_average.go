package costing

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	inv "smlcloudplatform/internal/goapi/inventory/models"
)

// MovingAverageEngine — ต้นทุนเฉลี่ยเคลื่อนที่ (Moving Weighted Average)
// คำนวณต้นทุนเฉลี่ยใหม่ทุกครั้งที่รับสินค้าเข้า
// สูตร: new_avg = (old_qty × old_avg + received_qty × received_cost) / (old_qty + received_qty)
type MovingAverageEngine struct{}

func (e *MovingAverageEngine) Method() string { return inv.CostingMethodMovingAverage }

// ProcessReceipt — รับสินค้าเข้า → คำนวณ avg ใหม่
func (e *MovingAverageEngine) ProcessReceipt(ctx context.Context, tx *sql.Tx, params inv.ReceiptParams) (*inv.CostTransactionResult, error) {
	balance, err := getOrCreateBalance(ctx, tx, params.HoldingCode, params.ItemCode, params.Barcode, params.WhCode, params.LocationCode)
	if err != nil {
		return nil, err
	}

	totalCost := params.Qty * params.UnitCost

	// คำนวณ avg ใหม่
	newQty := balance.CurrentQty + params.Qty
	newTotalValue := balance.CurrentTotalValue + totalCost
	newAvgCost := averageCostCalc(newTotalValue, newQty)

	// อัพเดท balance
	balance.CurrentQty = newQty
	balance.CurrentAvgCost = newAvgCost
	balance.CurrentTotalValue = newTotalValue
	balance.LastPurchaseCost = params.UnitCost
	now := time.Now()
	balance.LastPurchaseDate = &now

	if err := updateBalance(ctx, tx, balance); err != nil {
		return nil, err
	}

	// บันทึก transaction
	ct := &inv.InventoryCostTransaction{
		HoldingCode:       params.HoldingCode,
		ItemCode:          params.ItemCode,
		Barcode:           params.Barcode,
		WhCode:            params.WhCode,
		LocationCode:      params.LocationCode,
		TransactionType:   inv.TxTypePurchaseReceipt,
		TransFlag:         params.TransFlag,
		RefDocType:        params.RefDocType,
		RefDocNo:          params.RefDocNo,
		Qty:               params.Qty,
		UnitCost:          params.UnitCost,
		TotalCost:         totalCost,
		LotNumber:         params.LotNumber,
		ExpiryDate:        params.ExpiryDate,
		BalanceQty:        newQty,
		BalanceAvgCost:    newAvgCost,
		BalanceTotalValue: newTotalValue,
		CostingMethodUsed: inv.CostingMethodMovingAverage,
		TransactionDate:   params.ReceivedDate,
		CreatedBy:         params.CreatedBy,
	}
	if err := insertCostTransaction(ctx, tx, ct); err != nil {
		return nil, err
	}

	return &inv.CostTransactionResult{
		Transaction:    ct,
		BalanceQty:     newQty,
		BalanceAvgCost: newAvgCost,
		TotalValue:     newTotalValue,
	}, nil
}

// ProcessIssue — ตัดสินค้าออก → ใช้ avg ปัจจุบัน
func (e *MovingAverageEngine) ProcessIssue(ctx context.Context, tx *sql.Tx, params inv.IssueParams) (*inv.CostTransactionResult, error) {
	balance, err := getOrCreateBalance(ctx, tx, params.HoldingCode, params.ItemCode, params.Barcode, params.WhCode, params.LocationCode)
	if err != nil {
		return nil, err
	}

	if balance.CurrentQty < params.Qty {
		return nil, fmt.Errorf("สต็อกไม่พอ: คงเหลือ %.4f ต้องการ %.4f", balance.CurrentQty, params.Qty)
	}

	unitCost := balance.CurrentAvgCost
	totalCost := params.Qty * unitCost

	newQty := balance.CurrentQty - params.Qty
	newTotalValue := balance.CurrentTotalValue - totalCost

	// ถ้าของหมด → reset เป็น 0
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
		HoldingCode:       params.HoldingCode,
		ItemCode:          params.ItemCode,
		Barcode:           params.Barcode,
		WhCode:            params.WhCode,
		LocationCode:      params.LocationCode,
		TransactionType:   inv.TxTypeSalesIssue,
		TransFlag:         params.TransFlag,
		RefDocType:        params.RefDocType,
		RefDocNo:          params.RefDocNo,
		Qty:               -params.Qty,
		UnitCost:          unitCost,
		TotalCost:         -totalCost,
		LotNumber:         params.LotNumber,
		BalanceQty:        newQty,
		BalanceAvgCost:    newAvgCost,
		BalanceTotalValue: newTotalValue,
		CostingMethodUsed: inv.CostingMethodMovingAverage,
		TransactionDate:   params.TransactionDate,
		CreatedBy:         params.CreatedBy,
	}
	if err := insertCostTransaction(ctx, tx, ct); err != nil {
		return nil, err
	}

	return &inv.CostTransactionResult{
		Transaction:    ct,
		BalanceQty:     newQty,
		BalanceAvgCost: newAvgCost,
		TotalValue:     newTotalValue,
	}, nil
}

// ProcessSalesReturn — รับคืนจากลูกค้า → เพิ่ม qty ด้วย original cost
func (e *MovingAverageEngine) ProcessSalesReturn(ctx context.Context, tx *sql.Tx, params inv.SalesReturnParams) (*inv.CostTransactionResult, error) {
	balance, err := getOrCreateBalance(ctx, tx, params.HoldingCode, params.ItemCode, params.Barcode, params.WhCode, params.LocationCode)
	if err != nil {
		return nil, err
	}

	unitCost := params.OriginalCost
	if unitCost <= 0 {
		unitCost = balance.CurrentAvgCost
	}
	totalCost := params.Qty * unitCost

	newQty := balance.CurrentQty + params.Qty
	newTotalValue := balance.CurrentTotalValue + totalCost
	newAvgCost := averageCostCalc(newTotalValue, newQty)

	balance.CurrentQty = newQty
	balance.CurrentAvgCost = newAvgCost
	balance.CurrentTotalValue = newTotalValue

	if err := updateBalance(ctx, tx, balance); err != nil {
		return nil, err
	}

	ct := &inv.InventoryCostTransaction{
		HoldingCode:       params.HoldingCode,
		ItemCode:          params.ItemCode,
		Barcode:           params.Barcode,
		WhCode:            params.WhCode,
		LocationCode:      params.LocationCode,
		TransactionType:   inv.TxTypeSalesReturn,
		TransFlag:         params.TransFlag,
		RefDocType:        params.RefDocType,
		RefDocNo:          params.RefDocNo,
		Qty:               params.Qty,
		UnitCost:          unitCost,
		TotalCost:         totalCost,
		LotNumber:         params.LotNumber,
		BalanceQty:        newQty,
		BalanceAvgCost:    newAvgCost,
		BalanceTotalValue: newTotalValue,
		CostingMethodUsed: inv.CostingMethodMovingAverage,
		TransactionDate:   params.TransactionDate,
		CreatedBy:         params.CreatedBy,
	}
	if err := insertCostTransaction(ctx, tx, ct); err != nil {
		return nil, err
	}

	return &inv.CostTransactionResult{
		Transaction:    ct,
		BalanceQty:     newQty,
		BalanceAvgCost: newAvgCost,
		TotalValue:     newTotalValue,
	}, nil
}

// ProcessPurchaseReturn — ส่งคืน supplier → ตัด qty ด้วย avg ปัจจุบัน
func (e *MovingAverageEngine) ProcessPurchaseReturn(ctx context.Context, tx *sql.Tx, params inv.PurchaseReturnParams) (*inv.CostTransactionResult, error) {
	balance, err := getOrCreateBalance(ctx, tx, params.HoldingCode, params.ItemCode, params.Barcode, params.WhCode, params.LocationCode)
	if err != nil {
		return nil, err
	}

	unitCost := balance.CurrentAvgCost
	totalCost := params.Qty * unitCost

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
		HoldingCode:       params.HoldingCode,
		ItemCode:          params.ItemCode,
		Barcode:           params.Barcode,
		WhCode:            params.WhCode,
		LocationCode:      params.LocationCode,
		TransactionType:   inv.TxTypePurchaseReturn,
		TransFlag:         params.TransFlag,
		RefDocType:        params.RefDocType,
		RefDocNo:          params.RefDocNo,
		Qty:               -params.Qty,
		UnitCost:          unitCost,
		TotalCost:         -totalCost,
		BalanceQty:        newQty,
		BalanceAvgCost:    newAvgCost,
		BalanceTotalValue: newTotalValue,
		CostingMethodUsed: inv.CostingMethodMovingAverage,
		TransactionDate:   params.TransactionDate,
		CreatedBy:         params.CreatedBy,
	}
	if err := insertCostTransaction(ctx, tx, ct); err != nil {
		return nil, err
	}

	return &inv.CostTransactionResult{
		Transaction:    ct,
		BalanceQty:     newQty,
		BalanceAvgCost: newAvgCost,
		TotalValue:     newTotalValue,
	}, nil
}

// ProcessAdjustment — ปรับปรุง stock (เพิ่ม/ลด)
func (e *MovingAverageEngine) ProcessAdjustment(ctx context.Context, tx *sql.Tx, params inv.AdjustmentParams) (*inv.CostTransactionResult, error) {
	balance, err := getOrCreateBalance(ctx, tx, params.HoldingCode, params.ItemCode, params.Barcode, params.WhCode, params.LocationCode)
	if err != nil {
		return nil, err
	}

	var txType string
	var qty, totalCost, newQty, newTotalValue float64

	if params.IsIncrease {
		txType = inv.TxTypeAdjustmentIn
		unitCost := params.UnitCost
		if unitCost <= 0 {
			unitCost = balance.CurrentAvgCost
		}
		totalCost = params.Qty * unitCost
		qty = params.Qty

		newQty = balance.CurrentQty + params.Qty
		newTotalValue = balance.CurrentTotalValue + totalCost
	} else {
		txType = inv.TxTypeAdjustmentOut
		unitCost := balance.CurrentAvgCost
		totalCost = params.Qty * unitCost
		qty = -params.Qty

		newQty = balance.CurrentQty - params.Qty
		newTotalValue = balance.CurrentTotalValue - totalCost
		if newQty <= 0 {
			newQty = 0
			newTotalValue = 0
		}
	}
	newAvgCost := averageCostCalc(newTotalValue, newQty)

	balance.CurrentQty = newQty
	balance.CurrentAvgCost = newAvgCost
	balance.CurrentTotalValue = newTotalValue

	if err := updateBalance(ctx, tx, balance); err != nil {
		return nil, err
	}

	ct := &inv.InventoryCostTransaction{
		HoldingCode:       params.HoldingCode,
		ItemCode:          params.ItemCode,
		Barcode:           params.Barcode,
		WhCode:            params.WhCode,
		LocationCode:      params.LocationCode,
		TransactionType:   txType,
		TransFlag:         params.TransFlag,
		RefDocType:        params.RefDocType,
		RefDocNo:          params.RefDocNo,
		Qty:               qty,
		UnitCost:          params.UnitCost,
		TotalCost:         totalCost * float64(boolToSign(params.IsIncrease)),
		LotNumber:         params.LotNumber,
		ExpiryDate:        params.ExpiryDate,
		BalanceQty:        newQty,
		BalanceAvgCost:    newAvgCost,
		BalanceTotalValue: newTotalValue,
		CostingMethodUsed: inv.CostingMethodMovingAverage,
		TransactionDate:   params.TransactionDate,
		CreatedBy:         params.CreatedBy,
	}
	if err := insertCostTransaction(ctx, tx, ct); err != nil {
		return nil, err
	}

	return &inv.CostTransactionResult{
		Transaction:    ct,
		BalanceQty:     newQty,
		BalanceAvgCost: newAvgCost,
		TotalValue:     newTotalValue,
	}, nil
}

// GetCurrentValuation — ดูมูลค่าสินค้าปัจจุบัน
func (e *MovingAverageEngine) GetCurrentValuation(ctx context.Context, tx *sql.Tx, holdingCode, itemCode, whCode string) (*inv.StockValuation, error) {
	var val inv.StockValuation
	var err error
	if whCode == "" {
		err = tx.QueryRowContext(ctx,
			`SELECT itemcode, '' AS whcode,
			        COALESCE(SUM(currentqty), 0),
			        CASE WHEN COALESCE(SUM(currentqty), 0) = 0 THEN 0 ELSE COALESCE(SUM(currenttotalvalue), 0) / SUM(currentqty) END,
			        COALESCE(SUM(currenttotalvalue), 0)
			 FROM inventory_stock_balances
			 WHERE holding_code = $1 AND itemcode = $2
			 GROUP BY itemcode`,
			holdingCode, itemCode,
		).Scan(&val.ItemCode, &val.WhCode, &val.CurrentQty, &val.AverageCost, &val.TotalValue)
	} else {
		err = tx.QueryRowContext(ctx,
			`SELECT itemcode, whcode,
			        COALESCE(SUM(currentqty), 0),
			        CASE WHEN COALESCE(SUM(currentqty), 0) = 0 THEN 0 ELSE COALESCE(SUM(currenttotalvalue), 0) / SUM(currentqty) END,
			        COALESCE(SUM(currenttotalvalue), 0)
			 FROM inventory_stock_balances
			 WHERE holding_code = $1 AND itemcode = $2 AND whcode = $3
			 GROUP BY itemcode, whcode`,
			holdingCode, itemCode, whCode,
		).Scan(&val.ItemCode, &val.WhCode, &val.CurrentQty, &val.AverageCost, &val.TotalValue)
	}
	if err == sql.ErrNoRows {
		return &inv.StockValuation{ItemCode: itemCode, WhCode: whCode, CostingMethod: inv.CostingMethodMovingAverage}, nil
	}
	if err != nil {
		return nil, err
	}
	val.CostingMethod = inv.CostingMethodMovingAverage
	return &val, nil
}

func boolToSign(increase bool) int {
	if increase {
		return 1
	}
	return -1
}
