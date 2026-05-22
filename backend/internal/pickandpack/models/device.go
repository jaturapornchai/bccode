package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const deviceCollectionName = "pickandpackDevices"

type PickandpackDevice struct {
	Code string       `json:"code" bson:"code"`
	DeviceNumber string       `json:"devicenumber" bson:"devicenumber"`
	Name string       `json:"name" bson:"name"`
	DeviceType int8         `json:"devicetype" bson:"devicetype"` // ประเภทเครื่อง ex. 1 = Admin , 2 = Warehouse, 3 = Display
	ActivePin string       `json:"activepin" bson:"activepin"`
	Employees []PPEmployee `json:"employees" bson:"employees"`         // รายชื่อพนักงานที่ใช้เครื่องนี้
	WhCodes []string     `json:"whcode" bson:"whcode"`               // รหัสคลังสินค้า
	Locationcodes []string     `json:"locationcodes" bson:"locationcodes"` // รหัสตำแหน่งที่ตั้ง

}

type PPEmployee struct {
	models.DocIdentity `bson:"inline"`
	Code string          `json:"code" bson:"code"`
	Names *[]models.NameX `json:"names" bson:"names"`
}

type PickandpackDeviceInfo struct {
	models.DocIdentity `bson:"inline"`
	PickandpackDevice  `bson:"inline"`
}

func (PickandpackDeviceInfo) CollectionName() string {
	return deviceCollectionName
}

type PickandpackDeviceData struct {
	models.ShopIdentity   `bson:"inline"`
	PickandpackDeviceInfo `bson:"inline"`
}

type PickandpackDeviceDoc struct {
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	PickandpackDeviceData `bson:"inline"`
	models.ActivityDoc    `bson:"inline"`
}

func (PickandpackDeviceDoc) CollectionName() string {
	return deviceCollectionName
}

type PickandpackDeviceItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (PickandpackDeviceItemGuid) CollectionName() string {
	return deviceCollectionName
}

type PickandpackDeviceActivity struct {
	PickandpackDeviceData `bson:"inline"`
	models.ActivityTime   `bson:"inline"`
}

func (PickandpackDeviceActivity) CollectionName() string {
	return deviceCollectionName
}

type PickandpackDeviceDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (PickandpackDeviceDeleteActivity) CollectionName() string {
	return deviceCollectionName
}
