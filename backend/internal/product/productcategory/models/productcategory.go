package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const productcategoryCollectionName = "productCategories"

type ProductCategory struct {
	models.PartitionIdentity `bson:"inline"`
	ChildCount               int                           `json:"childcount" bson:"childcount"`
	ParentGUID               string                        `json:"parentguid" bson:"parentguid"`
	ParentGUIDAll            string                        `json:"parentguidall" bson:"parentguidall"`
	ImageUri                 string                        `json:"imageuri" bson:"imageuri"`
	Names                    *[]models.NameX               `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	XSorts                   *[]models.XSort               `json:"xsorts" bson:"xsorts" validate:"unique=Code,dive"`
	CodeList                 *[]CodeXSort                  `json:"codelist" bson:"codelist" validate:"unique=Barcode,dive"`
	UseImageOrColor          bool                          `json:"useimageorcolor" bson:"useimageorcolor"`
	ColorSelect              string                        `json:"colorselect" bson:"colorselect"`
	ColorSelectHex           string                        `json:"colorselecthex" bson:"colorselecthex"`
	IsDisabled               bool                          `json:"isdisabled" bson:"isdisabled"`
	CoverURI                 string                        `json:"coveruri" bson:"coveruri"`
	GroupNumber              int                           `json:"groupnumber" bson:"groupnumber"`
	TimeForSales             *[]ProductCategoryTimeForSale `json:"timeforsales" bson:"timeforsales"`
}

func (p *ProductCategory) EmptyOnNil() {

	if p.Names == nil {
		p.Names = &[]models.NameX{}
	}

	if p.XSorts == nil {
		p.XSorts = &[]models.XSort{}
	}

	if p.CodeList == nil {
		p.CodeList = &[]CodeXSort{}
	}

	if p.TimeForSales == nil {
		p.TimeForSales = &[]ProductCategoryTimeForSale{}
	}
}

type ProductCategoryTimeForSale struct {
	DaysOfWeek []int8 `json:"daysofweek" bson:"daysofweek"`
	FromDate   string `json:"fromdate" bson:"fromdate"`
	ToDate     string `json:"todate" bson:"todate"`
	FromTime   string `json:"fromtime" bson:"fromtime"`
	ToTime     string `json:"totime" bson:"totime"`
}

type CodeXSort struct {
	Code             string          `json:"code" bson:"code"`
	XOrder           uint            `json:"xorder" bson:"xorder" validate:"min=0,max=4294967295"`
	Barcode          string          `json:"barcode" bson:"barcode"`
	UnitCode         string          `json:"unitcode" bson:"unitcode"`
	UnitNames        *[]models.NameX `json:"unitnames" bson:"unitnames"`
	Names            *[]models.NameX `json:"names" bson:"names" `
	ManufacturerGUID string          `json:"manufacturerguid" bson:"manufacturerguid"`
}

type ProductCategoryInfo struct {
	models.DocIdentity `bson:"inline"`
	ProductCategory    `bson:"inline"`
}

func (ProductCategoryInfo) CollectionName() string {
	return productcategoryCollectionName
}

type ProductCategoryData struct {
	models.ShopIdentity `bson:"inline"`
	ProductCategoryInfo `bson:"inline"`
}

type ProductCategoryDoc struct {
	ID                  primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ProductCategoryData `bson:"inline"`
	models.ActivityDoc  `bson:"inline"`
}

func (ProductCategoryDoc) CollectionName() string {
	return productcategoryCollectionName
}

type ProductCategoryItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (ProductCategoryItemGuid) CollectionName() string {
	return productcategoryCollectionName
}

type ProductCategoryActivity struct {
	ProductCategoryData `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ProductCategoryActivity) CollectionName() string {
	return productcategoryCollectionName
}

type ProductCategoryDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ProductCategoryDeleteActivity) CollectionName() string {
	return productcategoryCollectionName
}

type BarcodesModifyReqesut struct {
}
