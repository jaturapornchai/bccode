package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const qrpaymentCollectionName = "qrpayment"

type QrPayment struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string          `json:"code" bson:"code" validate:"required"`
	QrNames                  *[]models.NameX `json:"qrnames" bson:"qrnames" validate:"required,min=1,unique=Code,dive"`
	QrType                   int             `json:"qrtype" bson:"qrtype"`
	IsActive                 bool            `json:"isactive" bson:"isactive"`
	Logo                     string          `json:"logo" bson:"logo"`
	BankCode                 string          `json:"bankcode" bson:"bankcode"`
	BankNames                *[]models.NameX `json:"banknames" bson:"banknames"`
	BookBankCode             string          `json:"bookbankcode" bson:"bookbankcode"`
	BookBankNames            *[]models.NameX `json:"bookbanknames" bson:"bookbanknames"`

	QrCode         string `json:"qrcode" bson:"qrcode"`
	ApiKey         string `json:"apikey" bson:"apikey"`
	BillerCode     string `json:"billercode" bson:"billercode"`
	BillerID       string `json:"billerid" bson:"billerid"`
	StoreID        string `json:"storeid" bson:"storeid"`
	TerminalID     string `json:"terminalid" bson:"terminalid"`
	MerchantName   string `json:"merchantname" bson:"merchantname"`
	AccessCode     string `json:"accesscode" bson:"accesscode"`
	BankCharge     string `json:"bankcharge" bson:"bankcharge"`
	CustomerCharge string `json:"customercharge" bson:"customercharge"`
	IsSlipSave     bool   `json:"isslipsave" bson:"isslipsave"`

	// 0 = เงินเข้าทันที , 1 = สิ้นวัน , 2 = วันถัดไป
	CloseQr int8   `json:"closeqr" bson:"closeqr"`
	Secret  string `json:"secret" bson:"secret"`
	Token   string `json:"token" bson:"token"`

	// PaymentCode   string          `json:"paymentcode" bson:"paymentcode"`
	// CountryCode   string          `json:"countrycode" bson:"countrycode"`
	// PaymentLogo   string          `json:"paymentlogo" bson:"paymentlogo"`
	// PaymentType   int8            `json:"paymenttype" bson:"paymenttype"`
	// FeeRate       float64         `json:"feerate" bson:"feerate"`
	// WalletPayType int16           `json:"wallettype" bson:"wallettype"`
	// Names         *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	// BookBankCode  string          `json:"bookbankcode" bson:"bookbankcode"`
	// BankCode      string          `json:"bankcode" bson:"bankcode"`
}

type QrPaymentInfo struct {
	models.DocIdentity `bson:"inline"`
	QrPayment          `bson:"inline"`
}

func (QrPaymentInfo) CollectionName() string {
	return qrpaymentCollectionName
}

type QrPaymentData struct {
	models.HoldingCodeentity `bson:"inline"`
	QrPaymentInfo            `bson:"inline"`
}

type QrPaymentDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	QrPaymentData      `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (QrPaymentDoc) CollectionName() string {
	return qrpaymentCollectionName
}

type QrPaymentItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (QrPaymentItemGuid) CollectionName() string {
	return qrpaymentCollectionName
}

type QrPaymentActivity struct {
	QrPaymentData       `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (QrPaymentActivity) CollectionName() string {
	return qrpaymentCollectionName
}

type QrPaymentDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (QrPaymentDeleteActivity) CollectionName() string {
	return qrpaymentCollectionName
}
