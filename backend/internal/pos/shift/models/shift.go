package models

import (
	"smlcloudplatform/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const shiftCollectionName = "shift"

type Shift struct {
	models.PartitionIdentity `bson:"inline"`
	UserCode                 string    `json:"usercode" bson:"usercode"`
	Username                 string    `json:"username" bson:"username"`
	PosId                    string    `json:"posid" bson:"posid"`
	DocNo                    string    `json:"docno" bson:"docno"`
	DocType                  int8      `json:"doctype" bson:"doctype"`
	DocDate                  time.Time `json:"docdate" bson:"docdate"`
	Remark                   string    `json:"remark" bson:"remark"`
	Amount                   float64   `json:"amount" bson:"amount"`
	CreditCard               float64   `json:"creditcard" bson:"creditcard"`
	PromptPay                float64   `json:"promptpay" bson:"promptpay"`
	Transfer                 float64   `json:"transfer" bson:"transfer"`
	Cheque                   float64   `json:"cheque" bson:"cheque"`
	Coupon                   float64   `json:"coupon" bson:"coupon"`
}

type ShiftInfo struct {
	models.DocIdentity `bson:"inline"`
	Shift              `bson:"inline"`
}

func (ShiftInfo) CollectionName() string {
	return shiftCollectionName
}

type ShiftData struct {
	models.ShopIdentity `bson:"inline"`
	ShiftInfo           `bson:"inline"`
}

type ShiftDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ShiftData          `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (ShiftDoc) CollectionName() string {
	return shiftCollectionName
}

type ShiftItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (ShiftItemGuid) CollectionName() string {
	return shiftCollectionName
}

type ShiftActivity struct {
	ShiftData           `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ShiftActivity) CollectionName() string {
	return shiftCollectionName
}

type ShiftDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ShiftDeleteActivity) CollectionName() string {
	return shiftCollectionName
}
