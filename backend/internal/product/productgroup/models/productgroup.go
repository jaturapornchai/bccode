package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const productgroupCollectionName = "productGroups"

type ProductGroup struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string          `json:"code" bson:"code"`
	Names                    *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
}

type ProductGroupInfo struct {
	models.DocIdentity `bson:"inline"`
	ProductGroup       `bson:"inline"`
}

func (ProductGroupInfo) CollectionName() string {
	return productgroupCollectionName
}

type ProductGroupData struct {
	models.ShopIdentity `bson:"inline"`
	ProductGroupInfo    `bson:"inline"`
}

type ProductGroupDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ProductGroupData   `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (ProductGroupDoc) CollectionName() string {
	return productgroupCollectionName
}

type ProductGroupItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (ProductGroupItemGuid) CollectionName() string {
	return productgroupCollectionName
}

type ProductGroupActivity struct {
	ProductGroupData    `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ProductGroupActivity) CollectionName() string {
	return productgroupCollectionName
}

type ProductGroupDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ProductGroupDeleteActivity) CollectionName() string {
	return productgroupCollectionName
}
