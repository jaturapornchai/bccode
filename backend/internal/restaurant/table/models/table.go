package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const tableCollectionName = "restaurantTables"

type Table struct {
	GroupNumber int             `json:"groupnumber" bson:"groupnumber"`
	Number      string          `json:"number" bson:"number"`
	Names       *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	Seat        int8            `json:"seat" bson:"seat"`
	Zone        string          `json:"zone" bson:"zone"`
	ZoneNumber  int8            `json:"zonenumber" bson:"zonenumber"`
	XOrder      uint            `json:"xorder" bson:"xorder" validate:"min=0,max=4294967295"`
}

type TableInfo struct {
	models.DocIdentity `bson:"inline"`
	Table              `bson:"inline"`
}

func (TableInfo) CollectionName() string {
	return tableCollectionName
}

type TableData struct {
	models.Identity `bson:"inline"`
	TableInfo       `bson:"inline"`
}

type TableDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TableData          `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
	models.LastUpdate  `bson:"inline"`
}

func (TableDoc) CollectionName() string {
	return tableCollectionName
}

// Extra
type TableItemGuid struct {
	Code string `json:"code" bson:"code" gorm:"code"`
}

func (TableItemGuid) CollectionName() string {
	return tableCollectionName
}

type TableActivity struct {
	TableData           `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (TableActivity) CollectionName() string {
	return tableCollectionName
}

type TableDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (TableDeleteActivity) CollectionName() string {
	return tableCollectionName
}

type TableInfoResponse struct {
	Success bool      `json:"success"`
	Data    TableInfo `json:"data,omitempty"`
}

type TablePageResponse struct {
	Success    bool                          `json:"success"`
	Data       []TableInfo                   `json:"data,omitempty"`
	Pagination models.PaginationDataResponse `json:"pagination,omitempty"`
}

type TableLastActivityResponse struct {
	New    []TableActivity       `json:"new" `
	Remove []TableDeleteActivity `json:"remove"`
}

type TableFetchUpdateResponse struct {
	Success    bool                          `json:"success"`
	Data       TableLastActivityResponse     `json:"data,omitempty"`
	Pagination models.PaginationDataResponse `json:"pagination,omitempty"`
}

type XOrderRequest struct {
	GuidFixed string `json:"guidfixed" bson:"guidfixed" validate:"required,min=1"`
	XOrder    uint   `json:"xorder" bson:"xorder" validate:"min=0,max=4294967295"`
}
