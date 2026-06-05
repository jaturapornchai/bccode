package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const categoryproductCollectionName = "categoryProductMaster"

type CategoryProduct struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string          `json:"code" bson:"code"`
	Names                    *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
}

type CategoryProductInfo struct {
	models.DocIdentity `bson:"inline"`
	CategoryProduct    `bson:"inline"`
}

func (CategoryProductInfo) CollectionName() string {
	return categoryproductCollectionName
}

type CategoryProductData struct {
	models.HoldingCodeentity `bson:"inline"`
	CategoryProductInfo      `bson:"inline"`
}

type CategoryProductDoc struct {
	ID                  primitive.ObjectID `json:"id" bson:"id,omitempty"`
	CategoryProductData `bson:"inline"`
	models.ActivityDoc  `bson:"inline"`
}

func (CategoryProductDoc) CollectionName() string {
	return categoryproductCollectionName
}

type CategoryProductItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (CategoryProductItemGuid) CollectionName() string {
	return categoryproductCollectionName
}

type CategoryProductActivity struct {
	CategoryProductData `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (CategoryProductActivity) CollectionName() string {
	return categoryproductCollectionName
}

type CategoryProductDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (CategoryProductDeleteActivity) CollectionName() string {
	return categoryproductCollectionName
}
