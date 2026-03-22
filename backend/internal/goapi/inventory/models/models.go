package models

import (
	"time"
)

// === Costing Method Constants ===

const (
	CostingMethodMovingAverage  = "moving_average"
	CostingMethodPeriodicAverage = "periodic_average"
	CostingMethodFIFO           = "fifo"
	CostingMethodLIFO           = "lifo"
	CostingMethodFEFO           = "fefo"
	CostingMethodLot            = "lot"
	CostingMethodStandard       = "standard"
)

// === Layer Type Constants ===

const (
	LayerTypePurchase    = "purchase"
	LayerTypeTransferIn  = "transfer_in"
	LayerTypeAdjustment  = "adjustment"
	LayerTypeSalesReturn = "sales_return"
	LayerTypeProduction  = "production"
	LayerTypeOpening     = "opening"
)

// === Transaction Type Constants ===

const (
	TxTypePurchaseReceipt = "purchase_receipt"
	TxTypePurchaseReturn  = "purchase_return"
	TxTypeSalesIssue      = "sales_issue"
	TxTypeSalesReturn     = "sales_return"
	TxTypeTransferOut     = "transfer_out"
	TxTypeTransferIn      = "transfer_in"
	TxTypeAdjustmentIn    = "adjustment_in"
	TxTypeAdjustmentOut   = "adjustment_out"
	TxTypeWriteOff        = "write_off"
	TxTypeAssemblyIssue   = "assembly_issue"
	TxTypeAssemblyReceipt = "assembly_receipt"
	TxTypeCostAdjustment  = "cost_adjustment"
)

// === Variance Type Constants ===

const (
	VariancePurchasePrice = "purchase_price"
	VarianceRevaluation   = "revaluation"
	VarianceProduction    = "production"
	VarianceTransfer      = "transfer"
)

// === Quality Status Constants ===

const (
	QualityStatusApproved   = "approved"
	QualityStatusPending    = "pending"
	QualityStatusRejected   = "rejected"
	QualityStatusQuarantine = "quarantine"
)

// === Period Status Constants ===

const (
	PeriodStatusOpen   = "open"
	PeriodStatusClosed = "closed"
	PeriodStatusLocked = "locked"
)

// === Allocation Method Constants ===

const (
	AllocByValue    = "by_value"
	AllocByQuantity = "by_quantity"
	AllocByWeight   = "by_weight"
	AllocByVolume   = "by_volume"
	AllocManual     = "manual"
)

// === Core Models ===

// InventoryCostLayer — ชั้นต้นทุน (ใช้กับ FIFO/LIFO/FEFO/Lot)
// แต่ละครั้งที่รับสินค้าเข้าจะสร้าง layer ใหม่
type InventoryCostLayer struct {
	ID               int64      `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ShopID           string     `json:"shop_id" gorm:"column:shopid"`
	ItemCode         string     `json:"item_code" gorm:"column:itemcode"`
	Barcode          string     `json:"barcode" gorm:"column:barcode"`
	WhCode           string     `json:"wh_code" gorm:"column:whcode"`
	LocationCode     string     `json:"location_code" gorm:"column:locationcode"`
	LayerType        string     `json:"layer_type" gorm:"column:layertype"`
	RefDocType       string     `json:"ref_doc_type" gorm:"column:refdoctype"`
	RefDocNo         string     `json:"ref_doc_no" gorm:"column:refdocno"`
	OriginalQty      float64    `json:"original_qty" gorm:"column:originalqty"`
	RemainingQty     float64    `json:"remaining_qty" gorm:"column:remainingqty"`
	UnitCost         float64    `json:"unit_cost" gorm:"column:unitcost"`
	LandedCostPerUnit float64   `json:"landed_cost_per_unit" gorm:"column:landedcostperunit"`
	TotalUnitCost    float64    `json:"total_unit_cost" gorm:"column:totalunitcost"`
	LotNumber        string     `json:"lot_number" gorm:"column:lotnumber"`
	SupplierLotNumber string    `json:"supplier_lot_number" gorm:"column:supplierlotnumber"`
	ManufacturingDate *time.Time `json:"manufacturing_date" gorm:"column:manufacturingdate"`
	ExpiryDate       *time.Time `json:"expiry_date" gorm:"column:expirydate"`
	QualityStatus    string     `json:"quality_status" gorm:"column:qualitystatus"`
	ReceivedDate     time.Time  `json:"received_date" gorm:"column:receiveddate"`
	CreatedAt        time.Time  `json:"created_at" gorm:"column:createdat"`
	UpdatedAt        time.Time  `json:"updated_at" gorm:"column:updatedat"`
}

func (InventoryCostLayer) TableName() string { return "inventory_cost_layers" }

// InventoryStockBalance — ยอดคงเหลือต่อสินค้า × คลัง (สรุป snapshot)
type InventoryStockBalance struct {
	ID               int64      `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ShopID           string     `json:"shop_id" gorm:"column:shopid"`
	ItemCode         string     `json:"item_code" gorm:"column:itemcode"`
	Barcode          string     `json:"barcode" gorm:"column:barcode"`
	WhCode           string     `json:"wh_code" gorm:"column:whcode"`
	LocationCode     string     `json:"location_code" gorm:"column:locationcode"`
	CurrentQty       float64    `json:"current_qty" gorm:"column:currentqty"`
	CurrentAvgCost   float64    `json:"current_avg_cost" gorm:"column:currentavgcost"`
	CurrentTotalValue float64   `json:"current_total_value" gorm:"column:currenttotalvalue"`
	LastPurchaseCost float64    `json:"last_purchase_cost" gorm:"column:lastpurchasecost"`
	LastPurchaseDate *time.Time `json:"last_purchase_date" gorm:"column:lastpurchasedate"`
	UpdatedAt        time.Time  `json:"updated_at" gorm:"column:updatedat"`
}

func (InventoryStockBalance) TableName() string { return "inventory_stock_balances" }

// InventoryCostTransaction — ประวัติทุก transaction ที่กระทบต้นทุน
type InventoryCostTransaction struct {
	ID                int64      `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ShopID            string     `json:"shop_id" gorm:"column:shopid"`
	ItemCode          string     `json:"item_code" gorm:"column:itemcode"`
	Barcode           string     `json:"barcode" gorm:"column:barcode"`
	WhCode            string     `json:"wh_code" gorm:"column:whcode"`
	LocationCode      string     `json:"location_code" gorm:"column:locationcode"`
	TransactionType   string     `json:"transaction_type" gorm:"column:transactiontype"`
	TransFlag         int        `json:"trans_flag" gorm:"column:transflag"`
	RefDocType        string     `json:"ref_doc_type" gorm:"column:refdoctype"`
	RefDocNo          string     `json:"ref_doc_no" gorm:"column:refdocno"`
	Qty               float64    `json:"qty" gorm:"column:qty"`
	UnitCost          float64    `json:"unit_cost" gorm:"column:unitcost"`
	TotalCost         float64    `json:"total_cost" gorm:"column:totalcost"`
	LandedCost        float64    `json:"landed_cost" gorm:"column:landedcost"`
	LotNumber         string     `json:"lot_number" gorm:"column:lotnumber"`
	ExpiryDate        *time.Time `json:"expiry_date" gorm:"column:expirydate"`
	CostLayerID       *int64     `json:"cost_layer_id" gorm:"column:costlayerid"`
	BalanceQty        float64    `json:"balance_qty" gorm:"column:balanceqty"`
	BalanceAvgCost    float64    `json:"balance_avg_cost" gorm:"column:balanceavgcost"`
	BalanceTotalValue float64    `json:"balance_total_value" gorm:"column:balancetotalvalue"`
	CostingMethodUsed string     `json:"costing_method_used" gorm:"column:costingmethodused"`
	TransactionDate   time.Time  `json:"transaction_date" gorm:"column:transactiondate"`
	AccountingPeriod  string     `json:"accounting_period" gorm:"column:accountingperiod"`
	CreatedBy         string     `json:"created_by" gorm:"column:createdby"`
	CreatedAt         time.Time  `json:"created_at" gorm:"column:createdat"`
	IsReversed        bool       `json:"is_reversed" gorm:"column:isreversed"`
	ReversedByID      *int64     `json:"reversed_by_id" gorm:"column:reversedbyid"`
}

func (InventoryCostTransaction) TableName() string { return "inventory_cost_transactions" }

// InventoryVariance — ผลต่างต้นทุน (Standard Cost)
type InventoryVariance struct {
	ID              int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ShopID          string    `json:"shop_id" gorm:"column:shopid"`
	ItemCode        string    `json:"item_code" gorm:"column:itemcode"`
	WhCode          string    `json:"wh_code" gorm:"column:whcode"`
	VarianceType    string    `json:"variance_type" gorm:"column:variancetype"`
	RefDocType      string    `json:"ref_doc_type" gorm:"column:refdoctype"`
	RefDocNo        string    `json:"ref_doc_no" gorm:"column:refdocno"`
	StandardCost    float64   `json:"standard_cost" gorm:"column:standardcost"`
	ActualCost      float64   `json:"actual_cost" gorm:"column:actualcost"`
	Qty             float64   `json:"qty" gorm:"column:qty"`
	VarianceAmount  float64   `json:"variance_amount" gorm:"column:varianceamount"`
	TransactionDate time.Time `json:"transaction_date" gorm:"column:transactiondate"`
	AccountingPeriod string   `json:"accounting_period" gorm:"column:accountingperiod"`
	CreatedAt       time.Time `json:"created_at" gorm:"column:createdat"`
}

func (InventoryVariance) TableName() string { return "inventory_variances" }

// InventoryLandedCost — ค่าใช้จ่ายประกอบ (ค่าขนส่ง, ภาษี, ประกัน)
type InventoryLandedCost struct {
	ID               int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ShopID           string    `json:"shop_id" gorm:"column:shopid"`
	RefDocType       string    `json:"ref_doc_type" gorm:"column:refdoctype"`
	RefDocNo         string    `json:"ref_doc_no" gorm:"column:refdocno"`
	CostType         string    `json:"cost_type" gorm:"column:costtype"`
	Description      string    `json:"description" gorm:"column:description"`
	TotalAmount      float64   `json:"total_amount" gorm:"column:totalamount"`
	Currency         string    `json:"currency" gorm:"column:currency"`
	ExchangeRate     float64   `json:"exchange_rate" gorm:"column:exchangerate"`
	AmountTHB        float64   `json:"amount_thb" gorm:"column:amountthb"`
	AllocationMethod string    `json:"allocation_method" gorm:"column:allocationmethod"`
	IsAllocated      bool      `json:"is_allocated" gorm:"column:isallocated"`
	AllocatedAt      *time.Time `json:"allocated_at" gorm:"column:allocatedat"`
	CreatedAt        time.Time  `json:"created_at" gorm:"column:createdat"`
}

func (InventoryLandedCost) TableName() string { return "inventory_landed_costs" }

// InventoryLandedCostAllocation — การกระจาย landed cost ไปแต่ละสินค้า
type InventoryLandedCostAllocation struct {
	ID               int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ShopID           string    `json:"shop_id" gorm:"column:shopid"`
	LandedCostID     int64     `json:"landed_cost_id" gorm:"column:landedcostid"`
	ItemCode         string    `json:"item_code" gorm:"column:itemcode"`
	WhCode           string    `json:"wh_code" gorm:"column:whcode"`
	CostLayerID      *int64    `json:"cost_layer_id" gorm:"column:costlayerid"`
	AllocatedAmount  float64   `json:"allocated_amount" gorm:"column:allocatedamount"`
	AllocatedPerUnit float64   `json:"allocated_per_unit" gorm:"column:allocatedperunit"`
	Qty              float64   `json:"qty" gorm:"column:qty"`
	CreatedAt        time.Time `json:"created_at" gorm:"column:createdat"`
}

func (InventoryLandedCostAllocation) TableName() string { return "inventory_landed_cost_allocations" }

// InventoryAccountingPeriod — งวดบัญชี
type InventoryAccountingPeriod struct {
	ID                  int64      `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ShopID              string     `json:"shop_id" gorm:"column:shopid"`
	PeriodCode          string     `json:"period_code" gorm:"column:periodcode"`
	StartDate           time.Time  `json:"start_date" gorm:"column:startdate"`
	EndDate             time.Time  `json:"end_date" gorm:"column:enddate"`
	Status              string     `json:"status" gorm:"column:status"`
	PeriodicAvgCalculated bool     `json:"periodic_avg_calculated" gorm:"column:periodicavgcalculated"`
	ClosedBy            string     `json:"closed_by" gorm:"column:closedby"`
	ClosedAt            *time.Time `json:"closed_at" gorm:"column:closedat"`
	CreatedAt           time.Time  `json:"created_at" gorm:"column:createdat"`
}

func (InventoryAccountingPeriod) TableName() string { return "inventory_accounting_periods" }

// === Request/Response Structs ===

// ReceiptParams — parameter สำหรับรับสินค้าเข้า
type ReceiptParams struct {
	ShopID       string    `json:"shop_id"`
	ItemCode     string    `json:"item_code"`
	Barcode      string    `json:"barcode"`
	WhCode       string    `json:"wh_code"`
	LocationCode string    `json:"location_code"`
	Qty          float64   `json:"qty"`
	UnitCost     float64   `json:"unit_cost"`
	RefDocType   string    `json:"ref_doc_type"`
	RefDocNo     string    `json:"ref_doc_no"`
	TransFlag    int       `json:"trans_flag"`
	LotNumber    string    `json:"lot_number,omitempty"`
	ExpiryDate   *time.Time `json:"expiry_date,omitempty"`
	ReceivedDate time.Time  `json:"received_date"`
	CreatedBy    string    `json:"created_by"`
}

// IssueParams — parameter สำหรับตัดสินค้าออก (ขาย/เบิก)
type IssueParams struct {
	ShopID       string    `json:"shop_id"`
	ItemCode     string    `json:"item_code"`
	Barcode      string    `json:"barcode"`
	WhCode       string    `json:"wh_code"`
	LocationCode string    `json:"location_code"`
	Qty          float64   `json:"qty"`
	RefDocType   string    `json:"ref_doc_type"`
	RefDocNo     string    `json:"ref_doc_no"`
	TransFlag    int       `json:"trans_flag"`
	LotNumber    string    `json:"lot_number,omitempty"`
	TransactionDate time.Time `json:"transaction_date"`
	CreatedBy    string    `json:"created_by"`
}

// TransferParams — parameter สำหรับโอนย้ายคลัง
type TransferParams struct {
	ShopID          string    `json:"shop_id"`
	ItemCode        string    `json:"item_code"`
	Barcode         string    `json:"barcode"`
	FromWhCode      string    `json:"from_wh_code"`
	FromLocationCode string   `json:"from_location_code"`
	ToWhCode        string    `json:"to_wh_code"`
	ToLocationCode  string    `json:"to_location_code"`
	Qty             float64   `json:"qty"`
	RefDocType      string    `json:"ref_doc_type"`
	RefDocNo        string    `json:"ref_doc_no"`
	TransFlag       int       `json:"trans_flag"`
	LotNumber       string    `json:"lot_number,omitempty"`
	TransactionDate time.Time `json:"transaction_date"`
	CreatedBy       string    `json:"created_by"`
}

// AdjustmentParams — parameter สำหรับปรับปรุง stock
type AdjustmentParams struct {
	ShopID       string    `json:"shop_id"`
	ItemCode     string    `json:"item_code"`
	Barcode      string    `json:"barcode"`
	WhCode       string    `json:"wh_code"`
	LocationCode string    `json:"location_code"`
	Qty          float64   `json:"qty"`
	UnitCost     float64   `json:"unit_cost"`
	IsIncrease   bool      `json:"is_increase"`
	RefDocType   string    `json:"ref_doc_type"`
	RefDocNo     string    `json:"ref_doc_no"`
	TransFlag    int       `json:"trans_flag"`
	LotNumber    string    `json:"lot_number,omitempty"`
	ExpiryDate   *time.Time `json:"expiry_date,omitempty"`
	TransactionDate time.Time `json:"transaction_date"`
	CreatedBy    string    `json:"created_by"`
}

// SalesReturnParams — parameter สำหรับรับคืนจากลูกค้า
type SalesReturnParams struct {
	ShopID       string    `json:"shop_id"`
	ItemCode     string    `json:"item_code"`
	Barcode      string    `json:"barcode"`
	WhCode       string    `json:"wh_code"`
	LocationCode string    `json:"location_code"`
	Qty          float64   `json:"qty"`
	OriginalCost float64   `json:"original_cost"`
	RefDocType   string    `json:"ref_doc_type"`
	RefDocNo     string    `json:"ref_doc_no"`
	TransFlag    int       `json:"trans_flag"`
	LotNumber    string    `json:"lot_number,omitempty"`
	TransactionDate time.Time `json:"transaction_date"`
	CreatedBy    string    `json:"created_by"`
}

// PurchaseReturnParams — parameter สำหรับส่งคืนสินค้าให้ supplier
type PurchaseReturnParams struct {
	ShopID       string    `json:"shop_id"`
	ItemCode     string    `json:"item_code"`
	Barcode      string    `json:"barcode"`
	WhCode       string    `json:"wh_code"`
	LocationCode string    `json:"location_code"`
	Qty          float64   `json:"qty"`
	RefDocType   string    `json:"ref_doc_type"`
	RefDocNo     string    `json:"ref_doc_no"`
	TransFlag    int       `json:"trans_flag"`
	CostLayerID  *int64    `json:"cost_layer_id,omitempty"`
	LotNumber    string    `json:"lot_number,omitempty"`
	TransactionDate time.Time `json:"transaction_date"`
	CreatedBy    string    `json:"created_by"`
}

// StockValuation — ผลลัพธ์มูลค่าสินค้าคงเหลือ
type StockValuation struct {
	ItemCode       string  `json:"item_code"`
	WhCode         string  `json:"wh_code"`
	CurrentQty     float64 `json:"current_qty"`
	AverageCost    float64 `json:"average_cost"`
	TotalValue     float64 `json:"total_value"`
	CostingMethod  string  `json:"costing_method"`
}

// CostTransactionResult — ผลลัพธ์หลังประมวลผล transaction
type CostTransactionResult struct {
	Transaction    *InventoryCostTransaction `json:"transaction"`
	CostLayer      *InventoryCostLayer       `json:"cost_layer,omitempty"`
	Variance       *InventoryVariance        `json:"variance,omitempty"`
	BalanceQty     float64                   `json:"balance_qty"`
	BalanceAvgCost float64                   `json:"balance_avg_cost"`
	TotalValue     float64                   `json:"total_value"`
}

// ProductCostingConfig — ข้อมูล costing method ต่อสินค้า (ดึงจาก product table)
type ProductCostingConfig struct {
	ItemCode              string   `json:"item_code"`
	CostingMethod         string   `json:"costing_method"`
	LotTrackingEnabled    bool     `json:"lot_tracking_enabled"`
	ExpiryTrackingEnabled bool     `json:"expiry_tracking_enabled"`
	ExpiryAlertDays       int      `json:"expiry_alert_days"`
	AutoBlockExpired      bool     `json:"auto_block_expired"`
	AllowNegativeStock    bool     `json:"allow_negative_stock"`
	StandardCost          float64  `json:"standard_cost"`
}

// InventoryValuationReport — รายงานมูลค่าสินค้าคงเหลือ
type InventoryValuationReport struct {
	ShopID     string                      `json:"shop_id"`
	AsOfDate   time.Time                   `json:"as_of_date"`
	Items      []InventoryValuationItem    `json:"items"`
	TotalValue float64                     `json:"total_value"`
}

type InventoryValuationItem struct {
	ItemCode      string  `json:"item_code"`
	ItemName      string  `json:"item_name"`
	WhCode        string  `json:"wh_code"`
	UnitCode      string  `json:"unit_code"`
	Qty           float64 `json:"qty"`
	AverageCost   float64 `json:"average_cost"`
	TotalValue    float64 `json:"total_value"`
	CostingMethod string  `json:"costing_method"`
}

// StockCardReport — รายงาน Stock Card (รายการเคลื่อนไหวต่อสินค้า)
type StockCardReport struct {
	ShopID    string           `json:"shop_id"`
	ItemCode  string           `json:"item_code"`
	ItemName  string           `json:"item_name"`
	FromDate  time.Time        `json:"from_date"`
	ToDate    time.Time        `json:"to_date"`
	Entries   []StockCardEntry `json:"entries"`
}

type StockCardEntry struct {
	Date            time.Time `json:"date"`
	DocNo           string    `json:"doc_no"`
	TransactionType string   `json:"transaction_type"`
	TransFlag       int      `json:"trans_flag"`
	QtyIn           float64  `json:"qty_in"`
	QtyOut          float64  `json:"qty_out"`
	UnitCost        float64  `json:"unit_cost"`
	TotalCost       float64  `json:"total_cost"`
	BalanceQty      float64  `json:"balance_qty"`
	BalanceValue    float64  `json:"balance_value"`
	AverageCost     float64  `json:"average_cost"`
}
