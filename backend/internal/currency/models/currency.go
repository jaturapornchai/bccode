package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const currencyCollectionName = "currency"

// ExchangeRateEntry - รายการอัตราแลกเปลี่ยนแต่ละวัน (embedded ใน Currency)
type ExchangeRateEntry struct {
	GuidFixed string  `json:"guid_fixed" bson:"guid_fixed"` // GUID สำหรับอ้างอิง
	Date string  `json:"date" bson:"date"`           // YYYY-MM-DD
	Rate float64 `json:"rate" bson:"rate"`           // อัตราแลกเปลี่ยนเป็นเงินบาท (1 USD = X THB)
}

// Currency - สกุลเงิน พร้อมประวัติอัตราแลกเปลี่ยน
type Currency struct {
	Code string              `json:"code" bson:"code"`                       // USD, EUR, JPY, etc.
	Name string              `json:"name" bson:"name"`                       // US Dollar, Euro, Japanese Yen
	Symbol string              `json:"symbol" bson:"symbol"`                   // $, €, ¥
	IsDisabled bool                `json:"isdisabled" bson:"isdisabled"`           // สถานะการใช้งาน
	ExchangeRates []ExchangeRateEntry `json:"exchange_rates" bson:"exchange_rates"`   // ประวัติอัตราแลกเปลี่ยน
}

type CurrencyInfo struct {
	models.DocIdentity `bson:"inline"`
	Currency  `bson:"inline"`
}

func (CurrencyInfo) CollectionName() string {
	return currencyCollectionName
}

type CurrencyData struct {
	models.ShopIdentity `bson:"inline"`
	CurrencyInfo  `bson:"inline"`
}

type CurrencyDoc struct {
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CurrencyData  `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (CurrencyDoc) CollectionName() string {
	return currencyCollectionName
}

type CurrencyActivity struct {
	CurrencyData  `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (CurrencyActivity) CollectionName() string {
	return currencyCollectionName
}

type CurrencyDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (CurrencyDeleteActivity) CollectionName() string {
	return currencyCollectionName
}
