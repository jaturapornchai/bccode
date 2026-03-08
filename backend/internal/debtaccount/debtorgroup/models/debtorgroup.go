package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const debtorgroupCollectionName = "debtorGroup"

type DebtorGroup struct {
	models.PartitionIdentity `bson:"inline"`
	GroupCode                string          `json:"groupcode" bson:"groupcode"`
	Names                    *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
}

type DebtorGroupInfo struct {
	models.DocIdentity `bson:"inline"`
	DebtorGroup        `bson:"inline"`
}

func (DebtorGroupInfo) CollectionName() string {
	return debtorgroupCollectionName
}

type DebtorGroupData struct {
	models.ShopIdentity `bson:"inline"`
	DebtorGroupInfo     `bson:"inline"`
}

type DebtorGroupDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	DebtorGroupData    `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (DebtorGroupDoc) CollectionName() string {
	return debtorgroupCollectionName
}

type DebtorGroupItemGuid struct {
	GroupCode string `json:"groupcode" bson:"groupcode"`
}

func (DebtorGroupItemGuid) CollectionName() string {
	return debtorgroupCollectionName
}

type DebtorGroupActivity struct {
	DebtorGroupData     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (DebtorGroupActivity) CollectionName() string {
	return debtorgroupCollectionName
}

type DebtorGroupDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (DebtorGroupDeleteActivity) CollectionName() string {
	return debtorgroupCollectionName
}
