package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const saleinvoicereturnCollectionName = "transactionsaleinvoicereturn"

type SaleInvoiceReturn struct {
	models.PartitionIdentity `bson:"inline"`
	transmodels.Transaction  `bson:"inline"`
	IsPOS                    bool `json:"ispos" bson:"ispos"`

	CouponNo          string  `json:"couponno" bson:"couponno"`
	CouponAmount      float64 `json:"couponamount" bson:"couponamount"`
	CouponDescription string  `json:"coupondescription" bson:"coupondescription"`

	QRCode       string  `json:"qrcode" bson:"qrcode"`
	QRCodeAmount float64 `json:"qrcodeamount" bson:"qrcodeamount"`

	ChequeNo          string  `json:"chequeno" bson:"chequeno"`
	ChequeBookNumber  string  `json:"chequebooknumber" bson:"chequebooknumber"`
	ChequeBookCode    string  `json:"chequebookcode" bson:"chequebookcode"`
	ChequeDueDate     string  `json:"chequeduedate" bson:"chequeduedate"`
	ChequeAmount      float64 `json:"chequeamount" bson:"chequeamount"`
	SaleChannelCode   string  `json:"salechannelcode" bson:"csalechannelode"`
	SaleChannelGP     float64 `json:"salechannelgp" bson:"salechannelgp"`
	SaleChannelGPType int8    `json:"salechannelgptype" bson:"salechannelgptype"`
	PointsCode        string  `json:"pointscode" bson:"pointscode"`

	RefTotalOriginal float64 `json:"reftotaloriginal" bson:"reftotaloriginal"` // มูลค่าตามใบกำกับเดิม
	RefTotalCorrect  float64 `json:"reftotalcorrect" bson:"reftotalcorrect"`   // มูลค่าที่ถูกต้อง
	RefTotalDiff     float64 `json:"reftotaldiff" bson:"reftotaldiff"`         // ผลต่าง

	SlipUrl            string   `json:"slipurl" bson:"slipurl"`
	SlipQrUrl          string   `json:"slipqrurl" bson:"slipqrurl"`
	SlipUrlHistories   []string `json:"slipurlhistories" bson:"slipurlhistories"`
	SlipQrUrlHistories []string `json:"slipqrurlhistories" bson:"slipqrurlhistories"`
	// PosID              string   `json:"posid" bson:"posid"`
	MachineCode     string `json:"machinecode" bson:"machinecode"`
	ZoneGroupNumber string `json:"zonegroupnumber" bson:"zonegroupnumber"`
}

type SaleInvoiceReturnInfo struct {
	models.DocIdentity `bson:"inline"`
	SaleInvoiceReturn  `bson:"inline"`
}

func (SaleInvoiceReturnInfo) CollectionName() string {
	return saleinvoicereturnCollectionName
}

type SaleInvoiceReturnData struct {
	models.HoldingCodeentity `bson:"inline"`
	SaleInvoiceReturnInfo    `bson:"inline"`
}

type SaleInvoiceReturnDoc struct {
	ID                    primitive.ObjectID `json:"id" bson:"id,omitempty"`
	SaleInvoiceReturnData `bson:"inline"`
	models.ActivityDoc    `bson:"inline"`
}

func (SaleInvoiceReturnDoc) CollectionName() string {
	return saleinvoicereturnCollectionName
}

type SaleInvoiceReturnItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (SaleInvoiceReturnItemGuid) CollectionName() string {
	return saleinvoicereturnCollectionName
}

type SaleInvoiceReturnActivity struct {
	SaleInvoiceReturnData `bson:"inline"`
	models.ActivityTime   `bson:"inline"`
}

func (SaleInvoiceReturnActivity) CollectionName() string {
	return saleinvoicereturnCollectionName
}

type SaleInvoiceReturnDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (SaleInvoiceReturnDeleteActivity) CollectionName() string {
	return saleinvoicereturnCollectionName
}
