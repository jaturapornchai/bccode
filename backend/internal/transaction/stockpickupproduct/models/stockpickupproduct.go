package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const stockpickupproductCollectionName = "transactionStockPickupProduct"

type StockPickupProduct struct {
	models.PartitionIdentity `bson:"inline"`
	transmodels.Transaction  `bson:"inline"`
}

type StockPickupProductInfo struct {
	models.DocIdentity `bson:"inline"`
	StockPickupProduct `bson:"inline"`
}

func (StockPickupProductInfo) CollectionName() string {
	return stockpickupproductCollectionName
}

type StockPickupProductData struct {
	models.ShopIdentity    `bson:"inline"`
	StockPickupProductInfo `bson:"inline"`
}

type StockPickupProductDoc struct {
	ID                     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	StockPickupProductData `bson:"inline"`
	models.ActivityDoc     `bson:"inline"`
}

func (StockPickupProductDoc) CollectionName() string {
	return stockpickupproductCollectionName
}

type StockPickupProductItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (StockPickupProductItemGuid) CollectionName() string {
	return stockpickupproductCollectionName
}

type StockPickupProductActivity struct {
	StockPickupProductData `bson:"inline"`
	models.ActivityTime    `bson:"inline"`
}

func (StockPickupProductActivity) CollectionName() string {
	return stockpickupproductCollectionName
}

type StockPickupProductDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (StockPickupProductDeleteActivity) CollectionName() string {
	return stockpickupproductCollectionName
}
