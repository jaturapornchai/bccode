package httpapi

import "time"

// ModuleTransFlag maps ERP URL segment to Champ transaction flag.
var ModuleTransFlags = map[string]int{
	"purchase-requisition":  2,   // PR ใบขอซื้อ
	"purchase-order":        6,   // PO ใบสั่งซื้อ
	"purchase":              12,  // PUR ใบซื้อสินค้า
	"purchase-return":       16,  // PRET ใบส่งคืนสินค้า
	"stock-prickup-product": 20,  // OUT เบิกใช้สินค้า
	"stock-return-product":  21,  // RET รับคืนจากการเบิก
	"quotation":             34,  // QT ใบเสนอราคา
	"sale-order":            36,  // SO ใบสั่งขาย
	"sale-invoice":          44,  // INV ขายสินค้า/บริการ
	"sale-invoice-return":   48,  // SRET รับคืนจากการขาย
	"stock-balance":         56,  // BAL สินค้ายกมา
	"stock-transfer":        62,  // XFR โอนสินค้า
	"stock-receive-product": 72,  // RCV รับสินค้าสำเร็จรูป
	"stock-adjustment":      310, // ADJ ปรับปรุงสต็อก
}

// ModulePrefixes maps ERP module to default document number prefix.
var ModulePrefixes = map[string]string{
	"purchase-requisition":  "PR",
	"purchase-order":        "PO",
	"purchase":              "PUR",
	"purchase-return":       "PRET",
	"stock-prickup-product": "OUT",
	"stock-return-product":  "RET",
	"quotation":             "QT",
	"sale-order":            "SO",
	"sale-invoice":          "INV",
	"sale-invoice-return":   "SR",
	"stock-balance":         "BAL",
	"stock-transfer":        "XFR",
	"stock-receive-product": "RCV",
	"stock-adjustment":      "ADJ",
}

type ErpDetailItem struct {
	ID             int     `json:"id,omitempty"`
	LineNumber     int     `json:"linenumber"`
	ItemCode       string  `json:"itemcode"`
	ItemName       string  `json:"itemname,omitempty"`
	Barcode        string  `json:"barcode,omitempty"`
	UnitCode       string  `json:"unitcode,omitempty"`
	UnitName       string  `json:"unitname,omitempty"`
	Qty            float64 `json:"qty"`
	Price          float64 `json:"price"`
	Discount       string  `json:"discount,omitempty"`
	DiscountAmount float64 `json:"discountamount,omitempty"`
	SumAmount      float64 `json:"sumamount"`
	VatCal         int     `json:"vatcal,omitempty"`
	WhCode         string  `json:"whcode,omitempty"`
	LocationCode   string  `json:"locationcode,omitempty"`
	Description    string  `json:"description,omitempty"`
}

type ErpTransactionDoc struct {
	ID             string          `json:"id,omitempty"`
	DocNo          string          `json:"docno"`
	DocDateTime    string          `json:"docdatetime"`
	TransFlag      int             `json:"transflag"`
	CustCode       string          `json:"custcode,omitempty"`
	CustName       string          `json:"custname,omitempty"`
	Description    string          `json:"description,omitempty"`
	TotalAmount    float64         `json:"totalamount"`
	TotalValue     float64         `json:"totalvalue,omitempty"`
	TotalDiscount  float64         `json:"totaldiscount,omitempty"`
	TotalVatValue  float64         `json:"totalvatvalue,omitempty"`
	TotalBeforeVat float64         `json:"totalbeforevat,omitempty"`
	TotalAfterVat  float64         `json:"totalaftervat,omitempty"`
	VatType        int             `json:"vattype,omitempty"`
	VatRate        float64         `json:"vatrate,omitempty"`
	Status         int             `json:"status,omitempty"` // 0=Draft, 1=Approved, 2=Cancelled
	IsClosed       bool            `json:"isclosed,omitempty"`
	IsCancel       bool            `json:"iscancel,omitempty"`
	BranchCode     string          `json:"branchcode,omitempty"`
	Remark         string          `json:"remark,omitempty"`
	Details        []ErpDetailItem `json:"details"`
	CreatedAt      *time.Time      `json:"created_at,omitempty"`
	ModifiedAt     *time.Time      `json:"modified_at,omitempty"`
}

type DocListResponse struct {
	Success    bool                `json:"success"`
	Data       []ErpTransactionDoc `json:"data"`
	Pagination Pagination          `json:"pagination"`
}

type DocItemResponse struct {
	Success bool              `json:"success"`
	Data    ErpTransactionDoc `json:"data"`
	Message string            `json:"message,omitempty"`
}

type Pagination struct {
	Total  int `json:"total"`
	Page   int `json:"page"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}
