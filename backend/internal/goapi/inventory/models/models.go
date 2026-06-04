package models

import (
	"time"
)

// === Costing Method Constants ===

const (
	CostingMethodMovingAverage   = "moving_average"
	CostingMethodPeriodicAverage = "periodic_average"
	CostingMethodFIFO            = "fifo"
	CostingMethodLIFO            = "lifo"
	CostingMethodFEFO            = "fefo"
	CostingMethodLot             = "lot"
	CostingMethodStandard        = "standard"
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
	ID                int64      `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	HoldingCode       string     `json:"holding_code" gorm:"column:holding_code"`
	ItemCode          string     `json:"item_code" gorm:"column:itemcode"`
	Barcode           string     `json:"barcode" gorm:"column:barcode"`
	WhCode            string     `json:"wh_code" gorm:"column:whcode"`
	LocationCode      string     `json:"location_code" gorm:"column:locationcode"`
	LayerType         string     `json:"layer_type" gorm:"column:layer_type"`
	RefDocType        string     `json:"ref_doc_type" gorm:"column:ref_doc_type"`
	RefDocNo          string     `json:"ref_doc_no" gorm:"column:ref_doc_no"`
	OriginalQty       float64    `json:"original_qty" gorm:"column:originalqty"`
	RemainingQty      float64    `json:"remaining_qty" gorm:"column:remainingqty"`
	UnitCost          float64    `json:"unit_cost" gorm:"column:unitcost"`
	LandedCostPerUnit float64    `json:"landed_cost_per_unit" gorm:"column:landedcostperunit"`
	TotalUnitCost     float64    `json:"total_unit_cost" gorm:"column:totalunitcost"`
	LotNumber         string     `json:"lot_number" gorm:"column:lot_number"`
	SupplierLotNumber string     `json:"supplier_lot_number" gorm:"column:supplier_lot_number"`
	ManufacturingDate *time.Time `json:"manufacturing_date" gorm:"column:manufacturing_date"`
	ExpiryDate        *time.Time `json:"expiry_date" gorm:"column:expiry_date"`
	QualityStatus     string     `json:"quality_status" gorm:"column:qualitystatus"`
	ReceivedDate      time.Time  `json:"received_date" gorm:"column:received_date"`
	CreatedAt         time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"column:updated_at"`
}

func (InventoryCostLayer) TableName() string { return "inventory_cost_layers" }

// InventoryStockBalance — ยอดคงเหลือต่อสินค้า × คลัง (สรุป snapshot)
type InventoryStockBalance struct {
	ID                int64      `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	HoldingCode       string     `json:"holding_code" gorm:"column:holding_code"`
	ItemCode          string     `json:"item_code" gorm:"column:itemcode"`
	Barcode           string     `json:"barcode" gorm:"column:barcode"`
	WhCode            string     `json:"wh_code" gorm:"column:whcode"`
	LocationCode      string     `json:"location_code" gorm:"column:locationcode"`
	CurrentQty        float64    `json:"current_qty" gorm:"column:currentqty"`
	ReservedQty       float64    `json:"reserved_qty" gorm:"column:reservedqty"`
	CurrentAvgCost    float64    `json:"current_avg_cost" gorm:"column:currentavgcost"`
	CurrentTotalValue float64    `json:"current_total_value" gorm:"column:currenttotalvalue"`
	LastPurchaseCost  float64    `json:"last_purchase_cost" gorm:"column:last_purchase_cost"`
	LastPurchaseDate  *time.Time `json:"last_purchase_date" gorm:"column:last_purchase_date"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"column:updated_at"`
}

func (InventoryStockBalance) TableName() string { return "inventory_stock_balances" }

// MarketplaceStockBalance stores marketplace availability by product dimensions.
// This projection is not an accounting cost source.
type MarketplaceStockBalance struct {
	ID              int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	HoldingCode     string    `json:"holding_code" gorm:"column:holding_code"`
	ItemCode        string    `json:"item_code" gorm:"column:item_code"`
	DimensionKey    string    `json:"dimension_key" gorm:"column:dimension_key"`
	DimensionValues string    `json:"dimension_values" gorm:"column:dimension_values"`
	CurrentQty      float64   `json:"current_qty" gorm:"column:current_qty"`
	ReservedQty     float64   `json:"reserved_qty" gorm:"column:reserved_qty"`
	AvailableQty    float64   `json:"available_qty" gorm:"column:available_qty"`
	Source          string    `json:"source" gorm:"column:source"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (MarketplaceStockBalance) TableName() string { return "marketplace_stock_balances" }

// MarketplaceDimensionPrice stores selling prices by product dimension and marketplace.
type MarketplaceDimensionPrice struct {
	ID              int64      `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	HoldingCode     string     `json:"holding_code" gorm:"column:holding_code"`
	ItemCode        string     `json:"item_code" gorm:"column:item_code"`
	Barcode         string     `json:"barcode" gorm:"column:barcode"`
	DimensionKey    string     `json:"dimension_key" gorm:"column:dimension_key"`
	DimensionValues string     `json:"dimension_values" gorm:"column:dimension_values"`
	PriceLevel      string     `json:"price_level" gorm:"column:price_level"`
	Marketplace     string     `json:"marketplace" gorm:"column:marketplace"`
	Currency        string     `json:"currency" gorm:"column:currency"`
	Price           float64    `json:"price" gorm:"column:price"`
	SalePrice       float64    `json:"sale_price" gorm:"column:sale_price"`
	CompareAtPrice  float64    `json:"compare_at_price" gorm:"column:compare_at_price"`
	EffectiveFrom   *time.Time `json:"effective_from" gorm:"column:effective_from"`
	EffectiveTo     *time.Time `json:"effective_to" gorm:"column:effective_to"`
	IsActive        bool       `json:"is_active" gorm:"column:is_active"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"column:updated_at"`
}

func (MarketplaceDimensionPrice) TableName() string { return "marketplace_dimension_prices" }

// InventoryCostTransaction — ประวัติทุก transaction ที่กระทบต้นทุน
type InventoryCostTransaction struct {
	ID                int64      `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	HoldingCode       string     `json:"holding_code" gorm:"column:holding_code"`
	ItemCode          string     `json:"item_code" gorm:"column:itemcode"`
	Barcode           string     `json:"barcode" gorm:"column:barcode"`
	WhCode            string     `json:"wh_code" gorm:"column:whcode"`
	LocationCode      string     `json:"location_code" gorm:"column:locationcode"`
	TransactionType   string     `json:"transaction_type" gorm:"column:transactiontype"`
	TransFlag         int        `json:"trans_flag" gorm:"column:transflag"`
	RefDocType        string     `json:"ref_doc_type" gorm:"column:ref_doc_type"`
	RefDocNo          string     `json:"ref_doc_no" gorm:"column:ref_doc_no"`
	Qty               float64    `json:"qty" gorm:"column:qty"`
	UnitCost          float64    `json:"unit_cost" gorm:"column:unitcost"`
	TotalCost         float64    `json:"total_cost" gorm:"column:total_cost"`
	LandedCost        float64    `json:"landed_cost" gorm:"column:landed_cost"`
	LotNumber         string     `json:"lot_number" gorm:"column:lot_number"`
	ExpiryDate        *time.Time `json:"expiry_date" gorm:"column:expiry_date"`
	CostLayerID       *int64     `json:"cost_layer_id" gorm:"column:cost_layer_id"`
	BalanceQty        float64    `json:"balance_qty" gorm:"column:balance_qty"`
	BalanceAvgCost    float64    `json:"balance_avg_cost" gorm:"column:balanceavgcost"`
	BalanceTotalValue float64    `json:"balance_total_value" gorm:"column:balancetotalvalue"`
	CostingMethodUsed string     `json:"costing_method_used" gorm:"column:costingmethodused"`
	TransactionDate   time.Time  `json:"transaction_date" gorm:"column:transaction_date"`
	AccountingPeriod  string     `json:"accounting_period" gorm:"column:accountingperiod"`
	CreatedBy         string     `json:"created_by" gorm:"column:createdby"`
	CreatedAt         time.Time  `json:"created_at" gorm:"column:created_at"`
	IsReversed        bool       `json:"is_reversed" gorm:"column:isreversed"`
	ReversedByID      *int64     `json:"reversed_by_id" gorm:"column:reversed_by_id"`
}

func (InventoryCostTransaction) TableName() string { return "inventory_cost_transactions" }

// InventoryVariance — ผลต่างต้นทุน (Standard Cost)
type InventoryVariance struct {
	ID               int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	HoldingCode      string    `json:"holding_code" gorm:"column:holding_code"`
	ItemCode         string    `json:"item_code" gorm:"column:itemcode"`
	WhCode           string    `json:"wh_code" gorm:"column:whcode"`
	VarianceType     string    `json:"variance_type" gorm:"column:variance_type"`
	RefDocType       string    `json:"ref_doc_type" gorm:"column:ref_doc_type"`
	RefDocNo         string    `json:"ref_doc_no" gorm:"column:ref_doc_no"`
	StandardCost     float64   `json:"standard_cost" gorm:"column:standardcost"`
	ActualCost       float64   `json:"actual_cost" gorm:"column:actualcost"`
	Qty              float64   `json:"qty" gorm:"column:qty"`
	VarianceAmount   float64   `json:"variance_amount" gorm:"column:variance_amount"`
	TransactionDate  time.Time `json:"transaction_date" gorm:"column:transaction_date"`
	AccountingPeriod string    `json:"accounting_period" gorm:"column:accountingperiod"`
	CreatedAt        time.Time `json:"created_at" gorm:"column:created_at"`
}

func (InventoryVariance) TableName() string { return "inventory_variances" }

// InventoryLandedCost — ค่าใช้จ่ายประกอบ (ค่าขนส่ง, ภาษี, ประกัน)
type InventoryLandedCost struct {
	ID               int64      `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	HoldingCode      string     `json:"holding_code" gorm:"column:holding_code"`
	RefDocType       string     `json:"ref_doc_type" gorm:"column:ref_doc_type"`
	RefDocNo         string     `json:"ref_doc_no" gorm:"column:ref_doc_no"`
	CostType         string     `json:"cost_type" gorm:"column:cost_type"`
	Description      string     `json:"description" gorm:"column:description"`
	TotalAmount      float64    `json:"total_amount" gorm:"column:total_amount"`
	Currency         string     `json:"currency" gorm:"column:currency"`
	ExchangeRate     float64    `json:"exchange_rate" gorm:"column:exchange_rate"`
	AmountTHB        float64    `json:"amount_thb" gorm:"column:amountthb"`
	AllocationMethod string     `json:"allocation_method" gorm:"column:allocationmethod"`
	IsAllocated      bool       `json:"is_allocated" gorm:"column:isallocated"`
	AllocatedAt      *time.Time `json:"allocated_at" gorm:"column:allocated_at"`
	CreatedAt        time.Time  `json:"created_at" gorm:"column:created_at"`
}

func (InventoryLandedCost) TableName() string { return "inventory_landed_costs" }

// InventoryLandedCostAllocation — การกระจาย landed cost ไปแต่ละสินค้า
type InventoryLandedCostAllocation struct {
	ID               int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	HoldingCode      string    `json:"holding_code" gorm:"column:holding_code"`
	LandedCostID     int64     `json:"landed_cost_id" gorm:"column:landed_cost_id"`
	ItemCode         string    `json:"item_code" gorm:"column:itemcode"`
	WhCode           string    `json:"wh_code" gorm:"column:whcode"`
	CostLayerID      *int64    `json:"cost_layer_id" gorm:"column:cost_layer_id"`
	AllocatedAmount  float64   `json:"allocated_amount" gorm:"column:allocated_amount"`
	AllocatedPerUnit float64   `json:"allocated_per_unit" gorm:"column:allocatedperunit"`
	Qty              float64   `json:"qty" gorm:"column:qty"`
	CreatedAt        time.Time `json:"created_at" gorm:"column:created_at"`
}

func (InventoryLandedCostAllocation) TableName() string { return "inventory_landed_cost_allocations" }

// InventoryAccountingPeriod — งวดบัญชี
type InventoryAccountingPeriod struct {
	ID                    int64      `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	HoldingCode           string     `json:"holding_code" gorm:"column:holding_code"`
	PeriodCode            string     `json:"period_code" gorm:"column:period_code"`
	StartDate             time.Time  `json:"start_date" gorm:"column:start_date"`
	EndDate               time.Time  `json:"end_date" gorm:"column:end_date"`
	Status                string     `json:"status" gorm:"column:status"`
	PeriodicAvgCalculated bool       `json:"periodic_avg_calculated" gorm:"column:periodicavgcalculated"`
	ClosedBy              string     `json:"closed_by" gorm:"column:closed_by"`
	ClosedAt              *time.Time `json:"closed_at" gorm:"column:closed_at"`
	CreatedAt             time.Time  `json:"created_at" gorm:"column:created_at"`
}

func (InventoryAccountingPeriod) TableName() string { return "inventory_accounting_periods" }

// === Request/Response Structs ===

// ReceiptParams — parameter สำหรับรับสินค้าเข้า
type ReceiptParams struct {
	HoldingCode  string     `json:"holding_code"`
	ItemCode     string     `json:"item_code"`
	Barcode      string     `json:"barcode"`
	WhCode       string     `json:"wh_code"`
	LocationCode string     `json:"location_code"`
	Qty          float64    `json:"qty"`
	UnitCost     float64    `json:"unit_cost"`
	RefDocType   string     `json:"ref_doc_type"`
	RefDocNo     string     `json:"ref_doc_no"`
	TransFlag    int        `json:"trans_flag"`
	LotNumber    string     `json:"lot_number,omitempty"`
	ExpiryDate   *time.Time `json:"expiry_date,omitempty"`
	ReceivedDate time.Time  `json:"received_date"`
	CreatedBy    string     `json:"created_by"`
}

// IssueParams — parameter สำหรับตัดสินค้าออก (ขาย/เบิก)
type IssueParams struct {
	HoldingCode     string    `json:"holding_code"`
	ItemCode        string    `json:"item_code"`
	Barcode         string    `json:"barcode"`
	WhCode          string    `json:"wh_code"`
	LocationCode    string    `json:"location_code"`
	Qty             float64   `json:"qty"`
	RefDocType      string    `json:"ref_doc_type"`
	RefDocNo        string    `json:"ref_doc_no"`
	TransFlag       int       `json:"trans_flag"`
	LotNumber       string    `json:"lot_number,omitempty"`
	TransactionDate time.Time `json:"transaction_date"`
	CreatedBy       string    `json:"created_by"`
}

// TransferParams — parameter สำหรับโอนย้ายคลัง
type TransferParams struct {
	HoldingCode      string    `json:"holding_code"`
	ItemCode         string    `json:"item_code"`
	Barcode          string    `json:"barcode"`
	FromWhCode       string    `json:"from_wh_code"`
	FromLocationCode string    `json:"from_location_code"`
	ToWhCode         string    `json:"to_wh_code"`
	ToLocationCode   string    `json:"to_location_code"`
	Qty              float64   `json:"qty"`
	RefDocType       string    `json:"ref_doc_type"`
	RefDocNo         string    `json:"ref_doc_no"`
	TransFlag        int       `json:"trans_flag"`
	LotNumber        string    `json:"lot_number,omitempty"`
	TransactionDate  time.Time `json:"transaction_date"`
	CreatedBy        string    `json:"created_by"`
}

// AdjustmentParams — parameter สำหรับปรับปรุง stock
type AdjustmentParams struct {
	HoldingCode     string     `json:"holding_code"`
	ItemCode        string     `json:"item_code"`
	Barcode         string     `json:"barcode"`
	WhCode          string     `json:"wh_code"`
	LocationCode    string     `json:"location_code"`
	Qty             float64    `json:"qty"`
	UnitCost        float64    `json:"unit_cost"`
	IsIncrease      bool       `json:"is_increase"`
	RefDocType      string     `json:"ref_doc_type"`
	RefDocNo        string     `json:"ref_doc_no"`
	TransFlag       int        `json:"trans_flag"`
	LotNumber       string     `json:"lot_number,omitempty"`
	ExpiryDate      *time.Time `json:"expiry_date,omitempty"`
	TransactionDate time.Time  `json:"transaction_date"`
	CreatedBy       string     `json:"created_by"`
}

// SalesReturnParams — parameter สำหรับรับคืนจากลูกค้า
type SalesReturnParams struct {
	HoldingCode     string    `json:"holding_code"`
	ItemCode        string    `json:"item_code"`
	Barcode         string    `json:"barcode"`
	WhCode          string    `json:"wh_code"`
	LocationCode    string    `json:"location_code"`
	Qty             float64   `json:"qty"`
	OriginalCost    float64   `json:"original_cost"`
	RefDocType      string    `json:"ref_doc_type"`
	RefDocNo        string    `json:"ref_doc_no"`
	TransFlag       int       `json:"trans_flag"`
	LotNumber       string    `json:"lot_number,omitempty"`
	TransactionDate time.Time `json:"transaction_date"`
	CreatedBy       string    `json:"created_by"`
}

// PurchaseReturnParams — parameter สำหรับส่งคืนสินค้าให้ supplier
type PurchaseReturnParams struct {
	HoldingCode     string    `json:"holding_code"`
	ItemCode        string    `json:"item_code"`
	Barcode         string    `json:"barcode"`
	WhCode          string    `json:"wh_code"`
	LocationCode    string    `json:"location_code"`
	Qty             float64   `json:"qty"`
	RefDocType      string    `json:"ref_doc_type"`
	RefDocNo        string    `json:"ref_doc_no"`
	TransFlag       int       `json:"trans_flag"`
	CostLayerID     *int64    `json:"cost_layer_id,omitempty"`
	LotNumber       string    `json:"lot_number,omitempty"`
	TransactionDate time.Time `json:"transaction_date"`
	CreatedBy       string    `json:"created_by"`
}

// StockValuation — ผลลัพธ์มูลค่าสินค้าคงเหลือ
type StockValuation struct {
	ItemCode      string  `json:"item_code"`
	WhCode        string  `json:"wh_code"`
	CurrentQty    float64 `json:"current_qty"`
	AverageCost   float64 `json:"average_cost"`
	TotalValue    float64 `json:"total_value"`
	CostingMethod string  `json:"costing_method"`
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
	ItemCode              string  `json:"item_code"`
	CostingMethod         string  `json:"costing_method"`
	LotTrackingEnabled    bool    `json:"lot_tracking_enabled"`
	ExpiryTrackingEnabled bool    `json:"expiry_tracking_enabled"`
	ExpiryAlertDays       int     `json:"expiry_alert_days"`
	AutoBlockExpired      bool    `json:"auto_block_expired"`
	AllowNegativeStock    bool    `json:"allow_negative_stock"`
	CostByWarehouse       bool    `json:"cost_by_warehouse"`
	StandardCost          float64 `json:"standard_cost"`
}

// InventoryValuationReport — รายงานมูลค่าสินค้าคงเหลือ
type InventoryValuationReport struct {
	HoldingCode string                   `json:"holding_code"`
	AsOfDate    time.Time                `json:"as_of_date"`
	Items       []InventoryValuationItem `json:"items"`
	TotalValue  float64                  `json:"total_value"`
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
	HoldingCode string           `json:"holding_code"`
	ItemCode    string           `json:"item_code"`
	ItemName    string           `json:"item_name"`
	FromDate    time.Time        `json:"from_date"`
	ToDate      time.Time        `json:"to_date"`
	Entries     []StockCardEntry `json:"entries"`
}

type StockCardEntry struct {
	Date            time.Time `json:"date"`
	DocNo           string    `json:"doc_no"`
	TransactionType string    `json:"transaction_type"`
	TransFlag       int       `json:"trans_flag"`
	QtyIn           float64   `json:"qty_in"`
	QtyOut          float64   `json:"qty_out"`
	UnitCost        float64   `json:"unit_cost"`
	TotalCost       float64   `json:"total_cost"`
	BalanceQty      float64   `json:"balance_qty"`
	BalanceValue    float64   `json:"balance_value"`
	AverageCost     float64   `json:"average_cost"`
}
