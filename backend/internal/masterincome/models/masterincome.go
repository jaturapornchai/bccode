package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const masterincomeCollectionName = "masterIncomes"

type MasterIncome struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string          `json:"code" bson:"code"`
	Names                    *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	AccountCode              string          `json:"accountcode" bson:"accountcode"`
	AccountName              string          `json:"accountname" bson:"accountname"`
}

type MasterIncomeInfo struct {
	models.DocIdentity `bson:"inline"`
	MasterIncome       `bson:"inline"`
}

func (MasterIncomeInfo) CollectionName() string {
	return masterincomeCollectionName
}

type MasterIncomeData struct {
	models.ShopIdentity `bson:"inline"`
	MasterIncomeInfo    `bson:"inline"`
}

type MasterIncomeDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	MasterIncomeData   `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (MasterIncomeDoc) CollectionName() string {
	return masterincomeCollectionName
}

type MasterIncomeItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (MasterIncomeItemGuid) CollectionName() string {
	return masterincomeCollectionName
}

type MasterIncomeActivity struct {
	MasterIncomeData    `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (MasterIncomeActivity) CollectionName() string {
	return masterincomeCollectionName
}

type MasterIncomeDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (MasterIncomeDeleteActivity) CollectionName() string {
	return masterincomeCollectionName
}
