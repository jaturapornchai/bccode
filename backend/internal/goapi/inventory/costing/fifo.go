package costing

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	inv "smlcloudplatform/internal/goapi/inventory/models"
)

// FIFOEngine — First In, First Out
// สินค้าที่ซื้อก่อนตัดออกก่อน — เก็บ cost layer แยกตามแต่ละครั้งที่รับเข้า
type FIFOEngine struct{}

func (e *FIFOEngine) Method() string { return inv.CostingMethodFIFO }

// ProcessReceipt — รับสินค้าเข้า → สร้าง cost layer ใหม่
func (e *FIFOEngine) ProcessReceipt(ctx context.Context, tx *sql.Tx, params inv.ReceiptParams) (*inv.CostTransactionResult, error) {
	balance, err := getOrCreateBalance(ctx, tx, params.HoldingCode, params.ItemCode, params.Barcode, params.WhCode, params.LocationCode)
	if err != nil {
		return nil, err
	}

	totalCost := params.Qty * params.UnitCost

	// สร้าง cost layer ใหม่
	layer := &inv.InventoryCostLayer{
		HoldingCode:   params.HoldingCode,
		ItemCode:      params.ItemCode,
		Barcode:       params.Barcode,
		WhCode:        params.WhCode,
		LocationCode:  params.LocationCode,
		LayerType:     inv.LayerTypePurchase,
		RefDocType:    params.RefDocType,
		RefDocNo:      params.RefDocNo,
		OriginalQty:   params.Qty,
		RemainingQty:  params.Qty,
		UnitCost:      params.UnitCost,
		TotalUnitCost: params.UnitCost,
		LotNumber:     params.LotNumber,
		ExpiryDate:    params.ExpiryDate,
		QualityStatus: inv.QualityStatusApproved,
		ReceivedDate:  params.ReceivedDate,
	}
	if err := insertCostLayer(ctx, tx, layer); err != nil {
		return nil, err
	}

	// อัพเดท balance
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
		CostLayerID:       &layer.ID,
		BalanceQty:        newQty,
		BalanceAvgCost:    newAvgCost,
		BalanceTotalValue: newTotalValue,
		CostingMethodUsed: inv.CostingMethodFIFO,
		TransactionDate:   params.ReceivedDate,
		CreatedBy:         params.CreatedBy,
	}
	if err := insertCostTransaction(ctx, tx, ct); err != nil {
		return nil, err
	}

	return &inv.CostTransactionResult{
		Transaction:    ct,
		CostLayer:      layer,
		BalanceQty:     newQty,
		BalanceAvgCost: newAvgCost,
		TotalValue:     newTotalValue,
	}, nil
}

// ProcessIssue — ตัดสินค้าออก → ตัด layer เก่าสุดก่อน (FIFO)
func (e *FIFOEngine) ProcessIssue(ctx context.Context, tx *sql.Tx, params inv.IssueParams) (*inv.CostTransactionResult, error) {
	balance, err := getOrCreateBalance(ctx, tx, params.HoldingCode, params.ItemCode, params.Barcode, params.WhCode, params.LocationCode)
	if err != nil {
		return nil, err
	}

	if balance.CurrentQty < params.Qty {
		return nil, fmt.Errorf("สต็อกไม่พอ: คงเหลือ %.4f ต้องการ %.4f", balance.CurrentQty, params.Qty)
	}

	// ดึง cost layers เรียงจากเก่าสุด (FIFO)
	totalCost, err := e.consumeLayers(ctx, tx, params.HoldingCode, params.ItemCode, params.WhCode, params.LocationCode, params.Qty)
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
		CostingMethodUsed: inv.CostingMethodFIFO,
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

// consumeLayers — ตัด cost layers ตามลำดับ FIFO (เก่าสุดก่อน)
// return: totalCost ที่ตัดออก
func (e *FIFOEngine) consumeLayers(ctx context.Context, tx *sql.Tx, holdingCode, itemCode, whCode, locationCode string, qtyNeeded float64) (float64, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT id, remainingqty, totalunitcost
		 FROM inventorycostlayers
		 WHERE holdingcode = $1 AND itemcode = $2 AND whcode = $3 AND locationcode = $4 AND remainingqty > 0
		 ORDER BY receiveddate ASC, id ASC
		 FOR UPDATE`,
		holdingCode, itemCode, whCode, locationCode,
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
		newLayerQty := layerQty - consume
		remaining -= consume

		// อัพเดท layer
		_, err := tx.ExecContext(ctx,
			`UPDATE inventorycostlayers SET remainingqty = $1, updatedat = NOW() WHERE id = $2`,
			newLayerQty, layerID,
		)
		if err != nil {
			return 0, fmt.Errorf("อัพเดท cost layer ไม่สำเร็จ: %w", err)
		}
	}

	if remaining > 0 {
		return 0, fmt.Errorf("cost layers ไม่พอ: ยังขาดอีก %.4f หน่วย", remaining)
	}

	return totalCost, nil
}

// ProcessSalesReturn — รับคืนจากลูกค้า → สร้าง layer ใหม่ด้วย original cost
func (e *FIFOEngine) ProcessSalesReturn(ctx context.Context, tx *sql.Tx, params inv.SalesReturnParams) (*inv.CostTransactionResult, error) {
	balance, err := getOrCreateBalance(ctx, tx, params.HoldingCode, params.ItemCode, params.Barcode, params.WhCode, params.LocationCode)
	if err != nil {
		return nil, err
	}

	unitCost := params.OriginalCost
	if unitCost <= 0 {
		unitCost = balance.CurrentAvgCost
	}

	// สร้าง cost layer ใหม่สำหรับของที่รับคืน
	layer := &inv.InventoryCostLayer{
		HoldingCode:   params.HoldingCode,
		ItemCode:      params.ItemCode,
		Barcode:       params.Barcode,
		WhCode:        params.WhCode,
		LocationCode:  params.LocationCode,
		LayerType:     inv.LayerTypeSalesReturn,
		RefDocType:    params.RefDocType,
		RefDocNo:      params.RefDocNo,
		OriginalQty:   params.Qty,
		RemainingQty:  params.Qty,
		UnitCost:      unitCost,
		TotalUnitCost: unitCost,
		QualityStatus: inv.QualityStatusApproved,
		ReceivedDate:  params.TransactionDate,
	}
	if err := insertCostLayer(ctx, tx, layer); err != nil {
		return nil, err
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
		CostLayerID:       &layer.ID,
		BalanceQty:        newQty,
		BalanceAvgCost:    newAvgCost,
		BalanceTotalValue: newTotalValue,
		CostingMethodUsed: inv.CostingMethodFIFO,
		TransactionDate:   params.TransactionDate,
		CreatedBy:         params.CreatedBy,
	}
	if err := insertCostTransaction(ctx, tx, ct); err != nil {
		return nil, err
	}

	return &inv.CostTransactionResult{
		Transaction:    ct,
		CostLayer:      layer,
		BalanceQty:     newQty,
		BalanceAvgCost: newAvgCost,
		TotalValue:     newTotalValue,
	}, nil
}

// ProcessPurchaseReturn — ส่งคืน supplier → ตัด layer ที่เกี่ยวข้อง
func (e *FIFOEngine) ProcessPurchaseReturn(ctx context.Context, tx *sql.Tx, params inv.PurchaseReturnParams) (*inv.CostTransactionResult, error) {
	balance, err := getOrCreateBalance(ctx, tx, params.HoldingCode, params.ItemCode, params.Barcode, params.WhCode, params.LocationCode)
	if err != nil {
		return nil, err
	}

	// ถ้าระบุ layerid → ตัด layer นั้นโดยเฉพาะ
	var unitCost, totalCost float64
	if params.CostLayerID != nil {
		unitCost, err = e.consumeSpecificLayer(ctx, tx, *params.CostLayerID, params.Qty)
		if err != nil {
			return nil, err
		}
		totalCost = params.Qty * unitCost
	} else {
		// ไม่ระบุ → ตัดตาม FIFO
		totalCost, err = e.consumeLayers(ctx, tx, params.HoldingCode, params.ItemCode, params.WhCode, params.LocationCode, params.Qty)
		if err != nil {
			return nil, err
		}
		unitCost = totalCost / params.Qty
	}

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
		CostLayerID:       params.CostLayerID,
		BalanceQty:        newQty,
		BalanceAvgCost:    newAvgCost,
		BalanceTotalValue: newTotalValue,
		CostingMethodUsed: inv.CostingMethodFIFO,
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

// consumeSpecificLayer — ตัด layer เฉพาะที่ระบุ
func (e *FIFOEngine) consumeSpecificLayer(ctx context.Context, tx *sql.Tx, layerID int64, qty float64) (float64, error) {
	var remainingQty, unitCost float64
	err := tx.QueryRowContext(ctx,
		`SELECT remainingqty, totalunitcost FROM inventorycostlayers WHERE id = $1 FOR UPDATE`,
		layerID,
	).Scan(&remainingQty, &unitCost)
	if err != nil {
		return 0, fmt.Errorf("ดึง cost layer %d ไม่สำเร็จ: %w", layerID, err)
	}
	if remainingQty < qty {
		return 0, fmt.Errorf("layer %d เหลือ %.4f ไม่พอตัด %.4f", layerID, remainingQty, qty)
	}

	newQty := remainingQty - qty
	_, err = tx.ExecContext(ctx,
		`UPDATE inventorycostlayers SET remainingqty = $1, updatedat = NOW() WHERE id = $2`,
		newQty, layerID,
	)
	if err != nil {
		return 0, err
	}
	return unitCost, nil
}

// ProcessAdjustment — ปรับปรุง stock
func (e *FIFOEngine) ProcessAdjustment(ctx context.Context, tx *sql.Tx, params inv.AdjustmentParams) (*inv.CostTransactionResult, error) {
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

		// สร้าง layer ใหม่
		layer := &inv.InventoryCostLayer{
			HoldingCode:   params.HoldingCode,
			ItemCode:      params.ItemCode,
			Barcode:       params.Barcode,
			WhCode:        params.WhCode,
			LocationCode:  params.LocationCode,
			LayerType:     inv.LayerTypeAdjustment,
			RefDocType:    params.RefDocType,
			RefDocNo:      params.RefDocNo,
			OriginalQty:   params.Qty,
			RemainingQty:  params.Qty,
			UnitCost:      unitCost,
			TotalUnitCost: unitCost,
			LotNumber:     params.LotNumber,
			ExpiryDate:    params.ExpiryDate,
			QualityStatus: inv.QualityStatusApproved,
			ReceivedDate:  params.TransactionDate,
		}
		if err := insertCostLayer(ctx, tx, layer); err != nil {
			return nil, err
		}

		totalCost = params.Qty * unitCost
		qty = params.Qty
		newQty = balance.CurrentQty + params.Qty
		newTotalValue = balance.CurrentTotalValue + totalCost
	} else {
		txType = inv.TxTypeAdjustmentOut
		totalCost, err = e.consumeLayers(ctx, tx, params.HoldingCode, params.ItemCode, params.WhCode, params.LocationCode, params.Qty)
		if err != nil {
			return nil, err
		}
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
		CostingMethodUsed: inv.CostingMethodFIFO,
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
func (e *FIFOEngine) GetCurrentValuation(ctx context.Context, tx *sql.Tx, holdingCode, itemCode, whCode string) (*inv.StockValuation, error) {
	var val inv.StockValuation
	var err error
	if whCode == "" {
		err = tx.QueryRowContext(ctx,
			`SELECT itemcode, '' AS whcode,
			        COALESCE(SUM(currentqty), 0),
			        CASE WHEN COALESCE(SUM(currentqty), 0) = 0 THEN 0 ELSE COALESCE(SUM(currenttotalvalue), 0) / SUM(currentqty) END,
			        COALESCE(SUM(currenttotalvalue), 0)
			 FROM inventorystockbalances
			 WHERE holdingcode = $1 AND itemcode = $2
			 GROUP BY itemcode`,
			holdingCode, itemCode,
		).Scan(&val.ItemCode, &val.WhCode, &val.CurrentQty, &val.AverageCost, &val.TotalValue)
	} else {
		err = tx.QueryRowContext(ctx,
			`SELECT itemcode, whcode,
			        COALESCE(SUM(currentqty), 0),
			        CASE WHEN COALESCE(SUM(currentqty), 0) = 0 THEN 0 ELSE COALESCE(SUM(currenttotalvalue), 0) / SUM(currentqty) END,
			        COALESCE(SUM(currenttotalvalue), 0)
			 FROM inventorystockbalances
			 WHERE holdingcode = $1 AND itemcode = $2 AND whcode = $3
			 GROUP BY itemcode, whcode`,
			holdingCode, itemCode, whCode,
		).Scan(&val.ItemCode, &val.WhCode, &val.CurrentQty, &val.AverageCost, &val.TotalValue)
	}
	if err == sql.ErrNoRows {
		return &inv.StockValuation{ItemCode: itemCode, WhCode: whCode, CostingMethod: inv.CostingMethodFIFO}, nil
	}
	if err != nil {
		return nil, err
	}
	val.CostingMethod = inv.CostingMethodFIFO
	return &val, nil
}
