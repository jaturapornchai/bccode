package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const saleorderCollectionName = "transactionSaleOrder"

type SaleOrder struct {
	models.PartitionIdentity `bson:"inline"`
	transmodels.Transaction  `bson:"inline"`
}
type SaleOrderInfo struct {
	models.DocIdentity `bson:"inline"`
	SaleOrder  `bson:"inline"`
}

func (SaleOrderInfo) CollectionName() string {
	return saleorderCollectionName
}

type SaleOrderData struct {
	models.ShopIdentity `bson:"inline"`
	SaleOrderInfo  `bson:"inline"`
}

type SaleOrderDoc struct {
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	SaleOrderData  `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (SaleOrderDoc) CollectionName() string {
	return saleorderCollectionName
}

type SaleOrderItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (SaleOrderItemGuid) CollectionName() string {
	return saleorderCollectionName
}

type SaleOrderActivity struct {
	SaleOrderData  `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (SaleOrderActivity) CollectionName() string {
	return saleorderCollectionName
}

type SaleOrderDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (SaleOrderDeleteActivity) CollectionName() string {
	return saleorderCollectionName
}
