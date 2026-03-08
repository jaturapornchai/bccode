package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const productBarcodeBOMCollectionName = "productBarcodeBOMs"

// type ProductBarcodeBOM struct {
// 	Level         int             `json:"level" gorm:"column:level"`
// 	MainBarcode   string          `json:"mainbarcode" gorm:"column:mainbarcode"`
// 	GuidFixed     string          `json:"guidfixed" gorm:"column:guidfixed"`
// 	Names         *[]models.NameX `json:"names" gorm:"column:names"`
// 	ItemUnitCode  string          `json:"itemunitcode" gorm:"column:itemunitcode"`
// 	ItemUnitNames *[]models.NameX `json:"itemunitnames" gorm:"column:itemunitnames"`
// 	Barcode       string          `json:"barcode" gorm:"column:barcode"`
// 	Condition     bool            `json:"condition" gorm:"column:condition"`
// 	DivideValue   float64         `json:"dividevalue" gorm:"column:dividevalue"`
// 	StandValue    float64         `json:"standvalue" gorm:"column:standvalue"`
// 	Qty           float64         `json:"qty" gorm:"column:qty"`
// }

type BOMProductBarcode struct {
	BarcodeGuidFixed string          `json:"guidfixed" bson:"guidfixed"`
	Level            int             `json:"level" bson:"level"`
	Names            *[]models.NameX `json:"names" bson:"names"`
	ItemUnitCode     string          `json:"itemunitcode" bson:"itemunitcode"`
	ItemUnitNames    *[]models.NameX `json:"itemunitnames" bson:"itemunitnames"`
	Barcode          string          `json:"barcode" bson:"barcode" validate:"required,min=1"`
	Condition        bool            `json:"condition" bson:"condition"`
	DivideValue      float64         `json:"dividevalue" bson:"dividevalue"`
	StandValue       float64         `json:"standvalue" bson:"standvalue"`
	Qty              float64         `json:"qty" bson:"qty"`
}

type ProductBarcodeBOMView struct {
	BOMProductBarcode `bson:"inline"`
	ImageURI          string                  `json:"imageuri" bson:"imageuri"`
	BOM               []ProductBarcodeBOMView `json:"bom" bson:"bom"`
}

type ProductBarcodeBOMViewInfo struct {
	models.DocIdentity    `bson:"inline"`
	ProductBarcodeBOMView `bson:"inline"`
	CheckSum              string `json:"checksum" bson:"checksum"`
}

func (ProductBarcodeBOMViewInfo) CollectionName() string {
	return productBarcodeBOMCollectionName
}

type ProductBarcodeBOMViewData struct {
	models.ShopIdentity       `bson:"inline"`
	ProductBarcodeBOMViewInfo `bson:"inline"`
}

type ProductBarcodeBOMViewDoc struct {
	ID                        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ProductBarcodeBOMViewData `bson:"inline"`
	models.ActivityDoc        `bson:"inline"`
}

func (ProductBarcodeBOMViewDoc) CollectionName() string {
	return productBarcodeBOMCollectionName
}

func (pbbv *ProductBarcodeBOMView) FromProductBarcode(doc ProductBarcodeData) {
	pbbv.BarcodeGuidFixed = doc.GuidFixed
	pbbv.Names = doc.Names
	pbbv.ItemUnitCode = doc.ItemUnitCode
	pbbv.ItemUnitNames = doc.ItemUnitNames
	pbbv.Barcode = doc.Barcode
	pbbv.Condition = doc.Condition
	pbbv.DivideValue = doc.DivideValue
	pbbv.StandValue = doc.StandValue
	pbbv.Qty = doc.Qty
	pbbv.ImageURI = doc.ImageURI
}

func (pbbv *ProductBarcodeBOMView) FromProductBOM(doc ProductBarcodeData, docBom BOMProductBarcode) {
	pbbv.BarcodeGuidFixed = docBom.BarcodeGuidFixed

	pbbv.Barcode = docBom.Barcode
	pbbv.Condition = docBom.Condition
	pbbv.DivideValue = docBom.DivideValue
	pbbv.StandValue = docBom.StandValue
	pbbv.Qty = docBom.Qty

	pbbv.Names = doc.Names
	pbbv.ItemUnitCode = doc.ItemUnitCode
	pbbv.ItemUnitNames = doc.ItemUnitNames
	pbbv.ImageURI = doc.ImageURI
}

type ProductBarcodeBOMHistoryInfo struct {
	models.DocIdentity       `bson:"inline"`
	ProductBarcodeBOMHistory models.DocIdentity `bson:"inline"`
}
