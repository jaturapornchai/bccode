package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const staffCollectionName = "restaurantStaffs"

type Staff struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string          `json:"code" bson:"code"`
	Names                    *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	Email                    string          `json:"email" bson:"email" validate:"omitempty,email"`
	Cashier                  bool            `json:"cashier" bson:"cashier"`
	Order                    bool            `json:"order" bson:"order"`
}

type StaffInfo struct {
	models.DocIdentity `bson:"inline"`
	Staff              `bson:"inline"`
}

func (StaffInfo) CollectionName() string {
	return staffCollectionName
}

type StaffData struct {
	models.ShopIdentity `bson:"inline"`
	StaffInfo           `bson:"inline"`
}

type StaffDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	StaffData          `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (StaffDoc) CollectionName() string {
	return staffCollectionName
}

type StaffItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (StaffItemGuid) CollectionName() string {
	return staffCollectionName
}

type StaffActivity struct {
	StaffData           `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (StaffActivity) CollectionName() string {
	return staffCollectionName
}

type StaffDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (StaffDeleteActivity) CollectionName() string {
	return staffCollectionName
}
