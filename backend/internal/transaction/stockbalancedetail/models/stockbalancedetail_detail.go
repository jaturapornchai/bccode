package models

import (
	"smlcloudplatform/internal/models"
	trans_models "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const stockbalanceDetailCollectionName = "transactionStockBalanceDetails"

type StockBalanceDetail struct {
	DocNo               string `json:"docno" bson:"docno" validate:"required"`
	trans_models.Detail `bson:"inline"`
}

func (StockBalanceDetail) CollectionName() string {
	return stockbalanceDetailCollectionName
}

type StockBalanceDetailInfo struct {
	models.DocIdentity `bson:"inline"`
	StockBalanceDetail `bson:"inline"`
}

func (StockBalanceDetailInfo) CollectionName() string {
	return stockbalanceDetailCollectionName
}

type StockBalanceDetailData struct {
	models.ShopIdentity    `bson:"inline"`
	StockBalanceDetailInfo `bson:"inline"`
}

type StockBalanceDetailDoc struct {
	ID                     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	StockBalanceDetailData `bson:"inline"`
	models.ActivityDoc     `bson:"inline"`
}

func (StockBalanceDetailDoc) CollectionName() string {
	return stockbalanceDetailCollectionName
}

type StockBalanceDetailItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (StockBalanceDetailItemGuid) CollectionName() string {
	return stockbalanceDetailCollectionName
}

type StockBalanceDetailActivity struct {
	StockBalanceDetailData `bson:"inline"`
	models.ActivityTime    `bson:"inline"`
}

func (StockBalanceDetailActivity) CollectionName() string {
	return stockbalanceDetailCollectionName
}

type StockBalanceDetailDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (StockBalanceDetailDeleteActivity) CollectionName() string {
	return stockbalanceDetailCollectionName
}
