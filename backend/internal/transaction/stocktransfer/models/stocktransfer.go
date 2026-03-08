package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const stocktransferCollectionName = "transactionStockTransfer"

type StockTransfer struct {
	models.PartitionIdentity `bson:"inline"`
	transmodels.Transaction  `bson:"inline"`
}

type StockTransferInfo struct {
	models.DocIdentity `bson:"inline"`
	StockTransfer      `bson:"inline"`
}

func (StockTransferInfo) CollectionName() string {
	return stocktransferCollectionName
}

type StockTransferData struct {
	models.ShopIdentity `bson:"inline"`
	StockTransferInfo   `bson:"inline"`
}

type StockTransferDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	StockTransferData  `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (StockTransferDoc) CollectionName() string {
	return stocktransferCollectionName
}

type StockTransferItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (StockTransferItemGuid) CollectionName() string {
	return stocktransferCollectionName
}

type StockTransferActivity struct {
	StockTransferData   `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (StockTransferActivity) CollectionName() string {
	return stocktransferCollectionName
}

type StockTransferDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (StockTransferDeleteActivity) CollectionName() string {
	return stocktransferCollectionName
}
