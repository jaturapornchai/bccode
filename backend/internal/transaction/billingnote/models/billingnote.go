package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const CollectionName = "transactionbillingnotes"

type BillingNote struct {
	models.PartitionIdentity `bson:"inline"`
	BusinessCode             string                    `json:"businesscode" bson:"businesscode"`
	DocNo                    string                    `json:"docno" bson:"docno"`
	DocDatetime              time.Time                 `json:"docdatetime" bson:"docdatetime"`
	DueDate                  time.Time                 `json:"duedate" bson:"duedate"`
	DocType                  int8                      `json:"doctype" bson:"doctype"`
	TransFlag                int8                      `json:"transflag" bson:"transflag"`
	CustCode                 string                    `json:"custcode" bson:"custcode"`
	CustNames                *[]models.NameX           `json:"custnames" bson:"custnames"`
	SaleCode                 string                    `json:"salecode" bson:"salecode"`
	SaleName                 string                    `json:"salename" bson:"salename"`
	TotalAmount              float64                   `json:"totalamount" bson:"totalamount"`
	TotalBalance             float64                   `json:"totalbalance" bson:"totalbalance"`
	TotalValue               float64                   `json:"totalvalue" bson:"totalvalue"`
	Description              string                    `json:"description" bson:"description"`
	Details                  *[]BillingNoteDetail      `json:"details" bson:"details"`
	PaymentDetail            transmodels.PaymentDetail `json:"paymentdetail" bson:"paymentdetail"`
	PaymentDetailRaw         string                    `json:"paymentdetailraw" bson:"paymentdetailraw"`
	RefDocNo                 string                    `json:"refdocno" bson:"refdocno"`
	RefDocDate               time.Time                 `json:"refdocdate" bson:"refdocdate"`
}

type BillingNoteDetail struct {
	Selected      bool      `json:"selected" bson:"selected"`
	DocNo         string    `json:"docno" bson:"docno"`
	DocDatetime   time.Time `json:"docdatetime" bson:"docdatetime"`
	DueDate       time.Time `json:"duedate" bson:"duedate"`
	TransFlag     int8      `json:"transflag" bson:"transflag"`
	DocType       int8      `json:"doctype" bson:"doctype"`
	Value         float64   `json:"value" bson:"value"`
	Balance       float64   `json:"balance" bson:"balance"`
	PaymentAmount float64   `json:"paymentamount" bson:"paymentamount"`
	BillBalance   float64   `json:"billbalance" bson:"billbalance"`
	LineNumber    int       `json:"linenumber" bson:"linenumber"`
	Remark        string    `json:"remark" bson:"remark"`
}

type BillingNoteInfo struct {
	models.DocIdentity `bson:"inline"`
	BillingNote        `bson:"inline"`
}

func (BillingNoteInfo) CollectionName() string {
	return CollectionName
}

type BillingNoteData struct {
	models.HoldingCodeentity `bson:"inline"`
	BillingNoteInfo          `bson:"inline"`
}

type BillingNoteDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	BillingNoteData    `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (BillingNoteDoc) CollectionName() string {
	return CollectionName
}

type BillingNoteItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (BillingNoteItemGuid) CollectionName() string {
	return CollectionName
}

type BillingNoteActivity struct {
	BillingNoteData     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (BillingNoteActivity) CollectionName() string {
	return CollectionName
}

type BillingNoteDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (BillingNoteDeleteActivity) CollectionName() string {
	return CollectionName
}
