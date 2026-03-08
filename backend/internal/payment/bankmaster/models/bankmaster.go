package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const bankmasterCollectionName = "bankMaster"

type BankMaster struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string          `json:"code" bson:"code"`
	Names                    *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	Logo                     *string         `json:"logo" bson:"logo"`
}

type BankMasterInfo struct {
	models.DocIdentity `bson:"inline"`
	BankMaster         `bson:"inline"`
}

func (BankMasterInfo) CollectionName() string {
	return bankmasterCollectionName
}

type BankMasterData struct {
	models.ShopIdentity `bson:"inline"`
	BankMasterInfo      `bson:"inline"`
}

type BankMasterDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	BankMasterData     `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (BankMasterDoc) CollectionName() string {
	return bankmasterCollectionName
}

type BankMasterItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (BankMasterItemGuid) CollectionName() string {
	return bankmasterCollectionName
}

type BankMasterActivity struct {
	BankMasterData      `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (BankMasterActivity) CollectionName() string {
	return bankmasterCollectionName
}

type BankMasterDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (BankMasterDeleteActivity) CollectionName() string {
	return bankmasterCollectionName
}
