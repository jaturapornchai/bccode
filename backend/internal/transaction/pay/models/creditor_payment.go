package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const payCollectionName = "transactionPay"

type Pay struct {
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
	Details                  *[]PayDetail              `json:"details" bson:"details"`
	PaymentDetail            transmodels.PaymentDetail `json:"paymentdetail" bson:"paymentdetail"`
	PaymentDetailRaw         string                    `json:"paymentdetailraw" bson:"paymentdetailraw"`

	PayCashAmount    float64 `json:"paycashamount" bson:"paycashamount"`
	PayCashChange    float64 `json:"paycashchange" bson:"paycashchange"`
	SumQrCode        float64 `json:"sumqrcode" bson:"sumqrcode"`               // ชำระเงินโดย QR Code
	SumCreditCard    float64 `json:"sumcreditcard" bson:"sumcreditcard"`       // ชำระเงินโดย Credit Card
	SumMoneyTransfer float64 `json:"summoneytransfer" bson:"summoneytransfer"` // ชำระเงินโดยเงินโอน
	SumCheque        float64 `json:"sumcheque" bson:"sumcheque"`               // ชำระเงินโดยเช็ค
	SumCoupon        float64 `json:"sumcoupon" bson:"sumcoupon"`               // ชำระเงินโดย Coupon
	SumCredit        float64 `json:"sumcredit" bson:"sumcredit"`
	RoundAmount      float64 `json:"roundamount" bson:"roundamount"`
	IsCancel         bool    `json:"iscancel" bson:"iscancel"`

	Branch PayBranch `json:"branch" bson:"branch"`
}

type PayBranch struct {
	models.DocIdentity `bson:"inline"`
	Code               string          `json:"code" bson:"code"`
	Names              *[]models.NameX `json:"names" bson:"names"`
}

type PayDetail struct {
	Selected      bool      `json:"selected" bson:"selected"`
	DocNo         string    `json:"docno" bson:"docno"`
	DocDatetime   time.Time `json:"docdatetime" bson:"docdatetime"`
	TransFlag     int8      `json:"transflag" bson:"transflag"`
	Value         float64   `json:"value" bson:"value"`
	Balance       float64   `json:"balance" bson:"balance"`
	PaymentAmount float64   `json:"paymentamount" bson:"paymentamount"`
}

type PayInfo struct {
	models.DocIdentity `bson:"inline"`
	Pay                `bson:"inline"`
}

func (PayInfo) CollectionName() string {
	return payCollectionName
}

type PayData struct {
	models.ShopIdentity `bson:"inline"`
	PayInfo             `bson:"inline"`
}

type PayDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	PayData            `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (PayDoc) CollectionName() string {
	return payCollectionName
}

type PayItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (PayItemGuid) CollectionName() string {
	return payCollectionName
}

type PayActivity struct {
	PayData             `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (PayActivity) CollectionName() string {
	return payCollectionName
}

type PayDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (PayDeleteActivity) CollectionName() string {
	return payCollectionName
}
