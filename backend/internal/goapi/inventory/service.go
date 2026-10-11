package inventory

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"smlcloudplatform/internal/doctype"
	"smlcloudplatform/internal/goapi/inventory/costing"
	m "smlcloudplatform/internal/goapi/inventory/models"
	"smlcloudplatform/internal/goapi/logger"
)

// InventoryCostingService — service หลักสำหรับ orchestrate ทุก costing operation
// จัดการ DB transaction และประมวลผลต้นทุนถัวเฉลี่ย (วิธีเดียว ตามคำสั่งเจ้าของ 2026-10-11)
type InventoryCostingService struct {
	db     *sql.DB
	engine *costing.MovingAverageEngine
}

// NewInventoryCostingService — สร้าง service
func NewInventoryCostingService(db *sql.DB) *InventoryCostingService {
	return &InventoryCostingService{db: db, engine: &costing.MovingAverageEngine{}}
}

// === Public Methods ===

// ProcessReceipt — รับสินค้าเข้า (ซื้อ, รับโอน, ปรับเพิ่ม)
func (s *InventoryCostingService) ProcessReceipt(ctx context.Context, params m.ReceiptParams) (*m.CostTransactionResult, error) {
	return s.withTransaction(ctx, func(tx *sql.Tx) (*m.CostTransactionResult, error) {
		logger.Info(fmt.Sprintf("รับสินค้า %s: qty=%.4f cost=%.4f", params.ItemCode, params.Qty, params.UnitCost))
		return s.engine.ProcessReceipt(ctx, tx, params)
	})
}

// ProcessIssue — ตัดสินค้าออก (ขาย, เบิก)
func (s *InventoryCostingService) ProcessIssue(ctx context.Context, params m.IssueParams) (*m.CostTransactionResult, error) {
	return s.withTransaction(ctx, func(tx *sql.Tx) (*m.CostTransactionResult, error) {
		logger.Info(fmt.Sprintf("ตัดสินค้า %s: qty=%.4f", params.ItemCode, params.Qty))
		return s.engine.ProcessIssue(ctx, tx, params)
	})
}

// ProcessSalesReturn — รับคืนจากลูกค้า
func (s *InventoryCostingService) ProcessSalesReturn(ctx context.Context, params m.SalesReturnParams) (*m.CostTransactionResult, error) {
	return s.withTransaction(ctx, func(tx *sql.Tx) (*m.CostTransactionResult, error) {
		return s.engine.ProcessSalesReturn(ctx, tx, params)
	})
}

// ProcessPurchaseReturn — ส่งคืนสินค้าให้ supplier
func (s *InventoryCostingService) ProcessPurchaseReturn(ctx context.Context, params m.PurchaseReturnParams) (*m.CostTransactionResult, error) {
	return s.withTransaction(ctx, func(tx *sql.Tx) (*m.CostTransactionResult, error) {
		return s.engine.ProcessPurchaseReturn(ctx, tx, params)
	})
}

// ProcessTransfer — โอนย้ายคลัง (ตัดออกจากต้นทาง + รับเข้าปลายทาง)
func (s *InventoryCostingService) ProcessTransfer(ctx context.Context, params m.TransferParams) (*m.CostTransactionResult, error) {
	return s.withTransaction(ctx, func(tx *sql.Tx) (*m.CostTransactionResult, error) {

		// 1. ตัดออกจากคลังต้นทาง
		issueParams := m.IssueParams{
			HoldingCode: params.HoldingCode, ItemCode: params.ItemCode, Barcode: params.Barcode,
			WhCode: params.FromWhCode, LocationCode: params.FromLocationCode,
			Qty: params.Qty, RefDocType: params.RefDocType, RefDocNo: params.RefDocNo,
			TransFlag: params.TransFlag, LotNumber: params.LotNumber,
			TransactionDate: params.TransactionDate, CreatedBy: params.CreatedBy,
		}
		issueResult, err := s.engine.ProcessIssue(ctx, tx, issueParams)
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
			HoldingCode: params.HoldingCode, ItemCode: params.ItemCode, Barcode: params.Barcode,
			WhCode: params.ToWhCode, LocationCode: params.ToLocationCode,
			Qty: params.Qty, UnitCost: unitCost,
			RefDocType: params.RefDocType, RefDocNo: params.RefDocNo,
			TransFlag: params.TransFlag, LotNumber: params.LotNumber,
			ReceivedDate: params.TransactionDate, CreatedBy: params.CreatedBy,
		}
		receiptResult, err := s.engine.ProcessReceipt(ctx, tx, receiptParams)
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
		return s.engine.ProcessAdjustment(ctx, tx, params)
	})
}

// GetValuation — ดูมูลค่าสินค้าปัจจุบัน
func (s *InventoryCostingService) GetValuation(ctx context.Context, holdingCode, itemCode, whCode string) (*m.StockValuation, error) {
	var result *m.StockValuation
	_, err := s.withTransaction(ctx, func(tx *sql.Tx) (*m.CostTransactionResult, error) {
		var costByWarehouse bool
		err := tx.QueryRowContext(ctx,
			`SELECT COALESCE(costbywarehouse, false)
			 FROM productcostingconfig WHERE holdingcode = $1 AND itemcode = $2`,
			holdingCode, itemCode,
		).Scan(&costByWarehouse)
		if err == sql.ErrNoRows {
			costByWarehouse = false
		} else if err != nil {
			return nil, fmt.Errorf("ดึงตัวเลือกต้นทุนแยกคลังไม่สำเร็จ: %w", err)
		}
		valuationWhCode := whCode
		if !costByWarehouse {
			valuationWhCode = ""
		}
		result, err = s.engine.GetCurrentValuation(ctx, tx, holdingCode, itemCode, valuationWhCode)
		return nil, err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetCostingConfig — ดึงตั้งค่าต้นทุนของสินค้า (ไม่มีตัวเลือกวิธีคิดต้นทุน: ถัวเฉลี่ยวิธีเดียว)
func (s *InventoryCostingService) GetCostingConfig(ctx context.Context, holdingCode, itemCode string) (*m.ProductCostingConfig, error) {
	var config m.ProductCostingConfig
	err := s.db.QueryRowContext(ctx,
		`SELECT itemcode, lottrackingenabled, expirytrackingenabled,
		        expiryalertdays, autoblockexpired, allownegativestock,
		        COALESCE(costbywarehouse, false)
		 FROM productcostingconfig WHERE holdingcode = $1 AND itemcode = $2`,
		holdingCode, itemCode,
	).Scan(&config.ItemCode, &config.LotTrackingEnabled,
		&config.ExpiryTrackingEnabled, &config.ExpiryAlertDays,
		&config.AutoBlockExpired, &config.AllowNegativeStock,
		&config.CostByWarehouse)
	if err == sql.ErrNoRows {
		return &m.ProductCostingConfig{ItemCode: itemCode}, nil
	}
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// UpdateCostingConfig — บันทึกตั้งค่าต้นทุนของสินค้า
func (s *InventoryCostingService) UpdateCostingConfig(ctx context.Context, holdingCode string, config *m.ProductCostingConfig) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO productcostingconfig
		 (holdingcode, itemcode, lottrackingenabled, expirytrackingenabled,
		  expiryalertdays, autoblockexpired, allownegativestock, costbywarehouse, updatedat)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW())
		 ON CONFLICT (holdingcode, itemcode) DO UPDATE SET
		  lottrackingenabled = EXCLUDED.lottrackingenabled,
		  expirytrackingenabled = EXCLUDED.expirytrackingenabled,
		  expiryalertdays = EXCLUDED.expiryalertdays,
		  autoblockexpired = EXCLUDED.autoblockexpired,
		  allownegativestock = EXCLUDED.allownegativestock,
		  costbywarehouse = EXCLUDED.costbywarehouse,
		  updatedat = NOW()`,
		holdingCode, config.ItemCode,
		config.LotTrackingEnabled, config.ExpiryTrackingEnabled,
		config.ExpiryAlertDays, config.AutoBlockExpired,
		config.AllowNegativeStock, config.CostByWarehouse,
	)
	return err
}

// GetStockCard — รายงาน Stock Card (ประวัติเคลื่อนไหวสินค้า)
// StockMovementRow — บรรทัดรายงานการเคลื่อนไหวสต็อก
type StockMovementRow struct {
	DocDate       string  `json:"docdate"`
	DocNo         string  `json:"docno"`
	TransFlag     int     `json:"transflag"`
	TransName     string  `json:"transname"`
	ItemCode      string  `json:"itemcode"`
	ItemName      string  `json:"itemname"`
	WhCode        string  `json:"whcode"`
	UnitCode      string  `json:"unitcode"`
	InQty         float64 `json:"inqty"`
	OutQty        float64 `json:"outqty"`
	BalanceQty    float64 `json:"balanceqty"`
	UnitCost      float64 `json:"unitcost"`
	TotalCost     float64 `json:"totalcost"`
	BalanceAmount float64 `json:"balanceamount"`
	AvgCost       float64 `json:"avgcost"`
	DocRef        string  `json:"docref"`
}

// ErrCompanyRequired - stock_ledger / stock_period_balance เก็บหลายบริษัทในฐานเดียวกัน รายงานต้องกรองบริษัทเดียวเสมอ:
// รหัสบริษัทว่างคือข้อผิดพลาดของผู้เรียก ไม่ใช่ "ทุกบริษัท" และไม่ใช่บริษัท 01
var ErrCompanyRequired = errors.New("inventory report company is required")

// GetStockMovementReport — ประวัติการเคลื่อนไหวสต็อกของบริษัท businessCode จาก stock_ledger
func (s *InventoryCostingService) GetStockMovementReport(ctx context.Context, businessCode, itemCode string, fromDate, toDate time.Time) ([]StockMovementRow, error) {
	if businessCode == "" {
		return nil, ErrCompanyRequired
	}
	query := `SELECT 
		s.docdatetime::date::text AS docdate,
		s.docno,
		s.transflag,
		CASE 
			WHEN s.transflag = ` + strconv.Itoa(doctype.Purchase) + ` THEN 'ซื้อสินค้า'
			WHEN s.transflag = ` + strconv.Itoa(doctype.SaleInvoice) + ` THEN 'ขายสินค้า'
			WHEN s.transflag = ` + strconv.Itoa(doctype.StockPickup) + ` THEN 'เบิกสินค้า'
			WHEN s.transflag = ` + strconv.Itoa(doctype.SaleInvoiceReturn) + ` THEN 'รับคืนจากการขาย'
			WHEN s.transflag = ` + strconv.Itoa(doctype.PurchaseReturn) + ` THEN 'ส่งคืนจากการซื้อ'
			ELSE 'เคลื่อนไหวสต็อก'
		END AS transname,
		s.itemcode,
		COALESCE(d.description, s.itemcode) AS itemname,
		s.whcode,
		s.unitcode,
		CASE WHEN s.direction = 1 THEN s.docqty ELSE 0 END AS inqty,
		CASE WHEN s.direction = -1 THEN s.docqty ELSE 0 END AS outqty,
		s.balanceqty,
		s.unitcost,
		s.amount,
		s.balanceamount,
		s.avgcost,
		s.docref
	FROM stock_ledger s
	LEFT JOIN (SELECT DISTINCT ON (businesscode, docno, itemcode) businesscode, docno, itemcode, description FROM docdetail WHERE description IS NOT NULL AND description != '') d 
	       ON d.businesscode = s.businesscode AND d.docno = s.docno AND d.itemcode = s.itemcode
	WHERE s.businesscode = $1
	  AND s.docdatetime::date >= $2::date
	  AND s.docdatetime::date <= $3::date
	  AND ($4 = '' OR s.itemcode = $4)
	ORDER BY s.docdatetime ASC, s.behindindex ASC, s.docno ASC, s.linenumber ASC`

	rows, err := s.db.QueryContext(ctx, query, businessCode, fromDate, toDate, itemCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []StockMovementRow
	for rows.Next() {
		var r StockMovementRow
		if err := rows.Scan(&r.DocDate, &r.DocNo, &r.TransFlag, &r.TransName,
			&r.ItemCode, &r.ItemName, &r.WhCode, &r.UnitCode,
			&r.InQty, &r.OutQty, &r.BalanceQty, &r.UnitCost,
			&r.TotalCost, &r.BalanceAmount, &r.AvgCost, &r.DocRef); err != nil {
			return nil, err
		}
		results = append(results, r)
	}

	return results, nil
}

// GetStockCard — รายงาน Stock Card (ประวัติเคลื่อนไหวสินค้า) ของบริษัท businessCode
func (s *InventoryCostingService) GetStockCard(ctx context.Context, holdingCode, businessCode, itemCode string, fromDate, toDate time.Time) (*m.StockCardReport, error) {
	if businessCode == "" {
		return nil, ErrCompanyRequired
	}
	report := &m.StockCardReport{
		HoldingCode: holdingCode,
		ItemCode:    itemCode,
		FromDate:    fromDate,
		ToDate:      toDate,
	}

	// สมุดสต็อก (stock_ledger) ของบริษัทนี้เป็นแหล่งเดียว: บริษัทที่ไม่มีความเคลื่อนไหว = รายงานว่าง
	// (ไม่ถอยไปตารางเดิม inventorycosttransactions ที่กรองแค่ holding — จะได้ข้อมูลทุกบริษัทในกลุ่ม)
	rows, err := s.db.QueryContext(ctx,
		`SELECT s.docdatetime, s.docno,
		        CASE 
		            WHEN s.transflag = `+strconv.Itoa(doctype.Purchase)+` THEN 'purchase'
		            WHEN s.transflag = `+strconv.Itoa(doctype.SaleInvoice)+` THEN 'sale'
		            WHEN s.transflag = `+strconv.Itoa(doctype.StockPickup)+` THEN 'issue'
		            ELSE 'other'
		        END AS transactiontype,
		        s.transflag,
		        COALESCE(d.description, s.itemcode) AS itemname,
		        CASE WHEN s.direction = 1 THEN s.docqty ELSE 0 END AS inqty,
		        CASE WHEN s.direction = -1 THEN s.docqty ELSE 0 END AS outqty,
		        s.unitcost, s.amount, s.balanceqty, s.balanceamount, s.avgcost
		 FROM stock_ledger s
		 LEFT JOIN (SELECT DISTINCT ON (businesscode, docno, itemcode) businesscode, docno, itemcode, description FROM docdetail WHERE description IS NOT NULL AND description != '') d 
		        ON d.businesscode = s.businesscode AND d.docno = s.docno AND d.itemcode = s.itemcode
		 WHERE s.businesscode = $1
		   AND s.itemcode = $2
		   AND s.docdatetime::date >= $3::date
		   AND s.docdatetime::date <= $4::date
		 ORDER BY s.docdatetime ASC, s.behindindex ASC, s.docno ASC, s.linenumber ASC`,
		businessCode, itemCode, fromDate, toDate,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var entry m.StockCardEntry
		var itemName string
		if err := rows.Scan(&entry.Date, &entry.DocNo, &entry.TransactionType, &entry.TransFlag,
			&itemName, &entry.QtyIn, &entry.QtyOut, &entry.UnitCost, &entry.TotalCost,
			&entry.BalanceQty, &entry.BalanceValue, &entry.AverageCost); err != nil {
			return nil, err
		}
		if report.ItemName == "" {
			report.ItemName = itemName
		}
		report.Entries = append(report.Entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return report, nil
}

// GetInventoryValuation — รายงานมูลค่าสินค้าคงเหลือของบริษัท businessCode
func (s *InventoryCostingService) GetInventoryValuation(ctx context.Context, holdingCode, businessCode string) (*m.InventoryValuationReport, error) {
	if businessCode == "" {
		return nil, ErrCompanyRequired
	}
	report := &m.InventoryValuationReport{
		HoldingCode: holdingCode,
		AsOfDate:    time.Now(),
	}

	// ยอดยกไปของบริษัทนี้ (stock_period_balance) เป็นแหล่งเดียว: บริษัทที่ไม่มีสต็อก = รายงานว่าง
	// (ไม่ถอยไปตารางเดิม inventorystockbalances ที่กรองแค่ holding — จะได้สต็อกทุกบริษัทในกลุ่ม)
	// ตารางเก็บยอดปิดทุกงวดที่มีความเคลื่อนไหว (PK businesscode, itemcode, whcode, periodkey) → ใช้งวดล่าสุดของแต่ละสินค้า/คลัง
	// ไม่งั้นสินค้าที่เคลื่อนไหวหลายเดือนจะถูกนับยอดคงเหลือซ้ำทุกงวด
	rows, err := s.db.QueryContext(ctx,
		`SELECT spb.itemcode, COALESCE(d.description, spb.itemcode) AS itemname, spb.whcode, spb.closeqty, spb.closeavgcost, spb.closeamount
		 FROM (SELECT DISTINCT ON (itemcode, whcode) itemcode, whcode, closeqty, closeavgcost, closeamount
		       FROM stock_period_balance WHERE businesscode = $1
		       ORDER BY itemcode, whcode, periodkey DESC) spb
		 LEFT JOIN (SELECT DISTINCT ON (itemcode) itemcode, description FROM docdetail WHERE businesscode = $1 AND description IS NOT NULL AND description != '') d ON d.itemcode = spb.itemcode
		 WHERE spb.closeqty > 0
		 ORDER BY spb.itemcode, spb.whcode`,
		businessCode,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item m.InventoryValuationItem
		if err := rows.Scan(&item.ItemCode, &item.ItemName, &item.WhCode, &item.Qty, &item.AverageCost, &item.TotalValue); err != nil {
			return nil, err
		}
		report.Items = append(report.Items, item)
		report.TotalValue += item.TotalValue
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return report, nil
}

// === Private Methods ===

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
