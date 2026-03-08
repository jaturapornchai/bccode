package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const customergroupCollectionName = "customerGroups"

type CustomerGroup struct {
	models.PartitionIdentity `bson:"inline"`
	GroupCode                string          `json:"groupcode" bson:"groupcode"`
	Names                    *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
}

type CustomerGroupInfo struct {
	models.DocIdentity `bson:"inline"`
	CustomerGroup      `bson:"inline"`
}

func (CustomerGroupInfo) CollectionName() string {
	return customergroupCollectionName
}

type CustomerGroupData struct {
	models.ShopIdentity `bson:"inline"`
	CustomerGroupInfo   `bson:"inline"`
}

type CustomerGroupDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CustomerGroupData  `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (CustomerGroupDoc) CollectionName() string {
	return customergroupCollectionName
}

type CustomerGroupItemGuid struct {
	GroupCode string `json:"groupcode" bson:"groupcode"`
}

func (CustomerGroupItemGuid) CollectionName() string {
	return customergroupCollectionName
}

type CustomerGroupActivity struct {
	CustomerGroupData   `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (CustomerGroupActivity) CollectionName() string {
	return customergroupCollectionName
}

type CustomerGroupDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (CustomerGroupDeleteActivity) CollectionName() string {
	return customergroupCollectionName
}
