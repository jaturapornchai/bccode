package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const stockadjustmentCollectionName = "transactionStockAdjustment"

type StockAdjustment struct {
	models.PartitionIdentity `bson:"inline"`
	transmodels.Transaction  `bson:"inline"`
}
type StockAdjustmentInfo struct {
	models.DocIdentity `bson:"inline"`
	StockAdjustment    `bson:"inline"`
}

func (StockAdjustmentInfo) CollectionName() string {
	return stockadjustmentCollectionName
}

type StockAdjustmentData struct {
	models.ShopIdentity `bson:"inline"`
	StockAdjustmentInfo `bson:"inline"`
}

type StockAdjustmentDoc struct {
	ID                  primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	StockAdjustmentData `bson:"inline"`
	models.ActivityDoc  `bson:"inline"`
}

func (StockAdjustmentDoc) CollectionName() string {
	return stockadjustmentCollectionName
}

type StockAdjustmentItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (StockAdjustmentItemGuid) CollectionName() string {
	return stockadjustmentCollectionName
}

type StockAdjustmentActivity struct {
	StockAdjustmentData `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (StockAdjustmentActivity) CollectionName() string {
	return stockadjustmentCollectionName
}

type StockAdjustmentDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (StockAdjustmentDeleteActivity) CollectionName() string {
	return stockadjustmentCollectionName
}
