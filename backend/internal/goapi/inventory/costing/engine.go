package costing

import (
	"context"
	"database/sql"
	"fmt"

	inv "smlcloudplatform/internal/goapi/inventory/models"
)

// CostingEngine — interface หลักสำหรับทุกวิธีคำนวณต้นทุน
// แต่ละ method implement interface นี้ (Moving Average, FIFO, LIFO, ฯลฯ)
type CostingEngine interface {
	// ProcessReceipt — รับสินค้าเข้า (ซื้อ, รับโอน, ปรับเพิ่ม)
	ProcessReceipt(ctx context.Context, tx *sql.Tx, params inv.ReceiptParams) (*inv.CostTransactionResult, error)

	// ProcessIssue — ตัดสินค้าออก (ขาย, เบิก, โอนออก)
	ProcessIssue(ctx context.Context, tx *sql.Tx, params inv.IssueParams) (*inv.CostTransactionResult, error)

	// ProcessSalesReturn — รับคืนจากลูกค้า
	ProcessSalesReturn(ctx context.Context, tx *sql.Tx, params inv.SalesReturnParams) (*inv.CostTransactionResult, error)

	// ProcessPurchaseReturn — ส่งคืนสินค้าให้ supplier
	ProcessPurchaseReturn(ctx context.Context, tx *sql.Tx, params inv.PurchaseReturnParams) (*inv.CostTransactionResult, error)

	// ProcessAdjustment — ปรับปรุง stock (เพิ่ม/ลด)
	ProcessAdjustment(ctx context.Context, tx *sql.Tx, params inv.AdjustmentParams) (*inv.CostTransactionResult, error)

	// GetCurrentValuation — ดูมูลค่าสินค้าปัจจุบัน
	GetCurrentValuation(ctx context.Context, tx *sql.Tx, shopID, itemCode, whCode string) (*inv.StockValuation, error)

	// Method — ชื่อวิธีคำนวณ
	Method() string
}

// NewCostingEngine — สร้าง engine ตาม costing method
func NewCostingEngine(method string) (CostingEngine, error) {
	switch method {
	case inv.CostingMethodMovingAverage:
		return &MovingAverageEngine{}, nil
	case inv.CostingMethodFIFO:
		return &FIFOEngine{}, nil
	case inv.CostingMethodLIFO:
		return &LIFOEngine{}, nil
	case inv.CostingMethodFEFO:
		return &FEFOEngine{}, nil
	case inv.CostingMethodStandard:
		return &StandardEngine{}, nil
	default:
		return nil, fmt.Errorf("ไม่รู้จัก costing method: %s", method)
	}
}

// === Helper Functions สำหรับทุก Engine ===

// getOrCreateBalance — ดึง stock balance หรือสร้างใหม่ถ้ายังไม่มี
func getOrCreateBalance(ctx context.Context, tx *sql.Tx, shopID, itemCode, barcode, whCode, locationCode string) (*inv.InventoryStockBalance, error) {
	var balance inv.InventoryStockBalance
	err := tx.QueryRowContext(ctx,
		`SELECT id, shopid, itemcode, barcode, whcode, locationcode,
		        currentqty, currentavgcost, currenttotalvalue,
		        lastpurchasecost, lastpurchasedate, updatedat
		 FROM inventory_stock_balances
		 WHERE shopid = $1 AND itemcode = $2 AND whcode = $3 AND locationcode = $4
		 FOR UPDATE`,
		shopID, itemCode, whCode, locationCode,
	).Scan(
		&balance.ID, &balance.ShopID, &balance.ItemCode, &balance.Barcode,
		&balance.WhCode, &balance.LocationCode,
		&balance.CurrentQty, &balance.CurrentAvgCost, &balance.CurrentTotalValue,
		&balance.LastPurchaseCost, &balance.LastPurchaseDate, &balance.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		// สร้าง balance ใหม่
		balance = inv.InventoryStockBalance{
			ShopID:       shopID,
			ItemCode:     itemCode,
			Barcode:      barcode,
			WhCode:       whCode,
			LocationCode: locationCode,
		}
		err = tx.QueryRowContext(ctx,
			`INSERT INTO inventory_stock_balances
			 (shopid, itemcode, barcode, whcode, locationcode, currentqty, currentavgcost, currenttotalvalue, updatedat)
			 VALUES ($1, $2, $3, $4, $5, 0, 0, 0, NOW())
			 RETURNING id`,
			shopID, itemCode, barcode, whCode, locationCode,
		).Scan(&balance.ID)
		if err != nil {
			return nil, fmt.Errorf("สร้าง stock balance ไม่สำเร็จ: %w", err)
		}
		return &balance, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ดึง stock balance ไม่สำเร็จ: %w", err)
	}
	return &balance, nil
}

// updateBalance — อัพเดท stock balance
func updateBalance(ctx context.Context, tx *sql.Tx, balance *inv.InventoryStockBalance) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE inventory_stock_balances
		 SET currentqty = $1, currentavgcost = $2, currenttotalvalue = $3,
		     lastpurchasecost = $4, lastpurchasedate = $5, updatedat = NOW()
		 WHERE id = $6`,
		balance.CurrentQty, balance.CurrentAvgCost, balance.CurrentTotalValue,
		balance.LastPurchaseCost, balance.LastPurchaseDate, balance.ID,
	)
	if err != nil {
		return fmt.Errorf("อัพเดท stock balance ไม่สำเร็จ: %w", err)
	}
	return nil
}

// insertCostTransaction — บันทึก cost transaction
func insertCostTransaction(ctx context.Context, tx *sql.Tx, ct *inv.InventoryCostTransaction) error {
	return tx.QueryRowContext(ctx,
		`INSERT INTO inventory_cost_transactions
		 (shopid, itemcode, barcode, whcode, locationcode, transactiontype, transflag,
		  refdoctype, refdocno, qty, unitcost, totalcost, landedcost, lotnumber, expirydate,
		  costlayerid, balanceqty, balanceavgcost, balancetotalvalue, costingmethodused,
		  transactiondate, accountingperiod, createdby, createdat, isreversed)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,NOW(),false)
		 RETURNING id`,
		ct.ShopID, ct.ItemCode, ct.Barcode, ct.WhCode, ct.LocationCode,
		ct.TransactionType, ct.TransFlag, ct.RefDocType, ct.RefDocNo,
		ct.Qty, ct.UnitCost, ct.TotalCost, ct.LandedCost,
		ct.LotNumber, ct.ExpiryDate, ct.CostLayerID,
		ct.BalanceQty, ct.BalanceAvgCost, ct.BalanceTotalValue,
		ct.CostingMethodUsed, ct.TransactionDate, ct.AccountingPeriod, ct.CreatedBy,
	).Scan(&ct.ID)
}

// insertCostLayer — บันทึก cost layer ใหม่
func insertCostLayer(ctx context.Context, tx *sql.Tx, layer *inv.InventoryCostLayer) error {
	return tx.QueryRowContext(ctx,
		`INSERT INTO inventory_cost_layers
		 (shopid, itemcode, barcode, whcode, locationcode, layertype, refdoctype, refdocno,
		  originalqty, remainingqty, unitcost, landedcostperunit, totalunitcost,
		  lotnumber, supplierlotnumber, manufacturingdate, expirydate, qualitystatus,
		  receiveddate, createdat, updatedat)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,NOW(),NOW())
		 RETURNING id`,
		layer.ShopID, layer.ItemCode, layer.Barcode, layer.WhCode, layer.LocationCode,
		layer.LayerType, layer.RefDocType, layer.RefDocNo,
		layer.OriginalQty, layer.RemainingQty, layer.UnitCost, layer.LandedCostPerUnit, layer.TotalUnitCost,
		layer.LotNumber, layer.SupplierLotNumber, layer.ManufacturingDate, layer.ExpiryDate, layer.QualityStatus,
		layer.ReceivedDate,
	).Scan(&layer.ID)
}

// averageCostCalc — คำนวณต้นทุนเฉลี่ย (ป้องกัน divide by zero)
func averageCostCalc(totalValue, totalQty float64) float64 {
	if totalQty <= 0 {
		return 0
	}
	return totalValue / totalQty
}
