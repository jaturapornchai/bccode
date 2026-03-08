package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const deviceCollectionName = "restarantDevices"

type Device struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string          `json:"code" bson:"code" validate:"required" `
	Type                     int16           `json:"type" bson:"type" `
	Names                    *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
}

type DeviceInfo struct {
	models.DocIdentity `bson:"inline"`
	Device             `bson:"inline"`
}

func (DeviceInfo) CollectionName() string {
	return deviceCollectionName
}

type DeviceData struct {
	models.ShopIdentity `bson:"inline"`
	DeviceInfo          `bson:"inline"`
}

type DeviceDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	DeviceData         `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (DeviceDoc) CollectionName() string {
	return deviceCollectionName
}

type DeviceItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (DeviceItemGuid) CollectionName() string {
	return deviceCollectionName
}

type DeviceActivity struct {
	DeviceData          `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (DeviceActivity) CollectionName() string {
	return deviceCollectionName
}

type DeviceDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (DeviceDeleteActivity) CollectionName() string {
	return deviceCollectionName
}
