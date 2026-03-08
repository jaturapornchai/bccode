package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const producttypeCollectionName = "productTypes"

type ProductType struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string          `json:"code" bson:"code"`
	Names                    *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
}

type ProductTypeInfo struct {
	models.DocIdentity `bson:"inline"`
	ProductType        `bson:"inline"`
}

func (ProductTypeInfo) CollectionName() string {
	return producttypeCollectionName
}

type ProductTypeData struct {
	models.ShopIdentity `bson:"inline"`
	ProductTypeInfo     `bson:"inline"`
}

type ProductTypeDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ProductTypeData    `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (ProductTypeDoc) CollectionName() string {
	return producttypeCollectionName
}

type ProductTypeItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (ProductTypeItemGuid) CollectionName() string {
	return producttypeCollectionName
}

type ProductTypeActivity struct {
	ProductTypeData     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ProductTypeActivity) CollectionName() string {
	return producttypeCollectionName
}

type ProductTypeDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ProductTypeDeleteActivity) CollectionName() string {
	return producttypeCollectionName
}
