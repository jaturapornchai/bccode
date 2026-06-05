package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const gradeproductCollectionName = "gradeProductMaster"

type GradeProduct struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string          `json:"code" bson:"code"`
	Names                    *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
}

type GradeProductInfo struct {
	models.DocIdentity `bson:"inline"`
	GradeProduct       `bson:"inline"`
}

func (GradeProductInfo) CollectionName() string {
	return gradeproductCollectionName
}

type GradeProductData struct {
	models.HoldingCodeentity `bson:"inline"`
	GradeProductInfo         `bson:"inline"`
}

type GradeProductDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"id,omitempty"`
	GradeProductData   `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (GradeProductDoc) CollectionName() string {
	return gradeproductCollectionName
}

type GradeProductItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (GradeProductItemGuid) CollectionName() string {
	return gradeproductCollectionName
}

type GradeProductActivity struct {
	GradeProductData    `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (GradeProductActivity) CollectionName() string {
	return gradeproductCollectionName
}

type GradeProductDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (GradeProductDeleteActivity) CollectionName() string {
	return gradeproductCollectionName
}
