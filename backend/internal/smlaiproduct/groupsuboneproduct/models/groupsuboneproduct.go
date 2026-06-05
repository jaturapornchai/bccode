package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const groupsuboneproductCollectionName = "groupsuboneProductMaster"

type GroupsuboneProduct struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string          `json:"code" bson:"code"`
	Names                    *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	GroupMainGuid            string          `json:"groupmainguid" bson:"groupmainguid"`
	GroupMainNames           *[]models.NameX `json:"groupmainnames" bson:"-"`
}

type GroupsuboneProductInfo struct {
	models.DocIdentity `bson:"inline"`
	GroupsuboneProduct `bson:"inline"`
}

func (GroupsuboneProductInfo) CollectionName() string {
	return groupsuboneproductCollectionName
}

type GroupsuboneProductData struct {
	models.HoldingCodeentity `bson:"inline"`
	GroupsuboneProductInfo   `bson:"inline"`
}

type GroupsuboneProductDoc struct {
	ID                     primitive.ObjectID `json:"id" bson:"id,omitempty"`
	GroupsuboneProductData `bson:"inline"`
	models.ActivityDoc     `bson:"inline"`
}

func (GroupsuboneProductDoc) CollectionName() string {
	return groupsuboneproductCollectionName
}

type GroupsuboneProductItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (GroupsuboneProductItemGuid) CollectionName() string {
	return groupsuboneproductCollectionName
}

type GroupsuboneProductActivity struct {
	GroupsuboneProductData `bson:"inline"`
	models.ActivityTime    `bson:"inline"`
}

func (GroupsuboneProductActivity) CollectionName() string {
	return groupsuboneproductCollectionName
}

type GroupsuboneProductDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (GroupsuboneProductDeleteActivity) CollectionName() string {
	return groupsuboneproductCollectionName
}
