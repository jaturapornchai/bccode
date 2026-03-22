package inventory

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"smlcloudplatform/internal/goapi/inventory/costing"
	m "smlcloudplatform/internal/goapi/inventory/models"
	"smlcloudplatform/internal/goapi/logger"
)

// InventoryCostingService — service หลักสำหรับ orchestrate ทุก costing operation
// จัดการ DB transaction, เลือก engine ที่ถูกต้อง, ประมวลผล
type InventoryCostingService struct {
	db *sql.DB
}

// NewInventoryCostingService — สร้าง service
func NewInventoryCostingService(db *sql.DB) *InventoryCostingService {
	return &InventoryCostingService{db: db}
}

// === Public Methods ===

// ProcessReceipt — รับสินค้าเข้า (ซื้อ, รับโอน, ปรับเพิ่ม)
func (s *InventoryCostingService) ProcessReceipt(ctx context.Context, params m.ReceiptParams) (*m.CostTransactionResult, error) {
	return s.withTransaction(ctx, func(tx *sql.Tx) (*m.CostTransactionResult, error) {
		engine, err := s.getEngineForProduct(ctx, tx, params.ShopID, params.ItemCode)
		if err != nil {
			return nil, err
		}
		logger.Info(fmt.Sprintf("รับสินค้า %s ด้วย %s: qty=%.4f cost=%.4f", params.ItemCode, engine.Method(), params.Qty, params.UnitCost))
		return engine.ProcessReceipt(ctx, tx, params)
	})
}

// ProcessIssue — ตัดสินค้าออก (ขาย, เบิก)
func (s *InventoryCostingService) ProcessIssue(ctx context.Context, params m.IssueParams) (*m.CostTransactionResult, error) {
	return s.withTransaction(ctx, func(tx *sql.Tx) (*m.CostTransactionResult, error) {
		engine, err := s.getEngineForProduct(ctx, tx, params.ShopID, params.ItemCode)
		if err != nil {
			return nil, err
		}
		logger.Info(fmt.Sprintf("ตัดสินค้า %s ด้วย %s: qty=%.4f", params.ItemCode, engine.Method(), params.Qty))
		return engine.ProcessIssue(ctx, tx, params)
	})
}

// ProcessSalesReturn — รับคืนจากลูกค้า
func (s *InventoryCostingService) ProcessSalesReturn(ctx context.Context, params m.SalesReturnParams) (*m.CostTransactionResult, error) {
	return s.withTransaction(ctx, func(tx *sql.Tx) (*m.CostTransactionResult, error) {
		engine, err := s.getEngineForProduct(ctx, tx, params.ShopID, params.ItemCode)
		if err != nil {
			return nil, err
		}
		return engine.ProcessSalesReturn(ctx, tx, params)
	})
}

// ProcessPurchaseReturn — ส่งคืนสินค้าให้ supplier
func (s *InventoryCostingService) ProcessPurchaseReturn(ctx context.Context, params m.PurchaseReturnParams) (*m.CostTransactionResult, error) {
	return s.withTransaction(ctx, func(tx *sql.Tx) (*m.CostTransactionResult, error) {
		engine, err := s.getEngineForProduct(ctx, tx, params.ShopID, params.ItemCode)
		if err != nil {
			return nil, err
		}
		return engine.ProcessPurchaseReturn(ctx, tx, params)
	})
}

// ProcessTransfer — โอนย้ายคลัง (ตัดออกจากต้นทาง + รับเข้าปลายทาง)
func (s *InventoryCostingService) ProcessTransfer(ctx context.Context, params m.TransferParams) (*m.CostTransactionResult, error) {
	return s.withTransaction(ctx, func(tx *sql.Tx) (*m.CostTransactionResult, error) {
		engine, err := s.getEngineForProduct(ctx, tx, params.ShopID, params.ItemCode)
		if err != nil {
			return nil, err
		}

		// 1. ตัดออกจากคลังต้นทาง
		issueParams := m.IssueParams{
			ShopID: params.ShopID, ItemCode: params.ItemCode, Barcode: params.Barcode,
			WhCode: params.FromWhCode, LocationCode: params.FromLocationCode,
			Qty: params.Qty, RefDocType: params.RefDocType, RefDocNo: params.RefDocNo,
			TransFlag: params.TransFlag, LotNumber: params.LotNumber,
			TransactionDate: params.TransactionDate, CreatedBy: params.CreatedBy,
		}
		issueResult, err := engine.ProcessIssue(ctx, tx, issueParams)
		if err != nil {
			return nil, fmt.Errorf("ตัดออกจากคลังต้นทาง %s ไม่สำเร็จ: %w", params.FromWhCode, err)
		}

		// ใช้ unit cost จากการตัดออก
		unitCost := issueResult.Transaction.UnitCost
		if unitCost < 0 {
			unitCost = -unitCost
		}

		// 2. รับเข้าคลังปลายทาง (ใช้ต้นทุนจากต้นทาง)
		receiptParams := m.ReceiptParams{
			ShopID: params.ShopID, ItemCode: params.ItemCode, Barcode: params.Barcode,
			WhCode: params.ToWhCode, LocationCode: params.ToLocationCode,
			Qty: params.Qty, UnitCost: unitCost,
			RefDocType: params.RefDocType, RefDocNo: params.RefDocNo,
			TransFlag: params.TransFlag, LotNumber: params.LotNumber,
			ReceivedDate: params.TransactionDate, CreatedBy: params.CreatedBy,
		}
		receiptResult, err := engine.ProcessReceipt(ctx, tx, receiptParams)
		if err != nil {
			return nil, fmt.Errorf("รับเข้าคลังปลายทาง %s ไม่สำเร็จ: %w", params.ToWhCode, err)
		}

		// อัพเดท transaction type เป็น transfer
		issueResult.Transaction.TransactionType = m.TxTypeTransferOut
		receiptResult.Transaction.TransactionType = m.TxTypeTransferIn

		return receiptResult, nil
	})
}

// ProcessAdjustment — ปรับปรุง stock
func (s *InventoryCostingService) ProcessAdjustment(ctx context.Context, params m.AdjustmentParams) (*m.CostTransactionResult, error) {
	return s.withTransaction(ctx, func(tx *sql.Tx) (*m.CostTransactionResult, error) {
		engine, err := s.getEngineForProduct(ctx, tx, params.ShopID, params.ItemCode)
		if err != nil {
			return nil, err
		}
		return engine.ProcessAdjustment(ctx, tx, params)
	})
}

// GetValuation — ดูมูลค่าสินค้าปัจจุบัน
func (s *InventoryCostingService) GetValuation(ctx context.Context, shopID, itemCode, whCode string) (*m.StockValuation, error) {
	var result *m.StockValuation
	_, err := s.withTransaction(ctx, func(tx *sql.Tx) (*m.CostTransactionResult, error) {
		engine, err := s.getEngineForProduct(ctx, tx, shopID, itemCode)
		if err != nil {
			return nil, err
		}
		result, err = engine.GetCurrentValuation(ctx, tx, shopID, itemCode, whCode)
		return nil, err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetCostingConfig — ดึง costing config ของสินค้า
func (s *InventoryCostingService) GetCostingConfig(ctx context.Context, shopID, itemCode string) (*m.ProductCostingConfig, error) {
	var config m.ProductCostingConfig
	err := s.db.QueryRowContext(ctx,
		`SELECT itemcode, costingmethod, lottrackingenabled, expirytrackingenabled,
		        expiryalertdays, autoblockexpired, allownegativestock, standardcost
		 FROM product_costing_config WHERE shopid = $1 AND itemcode = $2`,
		shopID, itemCode,
	).Scan(&config.ItemCode, &config.CostingMethod, &config.LotTrackingEnabled,
		&config.ExpiryTrackingEnabled, &config.ExpiryAlertDays,
		&config.AutoBlockExpired, &config.AllowNegativeStock, &config.StandardCost)
	if err == sql.ErrNoRows {
		return &m.ProductCostingConfig{
			ItemCode:      itemCode,
			CostingMethod: m.CostingMethodMovingAverage,
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// UpdateCostingConfig — ตั้งค่า costing method ของสินค้า
func (s *InventoryCostingService) UpdateCostingConfig(ctx context.Context, shopID string, config *m.ProductCostingConfig) error {
	validMethods := map[string]bool{
		m.CostingMethodMovingAverage:  true,
		m.CostingMethodPeriodicAverage: true,
		m.CostingMethodFIFO:           true,
		m.CostingMethodLIFO:           true,
		m.CostingMethodFEFO:           true,
		m.CostingMethodLot:            true,
		m.CostingMethodStandard:       true,
	}
	if !validMethods[config.CostingMethod] {
		return fmt.Errorf("ไม่รู้จัก costing method: %s", config.CostingMethod)
	}
	if config.CostingMethod == m.CostingMethodFEFO {
		if !config.ExpiryTrackingEnabled || !config.LotTrackingEnabled {
			return fmt.Errorf("FEFO ต้องเปิด expiry tracking + lot tracking")
		}
	}
	if config.CostingMethod == m.CostingMethodLot && !config.LotTrackingEnabled {
		return fmt.Errorf("Lot-Based ต้องเปิด lot tracking")
	}
	if config.CostingMethod == m.CostingMethodStandard && config.StandardCost <= 0 {
		return fmt.Errorf("Standard Cost ต้องตั้งค่า standard cost > 0")
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO product_costing_config
		 (shopid, itemcode, costingmethod, lottrackingenabled, expirytrackingenabled,
		  expiryalertdays, autoblockexpired, allownegativestock, standardcost, updatedat)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW())
		 ON CONFLICT (shopid, itemcode) DO UPDATE SET
		  costingmethod = EXCLUDED.costingmethod,
		  lottrackingenabled = EXCLUDED.lottrackingenabled,
		  expirytrackingenabled = EXCLUDED.expirytrackingenabled,
		  expiryalertdays = EXCLUDED.expiryalertdays,
		  autoblockexpired = EXCLUDED.autoblockexpired,
		  allownegativestock = EXCLUDED.allownegativestock,
		  standardcost = EXCLUDED.standardcost,
		  updatedat = NOW()`,
		shopID, config.ItemCode, config.CostingMethod,
		config.LotTrackingEnabled, config.ExpiryTrackingEnabled,
		config.ExpiryAlertDays, config.AutoBlockExpired,
		config.AllowNegativeStock, config.StandardCost,
	)
	return err
}

// GetStockCard — รายงาน Stock Card (ประวัติเคลื่อนไหวสินค้า)
func (s *InventoryCostingService) GetStockCard(ctx context.Context, shopID, itemCode string, fromDate, toDate time.Time) (*m.StockCardReport, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT transactiondate, refdocno, transactiontype, transflag,
		        qty, unitcost, totalcost, balanceqty, balancetotalvalue, balanceavgcost
		 FROM inventory_cost_transactions
		 WHERE shopid = $1 AND itemcode = $2 AND transactiondate >= $3 AND transactiondate <= $4
		 ORDER BY transactiondate ASC, id ASC`,
		shopID, itemCode, fromDate, toDate,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	report := &m.StockCardReport{
		ShopID:   shopID,
		ItemCode: itemCode,
		FromDate: fromDate,
		ToDate:   toDate,
	}

	for rows.Next() {
		var entry m.StockCardEntry
		var qty float64
		err := rows.Scan(&entry.Date, &entry.DocNo, &entry.TransactionType, &entry.TransFlag,
			&qty, &entry.UnitCost, &entry.TotalCost,
			&entry.BalanceQty, &entry.BalanceValue, &entry.AverageCost)
		if err != nil {
			return nil, err
		}
		if qty > 0 {
			entry.QtyIn = qty
		} else {
			entry.QtyOut = -qty
		}
		report.Entries = append(report.Entries, entry)
	}

	return report, nil
}

// GetCostLayers — ดู cost layers ที่ยังเหลือ (สำหรับ FIFO/LIFO/FEFO/Lot)
func (s *InventoryCostingService) GetCostLayers(ctx context.Context, shopID, itemCode, whCode string) ([]m.InventoryCostLayer, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, shopid, itemcode, barcode, whcode, locationcode,
		        layertype, refdoctype, refdocno, originalqty, remainingqty,
		        unitcost, landedcostperunit, totalunitcost,
		        lotnumber, supplierlotnumber, manufacturingdate, expirydate,
		        qualitystatus, receiveddate, createdat
		 FROM inventory_cost_layers
		 WHERE shopid = $1 AND itemcode = $2 AND whcode = $3 AND remainingqty > 0
		 ORDER BY receiveddate ASC, id ASC`,
		shopID, itemCode, whCode,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var layers []m.InventoryCostLayer
	for rows.Next() {
		var l m.InventoryCostLayer
		err := rows.Scan(&l.ID, &l.ShopID, &l.ItemCode, &l.Barcode, &l.WhCode, &l.LocationCode,
			&l.LayerType, &l.RefDocType, &l.RefDocNo, &l.OriginalQty, &l.RemainingQty,
			&l.UnitCost, &l.LandedCostPerUnit, &l.TotalUnitCost,
			&l.LotNumber, &l.SupplierLotNumber, &l.ManufacturingDate, &l.ExpiryDate,
			&l.QualityStatus, &l.ReceivedDate, &l.CreatedAt)
		if err != nil {
			return nil, err
		}
		layers = append(layers, l)
	}
	return layers, nil
}

// GetInventoryValuation — รายงานมูลค่าสินค้าคงเหลือทั้งร้าน
func (s *InventoryCostingService) GetInventoryValuation(ctx context.Context, shopID string) (*m.InventoryValuationReport, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT sb.itemcode, sb.whcode, sb.currentqty, sb.currentavgcost, sb.currenttotalvalue,
		        COALESCE(pc.costingmethod, 'moving_average')
		 FROM inventory_stock_balances sb
		 LEFT JOIN product_costing_config pc ON sb.shopid = pc.shopid AND sb.itemcode = pc.itemcode
		 WHERE sb.shopid = $1 AND sb.currentqty > 0
		 ORDER BY sb.itemcode, sb.whcode`,
		shopID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	report := &m.InventoryValuationReport{
		ShopID:   shopID,
		AsOfDate: time.Now(),
	}

	for rows.Next() {
		var item m.InventoryValuationItem
		err := rows.Scan(&item.ItemCode, &item.WhCode, &item.Qty, &item.AverageCost, &item.TotalValue, &item.CostingMethod)
		if err != nil {
			return nil, err
		}
		report.Items = append(report.Items, item)
		report.TotalValue += item.TotalValue
	}

	return report, nil
}

// === Private Methods ===

// getEngineForProduct — ดึง engine ที่ตั้งค่าไว้สำหรับสินค้า
func (s *InventoryCostingService) getEngineForProduct(ctx context.Context, tx *sql.Tx, shopID, itemCode string) (costing.CostingEngine, error) {
	var method string
	err := tx.QueryRowContext(ctx,
		`SELECT COALESCE(costingmethod, 'moving_average') FROM product_costing_config WHERE shopid = $1 AND itemcode = $2`,
		shopID, itemCode,
	).Scan(&method)
	if err == sql.ErrNoRows {
		method = m.CostingMethodMovingAverage
	} else if err != nil {
		return nil, fmt.Errorf("ดึง costing method ไม่สำเร็จ: %w", err)
	}

	return costing.NewCostingEngine(method)
}

// withTransaction — wrap operation ใน DB transaction
func (s *InventoryCostingService) withTransaction(ctx context.Context, fn func(tx *sql.Tx) (*m.CostTransactionResult, error)) (*m.CostTransactionResult, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("เริ่ม transaction ไม่สำเร็จ: %w", err)
	}

	result, err := fn(tx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			logger.Error(fmt.Sprintf("rollback ไม่สำเร็จ: %v", rbErr))
		}
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit ไม่สำเร็จ: %w", err)
	}

	return result, nil
}
