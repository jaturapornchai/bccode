package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const creditorgroupCollectionName = "creditorGroup"

type CreditorGroup struct {
	models.PartitionIdentity `bson:"inline"`
	GroupCode                string          `json:"groupcode" bson:"groupcode"`
	Names                    *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
}

type CreditorGroupInfo struct {
	models.DocIdentity `bson:"inline"`
	CreditorGroup      `bson:"inline"`
}

func (CreditorGroupInfo) CollectionName() string {
	return creditorgroupCollectionName
}

type CreditorGroupData struct {
	models.ShopIdentity `bson:"inline"`
	CreditorGroupInfo   `bson:"inline"`
}

type CreditorGroupDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CreditorGroupData  `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (CreditorGroupDoc) CollectionName() string {
	return creditorgroupCollectionName
}

type CreditorGroupItemGuid struct {
	GroupCode string `json:"groupcode" bson:"groupcode"`
}

func (CreditorGroupItemGuid) CollectionName() string {
	return creditorgroupCollectionName
}

type CreditorGroupActivity struct {
	CreditorGroupData   `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (CreditorGroupActivity) CollectionName() string {
	return creditorgroupCollectionName
}

type CreditorGroupDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (CreditorGroupDeleteActivity) CollectionName() string {
	return creditorgroupCollectionName
}
