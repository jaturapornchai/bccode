package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const purchaserequisitionCollectionName = "transactionpurchaserequisition"

// PurchaseRequisition — ใบขอซื้อ (PR)
// embed Transaction base + PR-specific fields
type PurchaseRequisition struct {
	models.PartitionIdentity `bson:"inline"`
	transmodels.Transaction  `bson:"inline"`

	// === PR-specific fields ===
	RequesterCode         string          `json:"requestercode" bson:"requestercode"`                                     // รหัสผู้ขอซื้อ (พนักงาน)
	RequesterName         string          `json:"requestername" bson:"requestername"`                                     // ชื่อผู้ขอซื้อ
	DepartmentCode        string          `json:"departmentcode" bson:"departmentcode"`                                   // รหัสแผนก
	DepartmentNames       *[]models.NameX `json:"departmentnames" bson:"departmentnames"`                                 // ชื่อแผนก (multi-lang)
	Purpose               string          `json:"purpose" bson:"purpose"`                                                 // วัตถุประสงค์/เหตุผลในการขอซื้อ
	BudgetCode            string          `json:"budgetcode,omitempty" bson:"budgetcode,omitempty"`                       // รหัสงบประมาณ
	BudgetAmount          float64         `json:"budgetamount,omitempty" bson:"budgetamount,omitempty"`                   // วงเงินงบประมาณ
	Urgency               int8            `json:"urgency" bson:"urgency"`                                                 // 1=ปกติ 2=เร่งด่วน 3=เร่งด่วนมาก
	RequestedDeliveryDate string          `json:"requesteddeliverydate,omitempty" bson:"requesteddeliverydate,omitempty"` // วันที่ต้องการรับสินค้า
	RefPODocNo            string          `json:"refpodocno,omitempty" bson:"refpodocno,omitempty"`                       // เลขที่ PO ที่สร้างจาก PR
	RefPOGuidFixed        string          `json:"refpoguidfixed,omitempty" bson:"refpoguidfixed,omitempty"`
	RefRFQDocNo           string          `json:"refrfqdocno,omitempty" bson:"refrfqdocno,omitempty"` // เลขที่ RFQ ที่สร้างจาก PR
	RefRFQGuidFixed       string          `json:"refrfqguidfixed,omitempty" bson:"refrfqguidfixed,omitempty"`
	ConversionStatus      string          `json:"conversionstatus,omitempty" bson:"conversionstatus,omitempty"` // none, partial, full

	// === PR เพิ่มเติม ===
	CreditDays       int             `json:"creditdays,omitempty" bson:"creditdays,omitempty"`             // เครดิตเทอม (วัน)
	PaymentCondition string          `json:"paymentcondition,omitempty" bson:"paymentcondition,omitempty"` // เงื่อนไขชำระเงิน
	JobCode          string          `json:"jobcode,omitempty" bson:"jobcode,omitempty"`                   // รหัสงาน/โครงการ
	JobNames         *[]models.NameX `json:"jobnames,omitempty" bson:"jobnames,omitempty"`                 // ชื่องาน/โครงการ (multi-lang)
	ShipToAddress    string          `json:"shiptoaddress,omitempty" bson:"shiptoaddress,omitempty"`       // ที่อยู่จัดส่ง
	ShipToName       string          `json:"shiptoname,omitempty" bson:"shiptoname,omitempty"`             // ชื่อผู้รับ
	ShipToLat        string          `json:"shiptolat,omitempty" bson:"shiptolat,omitempty"`               // ละติจูดที่อยู่จัดส่ง
	ShipToLng        string          `json:"shiptolng,omitempty" bson:"shiptolng,omitempty"`               // ลองจิจูดที่อยู่จัดส่ง
	DocRefNo         string          `json:"docsrefno,omitempty" bson:"docsrefno,omitempty"`               // เอกสารอ้างอิงภายนอก
	Attachments      []string        `json:"attachments,omitempty" bson:"attachments,omitempty"`           // รายการไฟล์แนบ (URIs)

	// === PR เพิ่มเติม (session 6) ===
	InternalNote           string  `json:"internalnote,omitempty" bson:"internalnote,omitempty"`                     // หมายเหตุภายใน (ไม่แสดงใน print)
	VendorNote             string  `json:"vendornote,omitempty" bson:"vendornote,omitempty"`                         // หมายเหตุถึง vendor (แสดงใน PO)
	PurchasingGroup        string  `json:"purchasinggroup,omitempty" bson:"purchasinggroup,omitempty"`               // กลุ่มจัดซื้อ (เช่น วัตถุดิบ/IT/ก่อสร้าง)
	OverDeliveryTolerance  float64 `json:"overdeliverytolerance,omitempty" bson:"overdeliverytolerance,omitempty"`   // % ยอมรับรับเกิน
	UnderDeliveryTolerance float64 `json:"underdeliverytolerance,omitempty" bson:"underdeliverytolerance,omitempty"` // % ยอมรับรับขาด
}

type PurchaseRequisitionInfo struct {
	models.DocIdentity  `bson:"inline"`
	PurchaseRequisition `bson:"inline"`
}

func (PurchaseRequisitionInfo) CollectionName() string {
	return purchaserequisitionCollectionName
}

type PurchaseRequisitionData struct {
	models.HoldingCodeentity `bson:"inline"`
	PurchaseRequisitionInfo  `bson:"inline"`
}

type PurchaseRequisitionDoc struct {
	ID                      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	PurchaseRequisitionData `bson:"inline"`
	models.ActivityDoc      `bson:"inline"`
}

func (PurchaseRequisitionDoc) CollectionName() string {
	return purchaserequisitionCollectionName
}

type PurchaseRequisitionItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (PurchaseRequisitionItemGuid) CollectionName() string {
	return purchaserequisitionCollectionName
}

type PurchaseRequisitionActivity struct {
	PurchaseRequisitionData `bson:"inline"`
	models.ActivityTime     `bson:"inline"`
}

func (PurchaseRequisitionActivity) CollectionName() string {
	return purchaserequisitionCollectionName
}

type PurchaseRequisitionDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (PurchaseRequisitionDeleteActivity) CollectionName() string {
	return purchaserequisitionCollectionName
}
