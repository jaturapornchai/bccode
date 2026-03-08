package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const unitCollectionName = "units"

type Unit struct {
	models.PartitionIdentity `bson:"inline"`
	UnitCode                 string `json:"unitcode" bson:"unitcode" validate:"required,max=100"`
	models.UnitName          `bson:"inline"`
	Names                    *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
}

type UnitInfo struct {
	models.DocIdentity `bson:"inline"`
	Unit               `bson:"inline"`
}

func (UnitInfo) CollectionName() string {
	return unitCollectionName
}

type UnitData struct {
	models.ShopIdentity `bson:"inline"`
	UnitInfo            `bson:"inline"`
}

type UnitDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UnitData           `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (UnitDoc) CollectionName() string {
	return unitCollectionName
}

type UnitItemGuid struct {
	UnitCode string `json:"unitcode" bson:"unitcode" `
}

func (UnitItemGuid) CollectionName() string {
	return unitCollectionName
}

type UnitActivity struct {
	UnitData            `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (UnitActivity) CollectionName() string {
	return unitCollectionName
}

type UnitDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (UnitDeleteActivity) CollectionName() string {
	return unitCollectionName
}
