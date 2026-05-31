package models

import (
	"smlcloudplatform/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const productBarcodeBOMCollectionName = "productBarcodeBOMs"

type BOMProductBarcode struct {
	BarcodeGuidFixed string          `json:"guid_fixed" bson:"guid_fixed"`
	Level            int             `json:"level" bson:"level"`
	Names            *[]models.NameX `json:"names" bson:"names"`
	ItemUnitCode     string          `json:"item_unit_code" bson:"item_unit_code"`
	ItemUnitNames    *[]models.NameX `json:"itemunitnames" bson:"itemunitnames"`
	Barcode          string          `json:"barcode" bson:"barcode" validate:"required,min=1"`
	RefType          string          `json:"ref_type" bson:"ref_type,omitempty"`
	Condition        bool            `json:"condition" bson:"condition"`
	DivideValue      float64         `json:"dividevalue" bson:"dividevalue"`
	StandValue       float64         `json:"standvalue" bson:"standvalue"`
	Qty              float64         `json:"qty" bson:"qty"`
	YieldPercent     float64         `json:"yield_percent" bson:"yield_percent,omitempty"`
	AverageCost      float64         `json:"averagecost" bson:"averagecost,omitempty"`
	Price            float64         `json:"price" bson:"price,omitempty"`
	MaterialType     int8            `json:"materialtype" bson:"materialtype,omitempty"`
}

type ProductBarcodeBOMView struct {
	BOMProductBarcode `bson:"inline"`
	ImageURI          string                   `json:"imageuri" bson:"imageuri"`
	BOM               *[]ProductBarcodeBOMView `json:"bom" bson:"bom"`
}

type ProductBarcodeBOMVersion struct {
	GuidFixed string                   `json:"guid_fixed" bson:"guid_fixed"`
	StartDate time.Time                `json:"start_date" bson:"start_date"`
	EndDate   *time.Time               `json:"end_date" bson:"end_date"`
	BOM       *[]ProductBarcodeBOMView `json:"bom" bson:"bom"`
}

type ProductBarcodeBOMSaveRequest struct {
	GuidFixed     string                     `json:"guid_fixed"`
	Barcode       string                     `json:"barcode"`
	Names         *[]models.NameX            `json:"names"`
	ItemUnitCode  string                     `json:"item_unit_code"`
	ItemUnitNames *[]models.NameX            `json:"itemunitnames"`
	Price         float64                    `json:"price"`
	BOM           []ProductBarcodeBOMView    `json:"bom"`
	BOMs          []ProductBarcodeBOMVersion `json:"boms"`
}

func (b *ProductBarcodeBOMView) EmptyOnNil() {

	if b.Names == nil {
		b.Names = &[]models.NameX{}
	}

	if b.ItemUnitNames == nil {
		b.ItemUnitNames = &[]models.NameX{}
	}

	if b.BOM == nil {
		b.BOM = &[]ProductBarcodeBOMView{}
	}
}

type ProductBarcodeBOMViewInfo struct {
	models.DocIdentity    `bson:"inline"`
	ProductBarcodeBOMView `bson:"inline"`
	CheckSum              string                      `json:"checksum" bson:"checksum"`
	IsCurrentUse          bool                        `json:"iscurrentuse" bson:"iscurrentuse"`
	UseInDate             time.Time                   `json:"useindate" bson:"useindate"`
	StartDate             time.Time                   `json:"startdate" bson:"startdate"`
	EndDate               *time.Time                  `json:"enddate" bson:"enddate"`
	BOMs                  *[]ProductBarcodeBOMVersion `json:"boms" bson:"boms,omitempty"`
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

type ProductBarcodeBOMViewGuid struct {
	models.DocIdentity `bson:"inline"`
}

func (ProductBarcodeBOMViewGuid) CollectionName() string {
	return productBarcodeBOMCollectionName
}

type ProductBarcodeBOMViewActivity struct {
	ProductBarcodeBOMViewData `bson:"inline"`
	models.ActivityTime       `bson:"inline"`
}

func (ProductBarcodeBOMViewActivity) CollectionName() string {
	return productBarcodeBOMCollectionName
}

type ProductBarcodeBOMViewDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ProductBarcodeBOMViewDeleteActivity) CollectionName() string {
	return productBarcodeBOMCollectionName
}
