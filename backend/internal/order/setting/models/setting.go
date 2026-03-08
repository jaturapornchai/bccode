package models

import (
	"smlcloudplatform/internal/models"
	branch "smlcloudplatform/internal/organization/branch/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const settingCollectionName = "orderSettings"

type OrderSetting struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string             `json:"code" bson:"code" validate:"required"`
	DocFormat                string             `json:"docformat" bson:"docformat"`
	Branch                   OrderSettingBranch `json:"branch" bson:"branch"`
	ActivePin                string             `json:"activepin" bson:"activepin"`
	// Slips             *[]OrderSettingSlip       `json:"slips" bson:"slips"`
	QRCodes *[]map[string]interface{} `json:"qrcodes" bson:"qrcodes"`
	// BillHeader *[]models.NameX `json:"billheader" bson:"billheader"`
	// BillFooter        *[]models.NameX           `json:"billfooter" bson:"billfooter"`
	MediaGUID string `json:"mediaguid" bson:"mediaguid"`

	// timezone.Timezone `bson:"inline"`
	TimeForSales        *[]OrderSettingTimeForSale `json:"timeforsales" bson:"timeforsales"` // เวลาขายเอลกอฮอล์
	LogoUrl             string                     `json:"logourl" bson:"logourl"`
	BackgroundUrl       string                     `json:"backgroundurl" bson:"backgroundurl"`
	LineOaImg           string                     `json:"lineoaimg" bson:"lineoaimg"`
	TableNumber         string                     `json:"tablenumber" bson:"tablenumber"` // เลขโต๊ะ
	DeviceType          int8                       `json:"devicetype" bson:"devicetype"`   // ประเภทเครื่อง ex.เครื่องลูกค้า,เครื่องพนักงาน
	IsPOSActive         bool                       `json:"isposactive" bson:"isposactive"` // ใช้งาน POS
	Label               string                     `json:"label" bson:"label"`
	SaleChannels        *[]string                  `json:"salechannels" bson:"salechannels"`
	AdminPin            string                     `json:"adminpin" bson:"adminpin"`
	CategoryGroupNumber int                        `json:"categorygroupnumber" bson:"categorygroupnumber"`
	KitchenGroupNumber  int                        `json:"kitchengroupnumber" bson:"kitchengroupnumber"`
	TableGroupNumber    int                        `json:"tablegroupnumber" bson:"tablegroupnumber"`
	ZoneGroupNumber     int                        `json:"zonegroupnumber" bson:"zonegroupnumber"`
	Emails              *[]string                  `json:"emails" bson:"emails"`

	// ภาษี
	IsVatRegister bool    `json:"isvatregister" bson:"isvatregister"` // จดทะเบียนภาษี
	VatType       int8    `json:"vattype" bson:"vattype"`             // ราคารวมภาษี หรือ ราคาไม่รวมภาษี
	VatRate       float64 `json:"vatrate" bson:"vatrate"`             // อัตราภาษี
}

type OrderSettingTimeForSale struct {
	Names *[]models.NameX `json:"names" bson:"names"`
	From  string          `json:"from" bson:"from"`
	To    string          `json:"to" bson:"to"`
}

type OrderSettingSlip struct {
	Code        string          `json:"code" bson:"code"`
	Name        string          `json:"name" bson:"name"`
	IsRequire   bool            `json:"isrequire" bson:"isrequire"`
	FormCode    string          `json:"formcode" bson:"formcode"`
	FormNames   *[]models.NameX `json:"formnames" bson:"formnames"`
	HeaderNames *[]models.NameX `json:"headernames" bson:"headernames"`
}

type POSEmployee struct {
	models.DocIdentity `bson:"inline"`
	Code               string    `json:"code" bson:"code"`
	Name               string    `json:"name" bson:"name"`
	Permissions        *[]string `json:"permissions" bson:"permissions"`
}

type OrderSettingBranch struct {
	models.DocIdentity `bson:"inline"`
	Code               string                         `json:"code" bson:"code"`
	Names              *[]models.NameX                `json:"names" bson:"names"`
	PaymentRounding    branch.PaymentRoundingSettings `json:"paymentrounding" bson:"paymentrounding"`
}

type OrderSettingWarehouse struct {
	models.DocIdentity `bson:"inline"`
	Code               string               `json:"code" bson:"code"`
	Names              *[]models.NameNormal `json:"names" bson:"names"`
}

type OrderSettingLocation struct {
	Code  string               `json:"code" bson:"code"`
	Names *[]models.NameNormal `json:"names" bson:"names"`
}

type SettingInfo struct {
	models.DocIdentity `bson:"inline"`
	OrderSetting       `bson:"inline"`
}

func (SettingInfo) CollectionName() string {
	return settingCollectionName
}

type SettingData struct {
	models.ShopIdentity `bson:"inline"`
	SettingInfo         `bson:"inline"`
}

type SettingDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	SettingData        `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (SettingDoc) CollectionName() string {
	return settingCollectionName
}

type SettingItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (SettingItemGuid) CollectionName() string {
	return settingCollectionName
}

type SettingActivity struct {
	SettingData         `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (SettingActivity) CollectionName() string {
	return settingCollectionName
}

type SettingDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (SettingDeleteActivity) CollectionName() string {
	return settingCollectionName
}
