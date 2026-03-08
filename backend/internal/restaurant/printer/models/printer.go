package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const printerCollectionName = "restaurantPrinters"

type Printer struct {
	Code    string          `json:"code" bson:"code"`
	Names   *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	Address string          `json:"address" bson:"address" `
	Type    int8            `json:"type" bson:"type"`
}

type PrinterInfo struct {
	models.DocIdentity `bson:"inline"`
	Printer            `bson:"inline"`
}

func (PrinterInfo) CollectionName() string {
	return printerCollectionName
}

type PrinterData struct {
	models.ShopIdentity `bson:"inline"`
	PrinterInfo         `bson:"inline"`
}

type PrinterDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	PrinterData        `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
	models.LastUpdate  `bson:"inline"`
}

func (PrinterDoc) CollectionName() string {
	return printerCollectionName
}

//Extra

type PrinterItemGuid struct {
	Code string `json:"code" bson:"code" gorm:"code"`
}

func (PrinterItemGuid) CollectionName() string {
	return printerCollectionName
}

type PrinterActivity struct {
	PrinterData         `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (PrinterActivity) CollectionName() string {
	return printerCollectionName
}

type PrinterDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (PrinterDeleteActivity) CollectionName() string {
	return printerCollectionName
}

type PrinterInfoResponse struct {
	Success bool        `json:"success"`
	Data    PrinterInfo `json:"data,omitempty"`
}

type PrinterPageResponse struct {
	Success    bool                          `json:"success"`
	Data       []PrinterInfo                 `json:"data,omitempty"`
	Pagination models.PaginationDataResponse `json:"pagination,omitempty"`
}

type PrinterLastActivityResponse struct {
	New    []PrinterActivity       `json:"new" `
	Remove []PrinterDeleteActivity `json:"remove"`
}

type PrinterFetchUpdateResponse struct {
	Success    bool                          `json:"success"`
	Data       PrinterLastActivityResponse   `json:"data,omitempty"`
	Pagination models.PaginationDataResponse `json:"pagination,omitempty"`
}
