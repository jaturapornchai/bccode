package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const rfqCollectionName = "transactionRequestForQuotation"

// VendorItem — รายการสินค้าของ vendor แต่ละราย
type VendorItem struct {
	LineNumber int             `json:"linenumber" bson:"linenumber"`
	ItemCode   string          `json:"itemcode" bson:"itemcode"`
	ItemNames  *[]models.NameX `json:"itemnames" bson:"itemnames"`
	UnitCode   string          `json:"unitcode" bson:"unitcode"`
	Qty        float64         `json:"qty" bson:"qty"`
	Price      float64         `json:"price" bson:"price"`
	SumAmount  float64         `json:"sumamount" bson:"sumamount"`
	Remark     string          `json:"remark,omitempty" bson:"remark,omitempty"`
}

// VendorEntry — ข้อมูล vendor แต่ละรายที่เสนอราคา
type VendorEntry struct {
	VendorCode     string          `json:"vendorcode" bson:"vendorcode"`
	VendorNames    *[]models.NameX `json:"vendornames" bson:"vendornames"`
	QuotationDocNo string          `json:"quotationdocno,omitempty" bson:"quotationdocno,omitempty"`
	QuotationDate  string          `json:"quotationdate,omitempty" bson:"quotationdate,omitempty"`
	CreditDays     int             `json:"creditdays" bson:"creditdays"`
	DeliveryDays   int             `json:"deliverydays" bson:"deliverydays"`
	DeliveryTerms  string          `json:"deliveryterms,omitempty" bson:"deliveryterms,omitempty"`
	QualityNotes   string          `json:"qualitynotes,omitempty" bson:"qualitynotes,omitempty"`
	TotalAmount    float64         `json:"totalamount" bson:"totalamount"`
	IsSelected     bool            `json:"isselected" bson:"isselected"`
	Items          []VendorItem    `json:"items" bson:"items"`
}

// RFQ — ใบสืบราคา (Request for Quotation)
type RFQ struct {
	models.PartitionIdentity `bson:"inline"`
	transmodels.Transaction  `bson:"inline"`

	// === RFQ-specific fields ===
	RefPRDocNo         string        `json:"refprdocno,omitempty" bson:"refprdocno,omitempty"`
	RefPRGuidFixed     string        `json:"refprguidfixed,omitempty" bson:"refprguidfixed,omitempty"`
	VendorEntries      []VendorEntry `json:"vendorentries" bson:"vendorentries"`
	SelectedVendor     string        `json:"selectedvendor,omitempty" bson:"selectedvendor,omitempty"`
	SelectionReason    string        `json:"selectionreason,omitempty" bson:"selectionreason,omitempty"`
	RefPODocNo         string        `json:"refpodocno,omitempty" bson:"refpodocno,omitempty"`
	RefPOGuidFixed     string        `json:"refpoguidfixed,omitempty" bson:"refpoguidfixed,omitempty"`
	SubmissionDeadline string        `json:"submissiondeadline,omitempty" bson:"submissiondeadline,omitempty"`
	MinVendors         int8          `json:"minvendors" bson:"minvendors"`
	ConversionStatus   string        `json:"conversionstatus,omitempty" bson:"conversionstatus,omitempty"`
}

type RFQInfo struct {
	models.DocIdentity `bson:"inline"`
	RFQ                `bson:"inline"`
}

func (RFQInfo) CollectionName() string {
	return rfqCollectionName
}

type RFQData struct {
	models.HoldingCodeentity `bson:"inline"`
	RFQInfo                  `bson:"inline"`
}

type RFQDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"id,omitempty"`
	RFQData            `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (RFQDoc) CollectionName() string {
	return rfqCollectionName
}

type RFQItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (RFQItemGuid) CollectionName() string {
	return rfqCollectionName
}

type RFQActivity struct {
	RFQData             `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (RFQActivity) CollectionName() string {
	return rfqCollectionName
}

type RFQDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (RFQDeleteActivity) CollectionName() string {
	return rfqCollectionName
}
