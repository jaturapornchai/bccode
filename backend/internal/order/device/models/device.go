package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const deviceCollectionName = "orderDevices"

type OrderDevice struct {
	Code         string `json:"code" bson:"code"`
	DeviceNumber string `json:"devicenumber" bson:"devicenumber"`
	DocFormat    string `json:"docformat" bson:"docformat"`
	DeviceType   int8   `json:"devicetype" bson:"devicetype"` // ประเภทเครื่อง ex.เครื่องลูกค้า,เครื่องพนักงาน
	ActivePin    string `json:"activepin" bson:"activepin"`
	IsPOSActive  bool   `json:"isposactive" bson:"isposactive"` // ใช้งาน POS
	SettingCode  string `json:"settingcode" bson:"settingcode"`
}

type OrderDeviceInfo struct {
	models.DocIdentity `bson:"inline"`
	OrderDevice        `bson:"inline"`
}

func (OrderDeviceInfo) CollectionName() string {
	return deviceCollectionName
}

type OrderDeviceData struct {
	models.ShopIdentity `bson:"inline"`
	OrderDeviceInfo     `bson:"inline"`
}

type OrderDeviceDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	OrderDeviceData    `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (OrderDeviceDoc) CollectionName() string {
	return deviceCollectionName
}

type OrderDeviceItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (OrderDeviceItemGuid) CollectionName() string {
	return deviceCollectionName
}

type OrderDeviceActivity struct {
	OrderDeviceData     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (OrderDeviceActivity) CollectionName() string {
	return deviceCollectionName
}

type OrderDeviceDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (OrderDeviceDeleteActivity) CollectionName() string {
	return deviceCollectionName
}
