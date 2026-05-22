package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const groupsubtwoproductCollectionName = "groupsubtwoProductMaster"

type GroupsubtwoProduct struct {
	models.PartitionIdentity `bson:"inline"`
	Code string          `json:"code" bson:"code"`
	Names *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	GroupMainGuid string          `json:"group_main_guid" bson:"group_main_guid"`
	GroupMainNames *[]models.NameX `json:"group_main_names" bson:"-"`
	GroupSubGuid string          `json:"group_sub_guid" bson:"group_sub_guid"`
	GroupSubNames *[]models.NameX `json:"group_sub_names" bson:"-"`
}

type GroupsubtwoProductInfo struct {
	models.DocIdentity `bson:"inline"`
	GroupsubtwoProduct `bson:"inline"`
}

func (GroupsubtwoProductInfo) CollectionName() string {
	return groupsubtwoproductCollectionName
}

type GroupsubtwoProductData struct {
	models.ShopIdentity    `bson:"inline"`
	GroupsubtwoProductInfo `bson:"inline"`
}

type GroupsubtwoProductDoc struct {
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`
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
