package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const dimensionCollectionName = "dimension"

type Dimension struct {
	models.PartitionIdentity `bson:"inline"`
	Names                    *[]models.NameX `json:"names" bson:"names"`
	IsDisabled               bool            `json:"isdisabled" bson:"isdisabled"`
	Items                    []DimensionItem `json:"items" bson:"items"`
}

type DimensionItem struct {
	models.DocIdentity `bson:"inline"`
	Names              *[]models.NameX `json:"names" bson:"names"`
	IsDisabled         bool            `json:"isdisabled" bson:"isdisabled"`
}

type DimensionInfo struct {
	models.DocIdentity `bson:"inline"`
	Dimension          `bson:"inline"`
}

func (DimensionInfo) CollectionName() string {
	return dimensionCollectionName
}

type DimensionData struct {
	models.ShopIdentity `bson:"inline"`
	DimensionInfo       `bson:"inline"`
}

type DimensionDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	DimensionData      `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (DimensionDoc) CollectionName() string {
	return dimensionCollectionName
}

type DimensionItemGuid struct {
	GuidFixed string `json:"guidfixed" bson:"guidfixed"`
}

func (DimensionItemGuid) CollectionName() string {
	return dimensionCollectionName
}

type DimensionActivity struct {
	DimensionData       `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (DimensionActivity) CollectionName() string {
	return dimensionCollectionName
}

type DimensionDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (DimensionDeleteActivity) CollectionName() string {
	return dimensionCollectionName
}
