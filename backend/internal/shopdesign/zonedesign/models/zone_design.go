package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const zoneDesignCollectionName = "zoneDesigns"

type ZoneDesign struct {
	Code        string        `json:"code" bson:"code"`
	XOrder      int           `json:"xorder" bson:"xorder"`
	Width       float64       `json:"width" bson:"width"`
	Height      float64       `json:"height" bson:"height"`
	Tables      []TableDesign `json:"tables" bson:"tables"`
	Props       []PropDesign  `json:"props" bson:"props"`
	models.Name `bson:"inline"`
}

type ZoneDesignInfo struct {
	models.DocIdentity `bson:"inline"`
	ZoneDesign         `bson:"inline"`
}

func (ZoneDesignInfo) CollectionName() string {
	return zoneDesignCollectionName
}

type ZoneDesignData struct {
	models.ShopIdentity `bson:"inline"`
	ZoneDesignInfo      `bson:"inline"`
}

type ZoneDesignDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ZoneDesignData     `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
	models.LastUpdate  `bson:"inline"`
}

func (ZoneDesignDoc) CollectionName() string {
	return zoneDesignCollectionName
}
