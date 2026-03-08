package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const reportqueryCollectionName = "reportQueryMongo"

type ReportQuery struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string         `json:"code" bson:"code" validate:"required,min=1,max=50"`
	Collection               string         `json:"collection"`
	Filter                   string         `json:"filter"`
	Fields                   *[]string      `json:"fields"`
	Params                   *[]ReportParam `json:"params"`
	IsApproved               bool           `json:"isapproved" bson:"isapproved"`
	IsActived                bool           `json:"isactived" bson:"isactived"`
}

type ReportParam struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type ReportQueryInfo struct {
	models.DocIdentity `bson:"inline"`
	ReportQuery        `bson:"inline"`
}

func (ReportQueryInfo) CollectionName() string {
	return reportqueryCollectionName
}

type ReportQueryData struct {
	models.ShopIdentity `bson:"inline"`
	ReportQueryInfo     `bson:"inline"`
}

type ReportQueryDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ReportQueryData    `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (ReportQueryDoc) CollectionName() string {
	return reportqueryCollectionName
}

type ReportQueryItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (ReportQueryItemGuid) CollectionName() string {
	return reportqueryCollectionName
}

type ReportQueryActivity struct {
	ReportQueryData     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ReportQueryActivity) CollectionName() string {
	return reportqueryCollectionName
}

type ReportQueryDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ReportQueryDeleteActivity) CollectionName() string {
	return reportqueryCollectionName
}
