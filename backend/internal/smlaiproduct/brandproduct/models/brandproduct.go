package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const brandproductCollectionName = "brandProductMaster"

type BrandProduct struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string          `json:"code" bson:"code"`
	Names                    *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
}

type BrandProductInfo struct {
	models.DocIdentity `bson:"inline"`
	BrandProduct       `bson:"inline"`
}

func (BrandProductInfo) CollectionName() string {
	return brandproductCollectionName
}

type BrandProductData struct {
	models.HoldingCodeentity `bson:"inline"`
	BrandProductInfo         `bson:"inline"`
}

type BrandProductDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"id,omitempty"`
	BrandProductData   `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (BrandProductDoc) CollectionName() string {
	return brandproductCollectionName
}

type BrandProductItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (BrandProductItemGuid) CollectionName() string {
	return brandproductCollectionName
}

type BrandProductActivity struct {
	BrandProductData    `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (BrandProductActivity) CollectionName() string {
	return brandproductCollectionName
}

type BrandProductDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (BrandProductDeleteActivity) CollectionName() string {
	return brandproductCollectionName
}
