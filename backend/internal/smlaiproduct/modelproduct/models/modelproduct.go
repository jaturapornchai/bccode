package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const modelproductCollectionName = "modelProductMaster"

type ModelProduct struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string          `json:"code" bson:"code"`
	Names                    *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
}

type ModelProductInfo struct {
	models.DocIdentity `bson:"inline"`
	ModelProduct       `bson:"inline"`
}

func (ModelProductInfo) CollectionName() string {
	return modelproductCollectionName
}

type ModelProductData struct {
	models.ShopIdentity `bson:"inline"`
	ModelProductInfo    `bson:"inline"`
}

type ModelProductDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ModelProductData   `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (ModelProductDoc) CollectionName() string {
	return modelproductCollectionName
}

type ModelProductItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (ModelProductItemGuid) CollectionName() string {
	return modelproductCollectionName
}

type ModelProductActivity struct {
	ModelProductData    `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ModelProductActivity) CollectionName() string {
	return modelproductCollectionName
}

type ModelProductDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ModelProductDeleteActivity) CollectionName() string {
	return modelproductCollectionName
}
