package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const groupsubtwoproductCollectionName = "groupsubtwoproductmaster"

type GroupsubtwoProduct struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string          `json:"code" bson:"code"`
	Names                    *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	GroupMainGuid            string          `json:"groupmainguid" bson:"groupmainguid"`
	GroupMainNames           *[]models.NameX `json:"groupmainnames" bson:"-"`
	GroupSubGuid             string          `json:"groupsubguid" bson:"groupsubguid"`
	GroupSubNames            *[]models.NameX `json:"groupsubnames" bson:"-"`
}

type GroupsubtwoProductInfo struct {
	models.DocIdentity `bson:"inline"`
	GroupsubtwoProduct `bson:"inline"`
}

func (GroupsubtwoProductInfo) CollectionName() string {
	return groupsubtwoproductCollectionName
}

type GroupsubtwoProductData struct {
	models.HoldingCodeentity `bson:"inline"`
	GroupsubtwoProductInfo   `bson:"inline"`
}

type GroupsubtwoProductDoc struct {
	ID                     primitive.ObjectID `json:"id" bson:"id,omitempty"`
	GroupsubtwoProductData `bson:"inline"`
	models.ActivityDoc     `bson:"inline"`
}

func (GroupsubtwoProductDoc) CollectionName() string {
	return groupsubtwoproductCollectionName
}

type GroupsubtwoProductItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (GroupsubtwoProductItemGuid) CollectionName() string {
	return groupsubtwoproductCollectionName
}

type GroupsubtwoProductActivity struct {
	GroupsubtwoProductData `bson:"inline"`
	models.ActivityTime    `bson:"inline"`
}

func (GroupsubtwoProductActivity) CollectionName() string {
	return groupsubtwoproductCollectionName
}

type GroupsubtwoProductDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (GroupsubtwoProductDeleteActivity) CollectionName() string {
	return groupsubtwoproductCollectionName
}
