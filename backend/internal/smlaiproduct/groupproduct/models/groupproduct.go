package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const groupproductCollectionName = "groupproductmaster"

type GroupProduct struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string          `json:"code" bson:"code"`
	Names                    *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
}

type GroupProductInfo struct {
	models.DocIdentity `bson:"inline"`
	GroupProduct       `bson:"inline"`
}

func (GroupProductInfo) CollectionName() string {
	return groupproductCollectionName
}

type GroupProductData struct {
	models.HoldingCodeentity `bson:"inline"`
	GroupProductInfo         `bson:"inline"`
}

type GroupProductDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	GroupProductData   `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (GroupProductDoc) CollectionName() string {
	return groupproductCollectionName
}

type GroupProductItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (GroupProductItemGuid) CollectionName() string {
	return groupproductCollectionName
}

type GroupProductActivity struct {
	GroupProductData    `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (GroupProductActivity) CollectionName() string {
	return groupproductCollectionName
}

type GroupProductDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (GroupProductDeleteActivity) CollectionName() string {
	return groupproductCollectionName
}
