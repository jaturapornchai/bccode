package models

import (
	"smlcloudplatform/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const productPriceHistoryCollectionName = "productBarcodesPriceHistory"

// ประวัติการแก้ไขราคาสินค้า
type ProductPriceHistory struct {
	ID                       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	models.HoldingCodeentity `bson:"inline"`
	models.DocIdentity       `bson:"inline"`

	// ข้อมูลสินค้า
	ProductBarcodeGUID string `json:"productbarcodeguid" bson:"productbarcodeguid"` // GUID ของ ProductBarcode
	Barcode            string `json:"barcode" bson:"barcode"`
	ProductName        string `json:"productname" bson:"productname"`

	// ข้อมูลราคา
	PriceType       string  `json:"pricetype" bson:"pricetype"`   // "normal", "member", "delivery"
	KeyNumber       int     `json:"key_number" bson:"key_number"` // 1=normal, 2=member, 3=delivery
	OldPrice        float64 `json:"oldprice" bson:"oldprice"`
	NewPrice        float64 `json:"newprice" bson:"newprice"`
	PriceDifference float64 `json:"pricedifference" bson:"pricedifference"` // NewPrice - OldPrice

	// ข้อมูลการทำรายการ
	Action    string    `json:"action" bson:"action"`       // "create", "update"
	CreatedBy string    `json:"createdby" bson:"createdby"` // Username ผู้ทำรายการ
	CreatedAt time.Time `json:"created_at" bson:"created_at"`

	// ข้อมูลเพิ่มเติม
	Remark string `json:"remark" bson:"remark"`
}

func (ProductPriceHistory) CollectionName() string {
	return productPriceHistoryCollectionName
}

// ProductPriceHistoryInfo สำหรับแสดงผล
type ProductPriceHistoryInfo struct {
	models.DocIdentity       `bson:"inline"`
	models.HoldingCodeentity `bson:"inline"`
	ProductPriceHistory      `bson:"inline"`
}

func (ProductPriceHistoryInfo) CollectionName() string {
	return productPriceHistoryCollectionName
}

// PriceChangeRequest สำหรับ request
type PriceChangeRequest struct {
	ProductBarcodeGUID string         `json:"productbarcodeguid" validate:"required"`
	Barcode            string         `json:"barcode" validate:"required"`
	ProductName        string         `json:"productname"`
	OldPrices          []ProductPrice `json:"oldprices"`
	NewPrices          []ProductPrice `json:"newprices"`
	Remark             string         `json:"remark"`
}

// PriceHistoryFilter สำหรับ filter
type PriceHistoryFilter struct {
	Barcode            string `json:"barcode"`
	ProductBarcodeGUID string `json:"productbarcodeguid"`
	Action             string `json:"action"`
	CreatedBy          string `json:"createdby"`
	FromDate           string `json:"fromdate"`
	ToDate             string `json:"todate"`
}
