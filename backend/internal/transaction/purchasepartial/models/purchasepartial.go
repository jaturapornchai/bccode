package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const saleorderCollectionName = "transactionpurchasepartial"

type Purchasepartial struct {
	models.PartitionIdentity `bson:"inline"`
	transmodels.Transaction  `bson:"inline"`
}
type PurchasepartialInfo struct {
	models.DocIdentity `bson:"inline"`
	Purchasepartial    `bson:"inline"`
}

func (PurchasepartialInfo) CollectionName() string {
	return saleorderCollectionName
}

type PurchasepartialData struct {
	models.HoldingCodeentity `bson:"inline"`
	PurchasepartialInfo      `bson:"inline"`
}

type PurchasepartialDoc struct {
	ID                  primitive.ObjectID `json:"id" bson:"id,omitempty"`
	PurchasepartialData `bson:"inline"`
	models.ActivityDoc  `bson:"inline"`
}

func (PurchasepartialDoc) CollectionName() string {
	return saleorderCollectionName
}

type PurchasepartialItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (PurchasepartialItemGuid) CollectionName() string {
	return saleorderCollectionName
}

type PurchasepartialActivity struct {
	PurchasepartialData `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (PurchasepartialActivity) CollectionName() string {
	return saleorderCollectionName
}

type PurchasepartialDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (PurchasepartialDeleteActivity) CollectionName() string {
	return saleorderCollectionName
}
