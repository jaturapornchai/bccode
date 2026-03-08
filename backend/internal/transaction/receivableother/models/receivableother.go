package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const collectionName = "transactionReceivableOthers"

type ReceivableOther struct {
	models.PartitionIdentity `bson:"inline"`
	DocNo                    string                    `json:"docno" bson:"docno"`
	DocDatetime              time.Time                 `json:"docdatetime" bson:"docdatetime"`
	DocType                  int8                      `json:"doctype" bson:"doctype"`
	TransFlag                int8                      `json:"transflag" bson:"transflag"`
	CustCode                 string                    `json:"custcode" bson:"custcode"`
	CustNames                *[]models.NameX           `json:"custnames" bson:"custnames"`
	SaleCode                 string                    `json:"salecode" bson:"salecode"`
	SaleName                 string                    `json:"salename" bson:"salename"`
	TotalPaymentAmount       float64                   `json:"totalpaymentamount" bson:"totalpaymentamount"`
	TotalAmount              float64                   `json:"totalamount" bson:"totalamount"`
	TotalBalance             float64                   `json:"totalbalance" bson:"totalbalance"`
	TotalValue               float64                   `json:"totalvalue" bson:"totalvalue"`
	Details                  *[]ReceivableOtherDetail  `json:"details" bson:"details"`
	PaymentDetail            transmodels.PaymentDetail `json:"paymentdetail" bson:"paymentdetail"`
	PaymentDetailRaw         string                    `json:"paymentdetailraw" bson:"paymentdetailraw"`
	RefDocNo                 string                    `json:"refdocno" bson:"refdocno"`     // เลขที่เอกสารอ้างอิง
	RefDocDate               time.Time                 `json:"refdocdate" bson:"refdocdate"` // วันที่เอกสารอ้างอิง

	PayCashAmount    float64 `json:"paycashamount" bson:"paycashamount"`
	SumQrCode        float64 `json:"sumqrcode" bson:"sumqrcode"`               // ชำระเงินโดย QR Code
	SumCreditCard    float64 `json:"sumcreditcard" bson:"sumcreditcard"`       // ชำระเงินโดย Credit Card
	SumMoneyTransfer float64 `json:"summoneytransfer" bson:"summoneytransfer"` // ชำระเงินโดยเงินโอน
	SumCheque        float64 `json:"sumcheque" bson:"sumcheque"`               // ชำระเงินโดยเช็ค
	SumCoupon        float64 `json:"sumcoupon" bson:"sumcoupon"`               // ชำระเงินโดย Coupon
	SumCredit        float64 `json:"sumcredit" bson:"sumcredit"`
	RoundAmount      float64 `json:"roundamount" bson:"roundamount"`
}

type ReceivableOtherDetail struct {
	Selected      bool      `json:"selected" bson:"selected"`
	DocNo         string    `json:"docno" bson:"docno"`
	DocDatetime   time.Time `json:"docdatetime" bson:"docdatetime"`
	TransFlag     int8      `json:"transflag" bson:"transflag"`
	Value         float64   `json:"value" bson:"value"`
	Balance       float64   `json:"balance" bson:"balance"`
	PaymentAmount float64   `json:"paymentamount" bson:"paymentamount"`
}

type ReceivableOtherInfo struct {
	models.DocIdentity `bson:"inline"`
	ReceivableOther    `bson:"inline"`
}

func (ReceivableOtherInfo) CollectionName() string {
	return collectionName
}

type ReceivableOtherData struct {
	models.ShopIdentity `bson:"inline"`
	ReceivableOtherInfo `bson:"inline"`
}

type ReceivableOtherDoc struct {
	ID                  primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ReceivableOtherData `bson:"inline"`
	models.ActivityDoc  `bson:"inline"`
}

func (ReceivableOtherDoc) CollectionName() string {
	return collectionName
}

type ReceivableOtherItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (ReceivableOtherItemGuid) CollectionName() string {
	return collectionName
}

type ReceivableOtherActivity struct {
	ReceivableOtherData `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ReceivableOtherActivity) CollectionName() string {
	return collectionName
}

type ReceivableOtherDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ReceivableOtherDeleteActivity) CollectionName() string {
	return collectionName
}
