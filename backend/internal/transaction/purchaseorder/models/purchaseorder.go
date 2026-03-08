package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const purchaseorderCollectionName = "transactionPurchaseOrder"

type PurchaseOrder struct {
	models.PartitionIdentity `bson:"inline"`
	transmodels.Transaction  `bson:"inline"`
}
type PurchaseOrderInfo struct {
	models.DocIdentity `bson:"inline"`
	PurchaseOrder      `bson:"inline"`
}

func (PurchaseOrderInfo) CollectionName() string {
	return purchaseorderCollectionName
}

type PurchaseOrderData struct {
	models.ShopIdentity `bson:"inline"`
	PurchaseOrderInfo   `bson:"inline"`
}

type PurchaseOrderDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	PurchaseOrderData  `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (PurchaseOrderDoc) CollectionName() string {
	return purchaseorderCollectionName
}

type PurchaseOrderItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (PurchaseOrderItemGuid) CollectionName() string {
	return purchaseorderCollectionName
}

type PurchaseOrderActivity struct {
	PurchaseOrderData   `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (PurchaseOrderActivity) CollectionName() string {
	return purchaseorderCollectionName
}

type PurchaseOrderDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (PurchaseOrderDeleteActivity) CollectionName() string {
	return purchaseorderCollectionName
}
