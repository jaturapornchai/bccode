package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const saleorderCollectionName = "transactionAccrualreceive"

type Accrualreceive struct {
	models.PartitionIdentity `bson:"inline"`
	transmodels.Transaction  `bson:"inline"`
	RefPurchasePartial       []string `json:"refpurchasepartial" bson:"refpurchasepartial"`
}

type AccrualreceiveInfo struct {
	models.DocIdentity `bson:"inline"`
	Accrualreceive     `bson:"inline"`
}

func (AccrualreceiveInfo) CollectionName() string {
	return saleorderCollectionName
}

type AccrualreceiveData struct {
	models.HoldingCodeentity `bson:"inline"`
	AccrualreceiveInfo       `bson:"inline"`
}

type AccrualreceiveDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	AccrualreceiveData `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (AccrualreceiveDoc) CollectionName() string {
	return saleorderCollectionName
}

type AccrualreceiveItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (AccrualreceiveItemGuid) CollectionName() string {
	return saleorderCollectionName
}

type AccrualreceiveActivity struct {
	AccrualreceiveData  `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (AccrualreceiveActivity) CollectionName() string {
	return saleorderCollectionName
}

type AccrualreceiveDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (AccrualreceiveDeleteActivity) CollectionName() string {
	return saleorderCollectionName
}
