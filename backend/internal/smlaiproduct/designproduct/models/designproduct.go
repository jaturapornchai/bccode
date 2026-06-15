package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const designproductCollectionName = "designproductmaster"

type DesignProduct struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string          `json:"code" bson:"code"`
	Names                    *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
}

type DesignProductInfo struct {
	models.DocIdentity `bson:"inline"`
	DesignProduct      `bson:"inline"`
}

func (DesignProductInfo) CollectionName() string {
	return designproductCollectionName
}

type DesignProductData struct {
	models.HoldingCodeentity `bson:"inline"`
	DesignProductInfo        `bson:"inline"`
}

type DesignProductDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	DesignProductData  `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (DesignProductDoc) CollectionName() string {
	return designproductCollectionName
}

type DesignProductItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (DesignProductItemGuid) CollectionName() string {
	return designproductCollectionName
}

type DesignProductActivity struct {
	DesignProductData   `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (DesignProductActivity) CollectionName() string {
	return designproductCollectionName
}

type DesignProductDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (DesignProductDeleteActivity) CollectionName() string {
	return designproductCollectionName
}
