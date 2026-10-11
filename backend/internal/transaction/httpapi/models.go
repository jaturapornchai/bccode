package httpapi

import (
	"time"

	"github.com/shopspring/decimal"
)

// ยอดเงิน/จำนวน/ราคา/ภาษีเป็น decimal.Decimal (Rule 20): JSON ตอบเป็นข้อความทศนิยม เช่น "107.5"; คำขอรับได้ทั้งตัวเลขและข้อความ;
// สแกน numeric ของ PostgreSQL ตรงตัว ไม่ผ่าน float — omitzero ละช่องที่เป็นศูนย์เหมือน omitempty เดิม
type ErpDetailItem struct {
	ID             int             `json:"id,omitempty"`
	LineNumber     int             `json:"linenumber"`
	ItemCode       string          `json:"itemcode"`
	ItemName       string          `json:"itemname,omitempty"`
	Barcode        string          `json:"barcode,omitempty"`
	UnitCode       string          `json:"unitcode,omitempty"`
	UnitName       string          `json:"unitname,omitempty"`
	Qty            decimal.Decimal `json:"qty"`
	Price          decimal.Decimal `json:"price"`
	Discount       string          `json:"discount,omitempty"`
	DiscountAmount decimal.Decimal `json:"discountamount,omitzero"`
	SumAmount      decimal.Decimal `json:"sumamount"`
	VatCal         int             `json:"vatcal,omitempty"`
	WhCode         string          `json:"whcode,omitempty"`
	LocationCode   string          `json:"locationcode,omitempty"`
	Description    string          `json:"description,omitempty"`
	DocRef         string          `json:"docref,omitempty"`
	// บรรทัดใบจ่ายชำระหนี้ที่ตัดหนี้รายใบ (Champ BCPaymentSub.PaybillNo/RefType): debt_document_id = แถวหนี้ในบัญชีย่อย,
	// paybillno = ใบเตรียมจ่ายที่ดึงมา (reftype 1); ว่าง/0 = ไม่อ้างอิง
	DebtDocumentID string `json:"debt_document_id,omitempty"`
	PayBillNo      string `json:"paybillno,omitempty"`
	RefType        int    `json:"reftype,omitempty"`
}

type ErpTransactionDoc struct {
	ID              string          `json:"id,omitempty"`
	DocNo           string          `json:"docno"`
	DocDateTime     string          `json:"docdatetime"`
	TransFlag       int             `json:"transflag"`
	CustCode        string          `json:"custcode,omitempty"`
	CustName        string          `json:"custname,omitempty"`
	DocRef          string          `json:"docref,omitempty"`
	DocRefTransFlag int             `json:"docreftransflag,omitempty"`
	Description     string          `json:"description,omitempty"`
	TotalAmount     decimal.Decimal `json:"totalamount"`
	TotalValue      decimal.Decimal `json:"totalvalue,omitzero"`
	TotalDiscount   decimal.Decimal `json:"totaldiscount,omitzero"`
	TotalVatValue   decimal.Decimal `json:"totalvatvalue,omitzero"`
	TotalBeforeVat  decimal.Decimal `json:"totalbeforevat,omitzero"`
	TotalAfterVat   decimal.Decimal `json:"totalaftervat,omitzero"`
	VatType         int             `json:"vattype,omitempty"`
	VatRate         decimal.Decimal `json:"vatrate,omitzero"`
	CreditDay       *int            `json:"creditday,omitempty"` // วันเครดิตเอกสารหนี้เจ้าหนี้ (Champ CreditDay); ไม่ส่ง = ของเจ้าหนี้ในทะเบียน
	DueDate         string          `json:"duedate,omitempty"`   // วันครบกำหนด YYYY-MM-DD (doc.perioddatetime, Champ DueDate)
	Status          int             `json:"status,omitempty"`    // 0=Draft, 1=Approved, 2=Cancelled
	IsClosed        bool            `json:"isclosed,omitempty"`
	IsCancel        bool            `json:"iscancel,omitempty"`
	ApprovalStatus  string          `json:"approvalstatus,omitempty"`
	ApprovedBy      string          `json:"approvedby,omitempty"`
	ApprovedAt      *time.Time      `json:"approvedat,omitempty"`
	CancelledBy     string          `json:"cancelledby,omitempty"`
	CancelledAt     *time.Time      `json:"cancelledat,omitempty"`
	CancelReason    string          `json:"cancelreason,omitempty"`
	CreatedBy       string          `json:"createdby,omitempty"`
	BranchCode      string          `json:"branchcode,omitempty"`
	Remark          string          `json:"remark,omitempty"`
	// GLStatus/GLDocNo answer a save: "posted" with the voucher number, "no_mapping" when the company has not set up
	// the posting pattern of this document type yet, "no_amount" for a zero-total document; empty = the type does not post.
	GLStatus   string          `json:"glstatus,omitempty"`
	GLDocNo    string          `json:"gldocno,omitempty"`
	Details    []ErpDetailItem `json:"details"`
	CreatedAt  *time.Time      `json:"created_at,omitempty"`
	ModifiedAt *time.Time      `json:"modified_at,omitempty"`
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

type ErpDocEvent struct {
	ID           int64     `json:"id"`
	BusinessCode string    `json:"businesscode"`
	TransFlag    int       `json:"transflag"`
	DocNo        string    `json:"docno"`
	Action       string    `json:"action"`
	FromStatus   string    `json:"from_status"`
	ToStatus     string    `json:"to_status"`
	Reason       string    `json:"reason"`
	Actor        string    `json:"actor"`
	CreatedAt    time.Time `json:"created_at"`
}

type DocEventsResponse struct {
	Success bool          `json:"success"`
	Data    []ErpDocEvent `json:"data"`
}

type BatchApproveItemResult struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type BatchApproveResponse struct {
	Success bool                     `json:"success"`
	Data    []BatchApproveItemResult `json:"data"`
}

type Pagination struct {
	Total  int `json:"total"`
	Page   int `json:"page"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}
