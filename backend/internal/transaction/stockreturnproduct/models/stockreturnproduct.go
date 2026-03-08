package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const stockreturnproductCollectionName = "transactionStockReturnProduct"

type StockReturnProduct struct {
	models.PartitionIdentity `bson:"inline"`
	transmodels.Transaction  `bson:"inline"`
}

type StockReturnProductInfo struct {
	models.DocIdentity `bson:"inline"`
	StockReturnProduct `bson:"inline"`
}

func (StockReturnProductInfo) CollectionName() string {
	return stockreturnproductCollectionName
}

type StockReturnProductData struct {
	models.ShopIdentity    `bson:"inline"`
	StockReturnProductInfo `bson:"inline"`
}

type StockReturnProductDoc struct {
	ID                     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	StockReturnProductData `bson:"inline"`
	models.ActivityDoc     `bson:"inline"`
}

func (StockReturnProductDoc) CollectionName() string {
	return stockreturnproductCollectionName
}

type StockReturnProductItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (StockReturnProductItemGuid) CollectionName() string {
	return stockreturnproductCollectionName
}

type StockReturnProductActivity struct {
	StockReturnProductData `bson:"inline"`
	models.ActivityTime    `bson:"inline"`
}

func (StockReturnProductActivity) CollectionName() string {
	return stockreturnproductCollectionName
}

type StockReturnProductDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (StockReturnProductDeleteActivity) CollectionName() string {
	return stockreturnproductCollectionName
}
