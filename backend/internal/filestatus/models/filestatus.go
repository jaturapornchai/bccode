package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const filestatusCollectionName = "fileStatus"

type FileStatus struct {
	models.PartitionIdentity `bson:"inline"`
	XOrder                   int                    `json:"xorder" bson:"xorder"`
	Username                 string                 `json:"-" bson:"username"`
	Menu                     string                 `json:"menu" bson:"menu"`
	JobID                    string                 `json:"jobid" bson:"jobid"`
	Path                     string                 `json:"path" bson:"path"`
	Status                   string                 `json:"status" bson:"status"`
	Filter                   map[string]interface{} `json:"filter" bson:"filter"`
}

type FileStatusInfo struct {
	models.DocIdentity `bson:"inline"`
	FileStatus         `bson:"inline"`
}

func (FileStatusInfo) CollectionName() string {
	return filestatusCollectionName
}

type FileStatusData struct {
	models.ShopIdentity `bson:"inline"`
	FileStatusInfo      `bson:"inline"`
}

type FileStatusDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	FileStatusData     `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (FileStatusDoc) CollectionName() string {
	return filestatusCollectionName
}

type FileStatusItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (FileStatusItemGuid) CollectionName() string {
	return filestatusCollectionName
}

type FileStatusActivity struct {
	FileStatusData      `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (FileStatusActivity) CollectionName() string {
	return filestatusCollectionName
}

type FileStatusDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (FileStatusDeleteActivity) CollectionName() string {
	return filestatusCollectionName
}
