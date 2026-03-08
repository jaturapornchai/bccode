package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const collectionName = "notifierDevices"

type NotifierDevice struct {
	models.PartitionIdentity `bson:"inline"`
	FCMToken                 string `json:"fcmtoken" bson:"fcmtoken" validate:"required,min=1,max=255"`
	DeviceID                 string `json:"deviceid" bson:"deviceid"`
	DeviceName               string `json:"devicename" bson:"devicename"`
}

type NotifierDeviceInfo struct {
	models.DocIdentity `bson:"inline"`
	NotifierDevice     `bson:"inline"`
}

func (NotifierDeviceInfo) CollectionName() string {
	return collectionName
}

type NotifierDeviceData struct {
	models.ShopIdentity `bson:"inline"`
	NotifierDeviceInfo  `bson:"inline"`
}

type NotifierDeviceDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	NotifierDeviceData `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (NotifierDeviceDoc) CollectionName() string {
	return collectionName
}

type NotifierDeviceAuth struct {
	ShopID      string `json:"shopid" bson:"shopid"`
	UserAddedBy string `json:"useraddedby" bson:"useraddedby"`
	RefCode     string `json:"refcode" bson:"refcode"`
}

type NotifierDeviceConfirmAuthPayload struct {
	RefCode    string `json:"refcode" bson:"refcode" validate:"required,min=1,max=255"`
	FCMToken   string `json:"fcmtoken" bson:"fcmtoken" validate:"required,min=1,max=255"`
	DeviceID   string `json:"deviceid" bson:"deviceid"`
	DeviceName string `json:"devicename" bson:"devicename"`
}
