package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const patternproductCollectionName = "patternproductmaster"

type PatternProduct struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string          `json:"code" bson:"code"`
	Names                    *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
}

type PatternProductInfo struct {
	models.DocIdentity `bson:"inline"`
	PatternProduct     `bson:"inline"`
}

func (PatternProductInfo) CollectionName() string {
	return patternproductCollectionName
}

type PatternProductData struct {
	models.HoldingCodeentity `bson:"inline"`
	PatternProductInfo       `bson:"inline"`
}

type PatternProductDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	PatternProductData `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (PatternProductDoc) CollectionName() string {
	return patternproductCollectionName
}

type PatternProductItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (PatternProductItemGuid) CollectionName() string {
	return patternproductCollectionName
}

type PatternProductActivity struct {
	PatternProductData  `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (PatternProductActivity) CollectionName() string {
	return patternproductCollectionName
}

type PatternProductDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (PatternProductDeleteActivity) CollectionName() string {
	return patternproductCollectionName
}
