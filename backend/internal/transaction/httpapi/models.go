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
	"paid":                  28,  // RE ใบเสร็จรับเงิน/รับชำระหนี้
	"pay":                   29,  // PV ใบสำคัญจ่าย/จ่ายชำระหนี้
	"receipt":               28,  // RE ใบเสร็จรับเงิน
	"payment":               29,  // PV ใบสำคัญจ่าย
	"purchasereturn":        16,  // PRN ใบส่งคืนสินค้าจากการซื้อ
	"salereturn":            48,  // SRN ใบรับคืนสินค้าจากการขาย
	"creditnote":            48,  // CN ใบลดหนี้
	"debitnote":             50,  // DN ใบเพิ่มหนี้
	"expense":               24,  // EXP บันทึกค่าใช้จ่าย
	"billingnote":           26,  // BL ใบวางบิล
	"apbillingreceipt":      27,  // BAP ใบรับวางบิล
	"deposit":               22,  // DEP ใบจ่ายเงินมัดจำ
	"depositrefund":         23,  // DPR ใบรับคืนเงินมัดจำ
	"paidadvance":           30,  // ADV ใบจ่ายเงินล่วงหน้า
	"advancepayment":        30,  // ADV ใบจ่ายเงินล่วงหน้า
	"paidadvancerefund":     31,  // ADR ใบรับคืนเงินจ่ายล่วงหน้า
	"advancepaymentrefund":  31,  // ADR ใบรับคืนเงินจ่ายล่วงหน้า
	"receivedeposit":        32,  // RDEP ใบรับเงินมัดจำ
	"receivedepositrefund":  33,  // RDR ใบจ่ายคืนเงินมัดจำ
	"receiveadvance":        34,  // RADV ใบรับเงินล่วงหน้า
	"receiveadvancerefund":  35,  // RADR ใบจ่ายคืนเงินรับล่วงหน้า
	"receivableother":       54,  // AR-OTH ลูกหนี้อื่น
	"debtorbeginningbalance": 52, // AR-BAL ลูกหนี้ยกมา
	"creditorbeginningbalance": 53, // AP-BAL เจ้าหนี้ยกมา
	"bank":                  60,  // BNK ธนาคาร/เงินสด
	"chequereceive":         61,  // CHQ-IN เช็ครับ
	"chequepayment":         62,  // CHQ-OUT เช็คจ่าย
	"chequereceived":        61,  // CHQ-IN เช็ครับ
	"chequeissued":          62,  // CHQ-OUT เช็คจ่าย
	"temporaryreceipt":      28,  // TRC ใบเสร็จรับเงินชั่วคราว
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
	"paid":                  "RE",
	"pay":                   "PV",
	"receipt":               "RE",
	"payment":               "PV",
	"purchasereturn":        "PRN",
	"salereturn":            "SRN",
	"creditnote":            "CN",
	"debitnote":             "DN",
	"expense":               "EXP",
	"billingnote":           "BL",
	"apbillingreceipt":      "BAP",
	"deposit":               "DEP",
	"depositrefund":         "DPR",
	"paidadvance":           "ADV",
	"advancepayment":        "ADV",
	"paidadvancerefund":     "ADR",
	"advancepaymentrefund":  "ADR",
	"receivedeposit":        "RDEP",
	"receivedepositrefund":  "RDR",
	"receiveadvance":        "RADV",
	"receiveadvancerefund":  "RADR",
	"receivableother":       "AR-OTH",
	"debtorbeginningbalance": "AR-BAL",
	"creditorbeginningbalance": "AP-BAL",
	"bank":                  "BNK",
	"chequereceive":         "CHQ-IN",
	"chequepayment":         "CHQ-OUT",
	"chequereceived":        "CHQ-IN",
	"chequeissued":          "CHQ-OUT",
	"temporaryreceipt":      "TRC",
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
	DocRef         string  `json:"docref,omitempty"`
}

type ErpTransactionDoc struct {
	ID             string          `json:"id,omitempty"`
	DocNo          string          `json:"docno"`
	DocDateTime    string          `json:"docdatetime"`
	TransFlag      int             `json:"transflag"`
	CustCode       string          `json:"custcode,omitempty"`
	CustName       string          `json:"custname,omitempty"`
	DocRef         string          `json:"docref,omitempty"`
	DocRefTransFlag int            `json:"docreftransflag,omitempty"`
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
	PostToGL       bool            `json:"post_to_gl,omitempty"`
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
