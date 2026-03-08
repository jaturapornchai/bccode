package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const masterexpenseCollectionName = "masterExpenses"

type MasterExpense struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string          `json:"code" bson:"code"`
	Names                    *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	AccountCode              string          `json:"accountcode" bson:"accountcode"`
	AccountName              string          `json:"accountname" bson:"accountname"`
}

type MasterExpenseInfo struct {
	models.DocIdentity `bson:"inline"`
	MasterExpense      `bson:"inline"`
}

func (MasterExpenseInfo) CollectionName() string {
	return masterexpenseCollectionName
}

type MasterExpenseData struct {
	models.ShopIdentity `bson:"inline"`
	MasterExpenseInfo   `bson:"inline"`
}

type MasterExpenseDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	MasterExpenseData  `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (MasterExpenseDoc) CollectionName() string {
	return masterexpenseCollectionName
}

type MasterExpenseItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (MasterExpenseItemGuid) CollectionName() string {
	return masterexpenseCollectionName
}

type MasterExpenseActivity struct {
	MasterExpenseData   `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (MasterExpenseActivity) CollectionName() string {
	return masterexpenseCollectionName
}

type MasterExpenseDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (MasterExpenseDeleteActivity) CollectionName() string {
	return masterexpenseCollectionName
}
