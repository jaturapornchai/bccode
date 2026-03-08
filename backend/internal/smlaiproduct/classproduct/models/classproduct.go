package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const classproductCollectionName = "classProductMaster"

type ClassProduct struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string          `json:"code" bson:"code"`
	Names                    *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
}

type ClassProductInfo struct {
	models.DocIdentity `bson:"inline"`
	ClassProduct       `bson:"inline"`
}

func (ClassProductInfo) CollectionName() string {
	return classproductCollectionName
}

type ClassProductData struct {
	models.ShopIdentity `bson:"inline"`
	ClassProductInfo    `bson:"inline"`
}

type ClassProductDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ClassProductData   `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (ClassProductDoc) CollectionName() string {
	return classproductCollectionName
}

type ClassProductItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (ClassProductItemGuid) CollectionName() string {
	return classproductCollectionName
}

type ClassProductActivity struct {
	ClassProductData    `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ClassProductActivity) CollectionName() string {
	return classproductCollectionName
}

type ClassProductDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ClassProductDeleteActivity) CollectionName() string {
	return classproductCollectionName
}
