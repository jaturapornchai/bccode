package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const businesstypeCollectionName = "organizationBusinessTypes"

type BusinessType struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string          `json:"code" bson:"code"`
	Names                    *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	IsDefault                bool            `json:"isdefault" bson:"isdefault"`
}

type BusinessTypeInfo struct {
	models.DocIdentity `bson:"inline"`
	BusinessType       `bson:"inline"`
}

func (BusinessTypeInfo) CollectionName() string {
	return businesstypeCollectionName
}

type BusinessTypeData struct {
	models.ShopIdentity `bson:"inline"`
	BusinessTypeInfo    `bson:"inline"`
}

type BusinessTypeDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	BusinessTypeData   `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (BusinessTypeDoc) CollectionName() string {
	return businesstypeCollectionName
}

type BusinessTypeItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (BusinessTypeItemGuid) CollectionName() string {
	return businesstypeCollectionName
}

type BusinessTypeActivity struct {
	BusinessTypeData    `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (BusinessTypeActivity) CollectionName() string {
	return businesstypeCollectionName
}

type BusinessTypeDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (BusinessTypeDeleteActivity) CollectionName() string {
	return businesstypeCollectionName
}
