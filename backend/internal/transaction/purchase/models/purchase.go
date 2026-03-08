package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const purchaseCollectionName = "transactionPurchase"

type Purchase struct {
	models.PartitionIdentity `bson:"inline"`
	transmodels.Transaction  `bson:"inline"`
}

type PurchaseInfo struct {
	models.DocIdentity `bson:"inline"`
	Purchase           `bson:"inline"`
}

func (PurchaseInfo) CollectionName() string {
	return purchaseCollectionName
}

type PurchaseData struct {
	models.ShopIdentity `bson:"inline"`
	PurchaseInfo        `bson:"inline"`
}

type PurchaseDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	PurchaseData       `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (PurchaseDoc) CollectionName() string {
	return purchaseCollectionName
}

type PurchaseItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (PurchaseItemGuid) CollectionName() string {
	return purchaseCollectionName
}

type PurchaseActivity struct {
	PurchaseData        `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (PurchaseActivity) CollectionName() string {
	return purchaseCollectionName
}

type PurchaseDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (PurchaseDeleteActivity) CollectionName() string {
	return purchaseCollectionName
}
