package models

import (
	"time"
)

// === Transaction Type Constants ===

const (
	TxTypePurchaseReceipt = "purchasereceipt"
	TxTypePurchaseReturn  = "purchasereturn"
	TxTypeSalesIssue      = "salesissue"
	TxTypeSalesReturn     = "salesreturn"
	TxTypeTransferOut     = "transferout"
	TxTypeTransferIn      = "transferin"
	TxTypeAdjustmentIn    = "adjustmentin"
	TxTypeAdjustmentOut   = "adjustmentout"
	TxTypeWriteOff        = "writeoff"
	TxTypeAssemblyIssue   = "assemblyissue"
	TxTypeAssemblyReceipt = "assemblyreceipt"
	TxTypeCostAdjustment  = "costadjustment"
)

// === Period Status Constants ===

const (
	PeriodStatusOpen   = "open"
	PeriodStatusClosed = "closed"
	PeriodStatusLocked = "locked"
)

// === Allocation Method Constants ===

const (
	AllocByValue    = "byvalue"
	AllocByQuantity = "byquantity"
	AllocByWeight   = "byweight"
	AllocByVolume   = "byvolume"
	AllocManual     = "manual"
)

// === Core Models ===

// InventoryStockBalance — ยอดคงเหลือต่อสินค้า × คลัง (สรุป snapshot)
type InventoryStockBalance struct {
	ID                int64      `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	HoldingCode       string     `json:"holdingcode" gorm:"column:holdingcode"`
	ItemCode          string     `json:"itemcode" gorm:"column:itemcode"`
	Barcode           string     `json:"barcode" gorm:"column:barcode"`
	WhCode            string     `json:"whcode" gorm:"column:whcode"`
	LocationCode      string     `json:"locationcode" gorm:"column:locationcode"`
	CurrentQty        float64    `json:"currentqty" gorm:"column:currentqty"`
	ReservedQty       float64    `json:"reservedqty" gorm:"column:reservedqty"`
	CurrentAvgCost    float64    `json:"currentavgcost" gorm:"column:currentavgcost"`
	CurrentTotalValue float64    `json:"currenttotalvalue" gorm:"column:currenttotalvalue"`
	LastPurchaseCost  float64    `json:"lastpurchasecost" gorm:"column:lastpurchasecost"`
	LastPurchaseDate  *time.Time `json:"lastpurchasedate" gorm:"column:lastpurchasedate"`
	UpdatedAt         time.Time  `json:"updatedat" gorm:"column:updatedat"`
}

func (InventoryStockBalance) TableName() string { return "inventorystockbalances" }

// MarketplaceStockBalance stores marketplace availability by product dimensions.
// This projection is not an accounting cost source.
type MarketplaceStockBalance struct {
	ID              int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	HoldingCode     string    `json:"holdingcode" gorm:"column:holdingcode"`
	ItemCode        string    `json:"itemcode" gorm:"column:itemcode"`
	DimensionKey    string    `json:"dimensionkey" gorm:"column:dimensionkey"`
	DimensionValues string    `json:"dimensionvalues" gorm:"column:dimensionvalues"`
	CurrentQty      float64   `json:"currentqty" gorm:"column:currentqty"`
	ReservedQty     float64   `json:"reservedqty" gorm:"column:reservedqty"`
	AvailableQty    float64   `json:"availableqty" gorm:"column:availableqty"`
	Source          string    `json:"source" gorm:"column:source"`
	UpdatedAt       time.Time `json:"updatedat" gorm:"column:updatedat"`
}

func (MarketplaceStockBalance) TableName() string { return "marketplacestockbalances" }

// MarketplaceDimensionPrice stores selling prices by product dimension and marketplace.
type MarketplaceDimensionPrice struct {
	ID              int64      `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	HoldingCode     string     `json:"holdingcode" gorm:"column:holdingcode"`
	ItemCode        string     `json:"itemcode" gorm:"column:itemcode"`
	Barcode         string     `json:"barcode" gorm:"column:barcode"`
	DimensionKey    string     `json:"dimensionkey" gorm:"column:dimensionkey"`
	DimensionValues string     `json:"dimensionvalues" gorm:"column:dimensionvalues"`
	PriceLevel      string     `json:"pricelevel" gorm:"column:pricelevel"`
	Marketplace     string     `json:"marketplace" gorm:"column:marketplace"`
	Currency        string     `json:"currency" gorm:"column:currency"`
	Price           float64    `json:"price" gorm:"column:price"`
	SalePrice       float64    `json:"saleprice" gorm:"column:saleprice"`
	CompareAtPrice  float64    `json:"compareatprice" gorm:"column:compareatprice"`
	EffectiveFrom   *time.Time `json:"effectivefrom" gorm:"column:effectivefrom"`
	EffectiveTo     *time.Time `json:"effectiveto" gorm:"column:effectiveto"`
	IsActive        bool       `json:"isactive" gorm:"column:isactive"`
	UpdatedAt       time.Time  `json:"updatedat" gorm:"column:updatedat"`
}

func (MarketplaceDimensionPrice) TableName() string { return "marketplacedimensionprices" }

// InventoryCostTransaction — ประวัติทุก transaction ที่กระทบต้นทุน
type InventoryCostTransaction struct {
	ID                int64      `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	HoldingCode       string     `json:"holdingcode" gorm:"column:holdingcode"`
	ItemCode          string     `json:"itemcode" gorm:"column:itemcode"`
	Barcode           string     `json:"barcode" gorm:"column:barcode"`
	WhCode            string     `json:"whcode" gorm:"column:whcode"`
	LocationCode      string     `json:"locationcode" gorm:"column:locationcode"`
	TransactionType   string     `json:"transactiontype" gorm:"column:transactiontype"`
	TransFlag         int        `json:"transflag" gorm:"column:transflag"`
	RefDocType        string     `json:"refdoctype" gorm:"column:refdoctype"`
	RefDocNo          string     `json:"refdocno" gorm:"column:refdocno"`
	Qty               float64    `json:"qty" gorm:"column:qty"`
	UnitCost          float64    `json:"unitcost" gorm:"column:unitcost"`
	TotalCost         float64    `json:"totalcost" gorm:"column:totalcost"`
	LandedCost        float64    `json:"landedcost" gorm:"column:landedcost"`
	LotNumber         string     `json:"lotnumber" gorm:"column:lotnumber"`
	ExpiryDate        *time.Time `json:"expirydate" gorm:"column:expirydate"`
	BalanceQty        float64    `json:"balanceqty" gorm:"column:balanceqty"`
	BalanceAvgCost    float64    `json:"balanceavgcost" gorm:"column:balanceavgcost"`
	BalanceTotalValue float64    `json:"balancetotalvalue" gorm:"column:balancetotalvalue"`
	TransactionDate   time.Time  `json:"transactiondate" gorm:"column:transactiondate"`
	AccountingPeriod  string     `json:"accountingperiod" gorm:"column:accountingperiod"`
	CreatedBy         string     `json:"createdby" gorm:"column:createdby"`
	CreatedAt         time.Time  `json:"createdat" gorm:"column:createdat"`
	IsReversed        bool       `json:"isreversed" gorm:"column:isreversed"`
	ReversedByID      *int64     `json:"reversedbyid" gorm:"column:reversedbyid"`
}

func (InventoryCostTransaction) TableName() string { return "inventorycosttransactions" }

// InventoryLandedCost — ค่าใช้จ่ายประกอบ (ค่าขนส่ง, ภาษี, ประกัน)
type InventoryLandedCost struct {
	ID               int64      `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	HoldingCode      string     `json:"holdingcode" gorm:"column:holdingcode"`
	RefDocType       string     `json:"refdoctype" gorm:"column:refdoctype"`
	RefDocNo         string     `json:"refdocno" gorm:"column:refdocno"`
	CostType         string     `json:"costtype" gorm:"column:costtype"`
	Description      string     `json:"description" gorm:"column:description"`
	TotalAmount      float64    `json:"totalamount" gorm:"column:totalamount"`
	Currency         string     `json:"currency" gorm:"column:currency"`
	ExchangeRate     float64    `json:"exchangerate" gorm:"column:exchangerate"`
	AmountTHB        float64    `json:"amountthb" gorm:"column:amountthb"`
	AllocationMethod string     `json:"allocationmethod" gorm:"column:allocationmethod"`
	IsAllocated      bool       `json:"isallocated" gorm:"column:isallocated"`
	AllocatedAt      *time.Time `json:"allocatedat" gorm:"column:allocatedat"`
	CreatedAt        time.Time  `json:"createdat" gorm:"column:createdat"`
}

func (InventoryLandedCost) TableName() string { return "inventorylandedcosts" }

// InventoryLandedCostAllocation — การกระจาย landed cost ไปแต่ละสินค้า
type InventoryLandedCostAllocation struct {
	ID               int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	HoldingCode      string    `json:"holdingcode" gorm:"column:holdingcode"`
	LandedCostID     int64     `json:"landedcostid" gorm:"column:landedcostid"`
	ItemCode         string    `json:"itemcode" gorm:"column:itemcode"`
	WhCode           string    `json:"whcode" gorm:"column:whcode"`
	AllocatedAmount  float64   `json:"allocatedamount" gorm:"column:allocatedamount"`
	AllocatedPerUnit float64   `json:"allocatedperunit" gorm:"column:allocatedperunit"`
	Qty              float64   `json:"qty" gorm:"column:qty"`
	CreatedAt        time.Time `json:"createdat" gorm:"column:createdat"`
}

func (InventoryLandedCostAllocation) TableName() string { return "inventorylandedcostallocations" }

// InventoryAccountingPeriod — งวดบัญชี
type InventoryAccountingPeriod struct {
	ID                    int64      `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	HoldingCode           string     `json:"holdingcode" gorm:"column:holdingcode"`
	PeriodCode            string     `json:"periodcode" gorm:"column:periodcode"`
	StartDate             time.Time  `json:"startdate" gorm:"column:startdate"`
	EndDate               time.Time  `json:"enddate" gorm:"column:enddate"`
	Status                string     `json:"status" gorm:"column:status"`
	ClosedBy              string     `json:"closedby" gorm:"column:closedby"`
	ClosedAt              *time.Time `json:"closedat" gorm:"column:closedat"`
	CreatedAt             time.Time  `json:"createdat" gorm:"column:createdat"`
}

func (InventoryAccountingPeriod) TableName() string { return "inventoryaccountingperiods" }

// === Request/Response Structs ===

// ReceiptParams — parameter สำหรับรับสินค้าเข้า
type ReceiptParams struct {
	HoldingCode  string     `json:"holdingcode"`
	ItemCode     string     `json:"itemcode"`
	Barcode      string     `json:"barcode"`
	WhCode       string     `json:"whcode"`
	LocationCode string     `json:"locationcode"`
	Qty          float64    `json:"qty"`
	UnitCost     float64    `json:"unitcost"`
	RefDocType   string     `json:"refdoctype"`
	RefDocNo     string     `json:"refdocno"`
	TransFlag    int        `json:"transflag"`
	LotNumber    string     `json:"lotnumber,omitempty"`
	ExpiryDate   *time.Time `json:"expirydate,omitempty"`
	ReceivedDate time.Time  `json:"receiveddate"`
	CreatedBy    string     `json:"createdby"`
}

// IssueParams — parameter สำหรับตัดสินค้าออก (ขาย/เบิก)
type IssueParams struct {
	HoldingCode     string    `json:"holdingcode"`
	ItemCode        string    `json:"itemcode"`
	Barcode         string    `json:"barcode"`
	WhCode          string    `json:"whcode"`
	LocationCode    string    `json:"locationcode"`
	Qty             float64   `json:"qty"`
	RefDocType      string    `json:"refdoctype"`
	RefDocNo        string    `json:"refdocno"`
	TransFlag       int       `json:"transflag"`
	LotNumber       string    `json:"lotnumber,omitempty"`
	TransactionDate time.Time `json:"transactiondate"`
	CreatedBy       string    `json:"createdby"`
}

// TransferParams — parameter สำหรับโอนย้ายคลัง
type TransferParams struct {
	HoldingCode      string    `json:"holdingcode"`
	ItemCode         string    `json:"itemcode"`
	Barcode          string    `json:"barcode"`
	FromWhCode       string    `json:"fromwhcode"`
	FromLocationCode string    `json:"fromlocationcode"`
	ToWhCode         string    `json:"towhcode"`
	ToLocationCode   string    `json:"tolocationcode"`
	Qty              float64   `json:"qty"`
	RefDocType       string    `json:"refdoctype"`
	RefDocNo         string    `json:"refdocno"`
	TransFlag        int       `json:"transflag"`
	LotNumber        string    `json:"lotnumber,omitempty"`
	TransactionDate  time.Time `json:"transactiondate"`
	CreatedBy        string    `json:"createdby"`
}

// AdjustmentParams — parameter สำหรับปรับปรุง stock
type AdjustmentParams struct {
	HoldingCode     string     `json:"holdingcode"`
	ItemCode        string     `json:"itemcode"`
	Barcode         string     `json:"barcode"`
	WhCode          string     `json:"whcode"`
	LocationCode    string     `json:"locationcode"`
	Qty             float64    `json:"qty"`
	UnitCost        float64    `json:"unitcost"`
	IsIncrease      bool       `json:"isincrease"`
	RefDocType      string     `json:"refdoctype"`
	RefDocNo        string     `json:"refdocno"`
	TransFlag       int        `json:"transflag"`
	LotNumber       string     `json:"lotnumber,omitempty"`
	ExpiryDate      *time.Time `json:"expirydate,omitempty"`
	TransactionDate time.Time  `json:"transactiondate"`
	CreatedBy       string     `json:"createdby"`
}

// SalesReturnParams — parameter สำหรับรับคืนจากลูกค้า
type SalesReturnParams struct {
	HoldingCode     string    `json:"holdingcode"`
	ItemCode        string    `json:"itemcode"`
	Barcode         string    `json:"barcode"`
	WhCode          string    `json:"whcode"`
	LocationCode    string    `json:"locationcode"`
	Qty             float64   `json:"qty"`
	OriginalCost    float64   `json:"originalcost"`
	RefDocType      string    `json:"refdoctype"`
	RefDocNo        string    `json:"refdocno"`
	TransFlag       int       `json:"transflag"`
	LotNumber       string    `json:"lotnumber,omitempty"`
	TransactionDate time.Time `json:"transactiondate"`
	CreatedBy       string    `json:"createdby"`
}

// PurchaseReturnParams — parameter สำหรับส่งคืนสินค้าให้ supplier
type PurchaseReturnParams struct {
	HoldingCode     string    `json:"holdingcode"`
	ItemCode        string    `json:"itemcode"`
	Barcode         string    `json:"barcode"`
	WhCode          string    `json:"whcode"`
	LocationCode    string    `json:"locationcode"`
	Qty             float64   `json:"qty"`
	RefDocType      string    `json:"refdoctype"`
	RefDocNo        string    `json:"refdocno"`
	TransFlag       int       `json:"transflag"`
	LotNumber       string    `json:"lotnumber,omitempty"`
	TransactionDate time.Time `json:"transactiondate"`
	CreatedBy       string    `json:"createdby"`
}

// StockValuation — ผลลัพธ์มูลค่าสินค้าคงเหลือ
type StockValuation struct {
	ItemCode      string  `json:"itemcode"`
	WhCode        string  `json:"whcode"`
	CurrentQty    float64 `json:"currentqty"`
	AverageCost   float64 `json:"averagecost"`
	TotalValue    float64 `json:"totalvalue"`
}

// CostTransactionResult — ผลลัพธ์หลังประมวลผล transaction
type CostTransactionResult struct {
	Transaction    *InventoryCostTransaction `json:"transaction"`
	BalanceQty     float64                   `json:"balanceqty"`
	BalanceAvgCost float64                   `json:"balanceavgcost"`
	TotalValue     float64                   `json:"totalvalue"`
}

// ProductCostingConfig — ตั้งค่าต้นทุนต่อสินค้า (ต้นทุนถัวเฉลี่ยวิธีเดียว: เลือกได้แค่ว่าคิดแยกคลังหรือไม่ และตัวเลือก lot/วันหมดอายุ)
type ProductCostingConfig struct {
	ItemCode              string  `json:"itemcode"`
	LotTrackingEnabled    bool    `json:"lottrackingenabled"`
	ExpiryTrackingEnabled bool    `json:"expirytrackingenabled"`
	ExpiryAlertDays       int     `json:"expiryalertdays"`
	AutoBlockExpired      bool    `json:"autoblockexpired"`
	AllowNegativeStock    bool    `json:"allownegativestock"`
	CostByWarehouse       bool    `json:"costbywarehouse"`
}

// InventoryValuationReport — รายงานมูลค่าสินค้าคงเหลือ
type InventoryValuationReport struct {
	HoldingCode string                   `json:"holdingcode"`
	AsOfDate    time.Time                `json:"asofdate"`
	Items       []InventoryValuationItem `json:"items"`
	TotalValue  float64                  `json:"totalvalue"`
}

type InventoryValuationItem struct {
	ItemCode      string  `json:"itemcode"`
	ItemName      string  `json:"itemname"`
	WhCode        string  `json:"whcode"`
	UnitCode      string  `json:"unitcode"`
	Qty           float64 `json:"qty"`
	AverageCost   float64 `json:"averagecost"`
	TotalValue    float64 `json:"totalvalue"`
}

// StockCardReport — รายงาน Stock Card (รายการเคลื่อนไหวต่อสินค้า)
type StockCardReport struct {
	HoldingCode string           `json:"holdingcode"`
	ItemCode    string           `json:"itemcode"`
	ItemName    string           `json:"itemname"`
	FromDate    time.Time        `json:"fromdate"`
	ToDate      time.Time        `json:"todate"`
	Entries     []StockCardEntry `json:"entries"`
}

type StockCardEntry struct {
	Date            time.Time `json:"date"`
	DocNo           string    `json:"docno"`
	TransactionType string    `json:"transactiontype"`
	TransFlag       int       `json:"transflag"`
	QtyIn           float64   `json:"qtyin"`
	QtyOut          float64   `json:"qtyout"`
	UnitCost        float64   `json:"unitcost"`
	TotalCost       float64   `json:"totalcost"`
	BalanceQty      float64   `json:"balanceqty"`
	BalanceValue    float64   `json:"balancevalue"`
	AverageCost     float64   `json:"averagecost"`
}
