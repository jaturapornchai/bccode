package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const stockreceiveproductCollectionName = "transactionStockReceiveProduct"

type StockReceiveProduct struct {
	models.PartitionIdentity `bson:"inline"`
	transmodels.Transaction  `bson:"inline"`
}

type StockReceiveProductInfo struct {
	models.DocIdentity  `bson:"inline"`
	StockReceiveProduct `bson:"inline"`
}

func (StockReceiveProductInfo) CollectionName() string {
	return stockreceiveproductCollectionName
}

type StockReceiveProductData struct {
	models.ShopIdentity     `bson:"inline"`
	StockReceiveProductInfo `bson:"inline"`
}

type StockReceiveProductDoc struct {
	ID                      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	StockReceiveProductData `bson:"inline"`
	models.ActivityDoc      `bson:"inline"`
}

func (StockReceiveProductDoc) CollectionName() string {
	return stockreceiveproductCollectionName
}

type StockReceiveProductItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (StockReceiveProductItemGuid) CollectionName() string {
	return stockreceiveproductCollectionName
}

type StockReceiveProductActivity struct {
	StockReceiveProductData `bson:"inline"`
	models.ActivityTime     `bson:"inline"`
}

func (StockReceiveProductActivity) CollectionName() string {
	return stockreceiveproductCollectionName
}

type StockReceiveProductDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (StockReceiveProductDeleteActivity) CollectionName() string {
	return stockreceiveproductCollectionName
}
