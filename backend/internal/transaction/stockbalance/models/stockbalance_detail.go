package models

import trans_models "smlcloudplatform/internal/transaction/models"

const stockbalanceDetailCollectionName = "transactionStockBalanceDetails"

type StockBalanceDetail struct {
	DocNo string `json:"docno" bson:"docno" validate:"required"`
	trans_models.Detail
}

func (StockBalanceDetail) CollectionName() string {
	return stockbalanceDetailCollectionName
}
