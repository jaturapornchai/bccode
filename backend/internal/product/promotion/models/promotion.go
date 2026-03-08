package models

import (
	"smlcloudplatform/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const promotionCollectionName = "productPromotions"

type Promotion struct {
	models.PartitionIdentity `bson:"inline"`
	PromotionType            int8                       `json:"promotiontype" bson:"promotiontype"`
	Index                    int64                      `json:"index" bson:"index"`
	Code                     string                     `json:"code" bson:"code"`
	Name                     *[]models.NameX            `json:"name" bson:"name"`
	DateBegin                time.Time                  `json:"datebegin" bson:"datebegin"`
	DateEnd                  time.Time                  `json:"dateend" bson:"dateend"`
	CustomerOnly             int8                       `json:"customeronly" bson:"customeronly"`
	DiscountText             string                     `json:"discounttext" bson:"discounttext"`
	LimitQty                 float64                    `json:"limitqty" bson:"limitqty"`
	PromotionQty             float64                    `json:"promotionqty" bson:"promotionqty"`
	LimitAmount              float64                    `json:"limitamount" bson:"limitamount"`
	PromotionBarcodeInclude  *[]PromotionBarcodeInclude `json:"promotionbarcodeinclude" bson:"promotionbarcodeinclude"`
}

type PromotionBarcodeInclude struct {
	PromotionProduct *[]ProductBarcode `json:"promotionproduct" bson:"promotionproduct"`
	IncludeProduct   *[]ProductBarcode `json:"includeproduct" bson:"includeproduct"`
}

type ProductBarcode struct {
	GuidFixed    string          `json:"guidfixed" bson:"guidfixed"`
	DiscountText string          `json:"discounttext" bson:"discounttext"`
	Barcode      string          `json:"barcode" bson:"barcode"`
	Name         *[]models.NameX `json:"name" bson:"name"`
	UnitCode     string          `json:"unitcode" bson:"unitcode"`
	UnitName     *[]models.NameX `json:"unitname" bson:"unitname"`
	Price        float64         `json:"price" bson:"price"`
	Qty          float64         `json:"qty" bson:"qty"`
}

type PromotionDetail struct {
	DetailType     int8           `json:"detailtype" bson:"detailtype"` // 0: discount, 1: buy x get y, 2: buy x get y with discount
	Minimum        float64        `json:"minimum" bson:"minimum"`
	Discount       float64        `json:"discount" bson:"discount"`
	ProductBarcode ProductBarcode `json:"productbarcode" bson:"productbarcode"`
}

type PromotionInfo struct {
	models.DocIdentity `bson:"inline"`
	Promotion          `bson:"inline"`
}

func (PromotionInfo) CollectionName() string {
	return promotionCollectionName
}

type PromotionData struct {
	models.ShopIdentity `bson:"inline"`
	PromotionInfo       `bson:"inline"`
}

type PromotionDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	PromotionData      `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (PromotionDoc) CollectionName() string {
	return promotionCollectionName
}

type PromotionItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (PromotionItemGuid) CollectionName() string {
	return promotionCollectionName
}

type PromotionActivity struct {
	PromotionData       `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (PromotionActivity) CollectionName() string {
	return promotionCollectionName
}

type PromotionDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (PromotionDeleteActivity) CollectionName() string {
	return promotionCollectionName
}
