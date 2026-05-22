package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const productCollectionName = "products"

type Product struct {
	models.PartitionIdentity `bson:"inline"`
	Code string             `json:"code" bson:"code"`
	Names *[]models.NameX    `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	GroupCode string             `json:"group_code" bson:"group_code"`
	GroupNames *[]models.NameX    `json:"group_names" bson:"group_names"`
	ManufacturerGUID string             `json:"manufacturerguid" bson:"manufacturerguid"`
	ManufacturerCode string             `json:"manufacturercode" bson:"manufacturercode"`
	ManufacturerNames *[]models.NameX    `json:"manufacturernames" bson:"manufacturernames"`
	Dimensions []ProductDimension `json:"dimensions" bson:"dimensions"`
	VatType int8               `json:"vat_type" bson:"vat_type"`
	Barcodes []Barcodes         `json:"barcodes,omitempty"`
	ItemType int8               `json:"item_type" bson:"item_type"`
	UnitGuid string             `json:"unitguid" bson:"unitguid"`
}

type Barcodes struct {
	GuidFixed string          `json:"guid_fixed" gorm:"-"`
	ItemUnitCode string          `json:"item_unit_code" gorm:"-"`
	ItemUnitNames *[]models.NameX `json:"itemunitnames" gorm:"-"`
	Barcode string          `json:"barcode" gorm:"-"`
	Prices *[]ProductPrice `json:"prices" gorm:"-"`
	Condition bool            `json:"condition" gorm:"-"`
	DivideValue float64         `json:"dividevalue" gorm:"-"`
	StandValue float64         `json:"standvalue" gorm:"-"`
	Qty float64         `json:"qty" gorm:"-"`
	IsMainBarcode bool            `json:"is_main_barcode" gorm:"-"`
}

type ProductPrice struct {
	KeyNumber int     `json:"key_number" gorm:"-"`
	Price float64 `json:"price" gorm:"-"`
}

type ProductDimension struct {
	models.DocIdentity `bson:"inline"`
	Names *[]models.NameX      `json:"names" bson:"names"`
	IsDisabled bool                 `json:"isdisabled" bson:"isdisabled"`
	Item ProductDimensionItem `json:"item" bson:"item"`
}

type ProductDimensionItem struct {
	models.DocIdentity `bson:"inline"`
	Names *[]models.NameX `json:"names" bson:"names"`
	IsDisabled bool            `json:"isdisabled" bson:"isdisabled"`
}

type ProductInfo struct {
	models.DocIdentity `bson:"inline"`
	Product  `bson:"inline"`
}

func (ProductInfo) CollectionName() string {
	return productCollectionName
}

type ProductData struct {
	models.ShopIdentity `bson:"inline"`
	ProductInfo  `bson:"inline"`
}

type ProductDoc struct {
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ProductData  `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (ProductDoc) CollectionName() string {
	return productCollectionName
}

type ProductItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (ProductItemGuid) CollectionName() string {
	return productCollectionName
}

type ProductActivity struct {
	ProductData  `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ProductActivity) CollectionName() string {
	return productCollectionName
}

type ProductDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ProductDeleteActivity) CollectionName() string {
	return productCollectionName
}
