package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const exchangeRateHistoryCollectionName = "exchangeRateHistory"

// ExchangeRateHistory - ประวัติอัตราแลกเปลี่ยน (แยก document ต่าง date)
type ExchangeRateHistory struct {
	Currency string  `json:"currency" bson:"currency"` // USD, EUR, JPY, etc.
	Date string  `json:"date" bson:"date"`         // YYYY-MM-DD
	Rate float64 `json:"rate" bson:"rate"`         // อัตราแลกเปลี่ยนเป็นเงินบาท (1 USD = X THB)
	Source string  `json:"source,omitempty" bson:"source,omitempty"` // แหล่งที่มา (manual, api, etc.)
	Note string  `json:"note,omitempty" bson:"note,omitempty"`     // หมายเหตุ
}

type ExchangeRateHistoryInfo struct {
	models.DocIdentity  `bson:"inline"`
	ExchangeRateHistory `bson:"inline"`
}

func (*ExchangeRateHistoryInfo) CollectionName() string {
	return exchangeRateHistoryCollectionName
}

type ExchangeRateHistoryData struct {
	models.ShopIdentity     `bson:"inline"`
	ExchangeRateHistoryInfo `bson:"inline"`
}

type ExchangeRateHistoryDoc struct {
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ExchangeRateHistoryData `bson:"inline"`
	models.ActivityDoc      `bson:"inline"`
}

func (*ExchangeRateHistoryDoc) CollectionName() string {
	return exchangeRateHistoryCollectionName
}

type ExchangeRateHistoryActivity struct {
	ExchangeRateHistoryData `bson:"inline"`
	models.ActivityTime     `bson:"inline"`
}

func (*ExchangeRateHistoryActivity) CollectionName() string {
	return exchangeRateHistoryCollectionName
}

type ExchangeRateHistoryDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (*ExchangeRateHistoryDeleteActivity) CollectionName() string {
	return exchangeRateHistoryCollectionName
}
