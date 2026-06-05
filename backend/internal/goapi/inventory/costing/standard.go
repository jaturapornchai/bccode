package costing

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	inv "smlcloudplatform/internal/goapi/inventory/models"
)

// StandardEngine — ต้นทุนมาตรฐาน (Standard Cost)
// กำหนดต้นทุนมาตรฐานล่วงหน้าต่อสินค้า
// รับเข้า/ขาย → บันทึกด้วย standard cost เสมอ
// ผลต่างจริง vs มาตรฐาน → บันทึกเป็น Variance
type StandardEngine struct{}

func (e *StandardEngine) Method() string { return inv.CostingMethodStandard }

func (e *StandardEngine) ProcessReceipt(ctx context.Context, tx *sql.Tx, params inv.ReceiptParams) (*inv.CostTransactionResult, error) {
	balance, err := getOrCreateBalance(ctx, tx, params.HoldingCode, params.ItemCode, params.Barcode, params.WhCode, params.LocationCode)
	if err != nil {
		return nil, err
	}

	// ดึง standard cost จาก product config
	standardCost, err := getStandardCost(ctx, tx, params.HoldingCode, params.ItemCode)
	if err != nil {
		return nil, err
	}

	// บันทึกที่ standard cost
	totalCost := params.Qty * standardCost
	newQty := balance.CurrentQty + params.Qty
	newTotalValue := balance.CurrentTotalValue + totalCost
	newAvgCost := standardCost // Standard Cost ใช้ค่าเดียวเสมอ

	balance.CurrentQty = newQty
	balance.CurrentAvgCost = newAvgCost
	balance.CurrentTotalValue = newTotalValue
	balance.LastPurchaseCost = params.UnitCost
	now := time.Now()
	balance.LastPurchaseDate = &now

	if err := updateBalance(ctx, tx, balance); err != nil {
		return nil, err
	}

	// บันทึก Purchase Price Variance (PPV)
	var variance *inv.InventoryVariance
	if params.UnitCost != standardCost {
		varianceAmount := (params.UnitCost - standardCost) * params.Qty
		variance = &inv.InventoryVariance{
			HoldingCode:     params.HoldingCode,
			ItemCode:        params.ItemCode,
			WhCode:          params.WhCode,
			VarianceType:    inv.VariancePurchasePrice,
			RefDocType:      params.RefDocType,
			RefDocNo:        params.RefDocNo,
			StandardCost:    standardCost,
			ActualCost:      params.UnitCost,
			Qty:             params.Qty,
			VarianceAmount:  varianceAmount,
			TransactionDate: params.ReceivedDate,
		}
		if err := insertVariance(ctx, tx, variance); err != nil {
			return nil, err
		}
	}

	ct := &inv.InventoryCostTransaction{
		HoldingCode: params.HoldingCode, ItemCode: params.ItemCode, Barcode: params.Barcode,
		WhCode: params.WhCode, LocationCode: params.LocationCode,
		TransactionType: inv.TxTypePurchaseReceipt, TransFlag: params.TransFlag,
		RefDocType: params.RefDocType, RefDocNo: params.RefDocNo,
		Qty: params.Qty, UnitCost: standardCost, TotalCost: totalCost,
		BalanceQty: newQty, BalanceAvgCost: newAvgCost, BalanceTotalValue: newTotalValue,
		CostingMethodUsed: inv.CostingMethodStandard,
		TransactionDate:   params.ReceivedDate, CreatedBy: params.CreatedBy,
	}
	if err := insertCostTransaction(ctx, tx, ct); err != nil {
		return nil, err
	}

	return &inv.CostTransactionResult{
		Transaction: ct, Variance: variance,
		BalanceQty: newQty, BalanceAvgCost: newAvgCost, TotalValue: newTotalValue,
	}, nil
}

func (e *StandardEngine) ProcessIssue(ctx context.Context, tx *sql.Tx, params inv.IssueParams) (*inv.CostTransactionResult, error) {
	balance, err := getOrCreateBalance(ctx, tx, params.HoldingCode, params.ItemCode, params.Barcode, params.WhCode, params.LocationCode)
	if err != nil {
		return nil, err
	}
	if balance.CurrentQty < params.Qty {
		return nil, fmt.Errorf("สต็อกไม่พอ: คงเหลือ %.4f ต้องการ %.4f", balance.CurrentQty, params.Qty)
	}

	standardCost, err := getStandardCost(ctx, tx, params.HoldingCode, params.ItemCode)
	if err != nil {
		return nil, err
	}

	totalCost := params.Qty * standardCost
	newQty := balance.CurrentQty - params.Qty
	newTotalValue := balance.CurrentTotalValue - totalCost
	if newQty <= 0 {
		newQty = 0
		newTotalValue = 0
	}

	balance.CurrentQty = newQty
	balance.CurrentAvgCost = standardCost
	balance.CurrentTotalValue = newTotalValue
	if err := updateBalance(ctx, tx, balance); err != nil {
		return nil, err
	}

	ct := &inv.InventoryCostTransaction{
		HoldingCode: params.HoldingCode, ItemCode: params.ItemCode, Barcode: params.Barcode,
		WhCode: params.WhCode, LocationCode: params.LocationCode,
		TransactionType: inv.TxTypeSalesIssue, TransFlag: params.TransFlag,
		RefDocType: params.RefDocType, RefDocNo: params.RefDocNo,
		Qty: -params.Qty, UnitCost: standardCost, TotalCost: -totalCost,
		BalanceQty: newQty, BalanceAvgCost: standardCost, BalanceTotalValue: newTotalValue,
		CostingMethodUsed: inv.CostingMethodStandard,
		TransactionDate:   params.TransactionDate, CreatedBy: params.CreatedBy,
	}
	if err := insertCostTransaction(ctx, tx, ct); err != nil {
		return nil, err
	}

	return &inv.CostTransactionResult{
		Transaction: ct, BalanceQty: newQty, BalanceAvgCost: standardCost, TotalValue: newTotalValue,
	}, nil
}

func (e *StandardEngine) ProcessSalesReturn(ctx context.Context, tx *sql.Tx, params inv.SalesReturnParams) (*inv.CostTransactionResult, error) {
	balance, err := getOrCreateBalance(ctx, tx, params.HoldingCode, params.ItemCode, params.Barcode, params.WhCode, params.LocationCode)
	if err != nil {
		return nil, err
	}

	standardCost, err := getStandardCost(ctx, tx, params.HoldingCode, params.ItemCode)
	if err != nil {
		return nil, err
	}

	totalCost := params.Qty * standardCost
	newQty := balance.CurrentQty + params.Qty
	newTotalValue := balance.CurrentTotalValue + totalCost

	balance.CurrentQty = newQty
	balance.CurrentAvgCost = standardCost
	balance.CurrentTotalValue = newTotalValue
	if err := updateBalance(ctx, tx, balance); err != nil {
		return nil, err
	}

	// Variance ถ้า original cost ≠ standard
	var variance *inv.InventoryVariance
	if params.OriginalCost > 0 && params.OriginalCost != standardCost {
		varianceAmount := (params.OriginalCost - standardCost) * params.Qty
		variance = &inv.InventoryVariance{
			HoldingCode: params.HoldingCode, ItemCode: params.ItemCode, WhCode: params.WhCode,
			VarianceType: inv.VariancePurchasePrice,
			RefDocType:   params.RefDocType, RefDocNo: params.RefDocNo,
			StandardCost: standardCost, ActualCost: params.OriginalCost,
			Qty: params.Qty, VarianceAmount: varianceAmount,
			TransactionDate: params.TransactionDate,
		}
		if err := insertVariance(ctx, tx, variance); err != nil {
			return nil, err
		}
	}

	ct := &inv.InventoryCostTransaction{
		HoldingCode: params.HoldingCode, ItemCode: params.ItemCode, Barcode: params.Barcode,
		WhCode: params.WhCode, LocationCode: params.LocationCode,
		TransactionType: inv.TxTypeSalesReturn, TransFlag: params.TransFlag,
		RefDocType: params.RefDocType, RefDocNo: params.RefDocNo,
		Qty: params.Qty, UnitCost: standardCost, TotalCost: totalCost,
		BalanceQty: newQty, BalanceAvgCost: standardCost, BalanceTotalValue: newTotalValue,
		CostingMethodUsed: inv.CostingMethodStandard,
		TransactionDate:   params.TransactionDate, CreatedBy: params.CreatedBy,
	}
	if err := insertCostTransaction(ctx, tx, ct); err != nil {
		return nil, err
	}

	return &inv.CostTransactionResult{
		Transaction: ct, Variance: variance,
		BalanceQty: newQty, BalanceAvgCost: standardCost, TotalValue: newTotalValue,
	}, nil
}

func (e *StandardEngine) ProcessPurchaseReturn(ctx context.Context, tx *sql.Tx, params inv.PurchaseReturnParams) (*inv.CostTransactionResult, error) {
	balance, err := getOrCreateBalance(ctx, tx, params.HoldingCode, params.ItemCode, params.Barcode, params.WhCode, params.LocationCode)
	if err != nil {
		return nil, err
	}

	standardCost, err := getStandardCost(ctx, tx, params.HoldingCode, params.ItemCode)
	if err != nil {
		return nil, err
	}

	totalCost := params.Qty * standardCost
	newQty := balance.CurrentQty - params.Qty
	newTotalValue := balance.CurrentTotalValue - totalCost
	if newQty <= 0 {
		newQty = 0
		newTotalValue = 0
	}

	balance.CurrentQty = newQty
	balance.CurrentAvgCost = standardCost
	balance.CurrentTotalValue = newTotalValue
	if err := updateBalance(ctx, tx, balance); err != nil {
		return nil, err
	}

	ct := &inv.InventoryCostTransaction{
		HoldingCode: params.HoldingCode, ItemCode: params.ItemCode, Barcode: params.Barcode,
		WhCode: params.WhCode, LocationCode: params.LocationCode,
		TransactionType: inv.TxTypePurchaseReturn, TransFlag: params.TransFlag,
		RefDocType: params.RefDocType, RefDocNo: params.RefDocNo,
		Qty: -params.Qty, UnitCost: standardCost, TotalCost: -totalCost,
		BalanceQty: newQty, BalanceAvgCost: standardCost, BalanceTotalValue: newTotalValue,
		CostingMethodUsed: inv.CostingMethodStandard,
		TransactionDate:   params.TransactionDate, CreatedBy: params.CreatedBy,
	}
	if err := insertCostTransaction(ctx, tx, ct); err != nil {
		return nil, err
	}

	return &inv.CostTransactionResult{
		Transaction: ct, BalanceQty: newQty, BalanceAvgCost: standardCost, TotalValue: newTotalValue,
	}, nil
}

func (e *StandardEngine) ProcessAdjustment(ctx context.Context, tx *sql.Tx, params inv.AdjustmentParams) (*inv.CostTransactionResult, error) {
	balance, err := getOrCreateBalance(ctx, tx, params.HoldingCode, params.ItemCode, params.Barcode, params.WhCode, params.LocationCode)
	if err != nil {
		return nil, err
	}

	standardCost, err := getStandardCost(ctx, tx, params.HoldingCode, params.ItemCode)
	if err != nil {
		return nil, err
	}

	var txType string
	var qty, totalCost, newQty, newTotalValue float64

	if params.IsIncrease {
		txType = inv.TxTypeAdjustmentIn
		totalCost = params.Qty * standardCost
		qty = params.Qty
		newQty = balance.CurrentQty + params.Qty
		newTotalValue = balance.CurrentTotalValue + totalCost
	} else {
		txType = inv.TxTypeAdjustmentOut
		totalCost = params.Qty * standardCost
		qty = -params.Qty
		newQty = balance.CurrentQty - params.Qty
		newTotalValue = balance.CurrentTotalValue - totalCost
		if newQty <= 0 {
			newQty = 0
			newTotalValue = 0
		}
	}

	balance.CurrentQty = newQty
	balance.CurrentAvgCost = standardCost
	balance.CurrentTotalValue = newTotalValue
	if err := updateBalance(ctx, tx, balance); err != nil {
		return nil, err
	}

	ct := &inv.InventoryCostTransaction{
		HoldingCode: params.HoldingCode, ItemCode: params.ItemCode, Barcode: params.Barcode,
		WhCode: params.WhCode, LocationCode: params.LocationCode,
		TransactionType: txType, TransFlag: params.TransFlag,
		RefDocType: params.RefDocType, RefDocNo: params.RefDocNo,
		Qty: qty, UnitCost: standardCost, TotalCost: totalCost * float64(boolToSign(params.IsIncrease)),
		BalanceQty: newQty, BalanceAvgCost: standardCost, BalanceTotalValue: newTotalValue,
		CostingMethodUsed: inv.CostingMethodStandard,
		TransactionDate:   params.TransactionDate, CreatedBy: params.CreatedBy,
	}
	if err := insertCostTransaction(ctx, tx, ct); err != nil {
		return nil, err
	}

	return &inv.CostTransactionResult{
		Transaction: ct, BalanceQty: newQty, BalanceAvgCost: standardCost, TotalValue: newTotalValue,
	}, nil
}

func (e *StandardEngine) GetCurrentValuation(ctx context.Context, tx *sql.Tx, holdingCode, itemCode, whCode string) (*inv.StockValuation, error) {
	fifo := &FIFOEngine{}
	val, err := fifo.GetCurrentValuation(ctx, tx, holdingCode, itemCode, whCode)
	if err != nil {
		return nil, err
	}
	val.CostingMethod = inv.CostingMethodStandard
	return val, nil
}

// === Helper Functions ===

// getStandardCost — ดึง standard cost จาก product config
func getStandardCost(ctx context.Context, tx *sql.Tx, holdingCode, itemCode string) (float64, error) {
	var cost float64
	err := tx.QueryRowContext(ctx,
		`SELECT COALESCE(standardcost, 0) FROM product_costing_config WHERE holdingcode = $1 AND itemcode = $2`,
		holdingCode, itemCode,
	).Scan(&cost)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("ไม่พบ standard cost สำหรับสินค้า %s — ต้องตั้งค่า standard cost ก่อน", itemCode)
	}
	if err != nil {
		return 0, fmt.Errorf("ดึง standard cost ไม่สำเร็จ: %w", err)
	}
	if cost <= 0 {
		return 0, fmt.Errorf("standard cost ของสินค้า %s ต้อง > 0", itemCode)
	}
	return cost, nil
}

// insertVariance — บันทึก variance
func insertVariance(ctx context.Context, tx *sql.Tx, v *inv.InventoryVariance) error {
	return tx.QueryRowContext(ctx,
		`INSERT INTO inventory_variances
		 (holdingcode, itemcode, whcode, variancetype, refdoctype, refdocno,
		  standardcost, actualcost, qty, varianceamount,
		  transactiondate, accountingperiod, createdat)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,NOW())
		 RETURNING id`,
		v.HoldingCode, v.ItemCode, v.WhCode, v.VarianceType, v.RefDocType, v.RefDocNo,
		v.StandardCost, v.ActualCost, v.Qty, v.VarianceAmount,
		v.TransactionDate, v.AccountingPeriod,
	).Scan(&v.ID)
}
