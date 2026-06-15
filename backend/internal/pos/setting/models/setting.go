package models

import (
	salechannel_models "smlcloudplatform/internal/channel/salechannel/models"
	"smlcloudplatform/internal/models"
	"smlcloudplatform/internal/models/timezone"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const settingCollectionName = "possettings"

type Setting struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string                    `json:"code" bson:"code" validate:"required"`
	DocCode                  string                    `json:"doccode" bson:"doccode"`
	DeviceNumber             string                    `json:"devicenumber" bson:"devicenumber"`
	DocFormatTaxInv          string                    `json:"docformattaxinv" bson:"docformattaxinv"`
	DocFormatInv             string                    `json:"docformatinv" bson:"docformatinv"`
	ReceiptForm              string                    `json:"receiptform" bson:"receiptform"`
	Emails                   *[]string                 `json:"emails" bson:"emails"`
	Warehouse                PosSettingWarehouse       `json:"warehouse" bson:"warehouse"`
	Location                 PosSettingLocation        `json:"location" bson:"location"`
	Branch                   PosSettingBranch          `json:"branch" bson:"branch"`
	ActivePin                string                    `json:"activepin" bson:"activepin"`
	Employees                []POSEmployee             `json:"employees" bson:"employees"`
	DocFormateSaleReturn     string                    `json:"docformatesalereturn" bson:"docformatesalereturn"`
	Slips                    *[]PosSettingSlip         `json:"slips" bson:"slips"`
	IsEJournal               bool                      `json:"isejournal" bson:"isejournal"`
	Wallet                   string                    `json:"wallet" bson:"wallet"`
	LineOaImg                string                    `json:"lineoaimg" bson:"lineoaimg"`
	QRCodes                  *[]map[string]interface{} `json:"qrcodes" bson:"qrcodes"`
	CreditCards              *[]map[string]interface{} `json:"creditcards" bson:"creditcards"`
	Transfers                *[]map[string]interface{} `json:"transfers" bson:"transfers"` // Book Bank Transfer เงินโอน
	BillHeader               *[]models.NameX           `json:"billheader" bson:"billheader"`
	BillFooter               *[]models.NameX           `json:"billfooter" bson:"billfooter"`
	MediaGUID                string                    `json:"mediaguid" bson:"mediaguid"`
	CategoryGroupNumber      int                       `json:"categorygroupnumber" bson:"categorygroupnumber"`
	KitchenGroupNumber       int                       `json:"kitchengroupnumber" bson:"kitchengroupnumber"`
	TableGroupNumber         int                       `json:"tablegroupnumber" bson:"tablegroupnumber"`
	ZoneGroupNumber          int                       `json:"zonegroupnumber" bson:"zonegroupnumber"`

	timezone.Timezone `bson:"inline"`
	TimeForSales      *[]PosSettingTimeForSale          `json:"timeforsales" bson:"timeforsales"`
	LogoUrl           string                            `json:"logourl" bson:"logourl"`
	IsUseCreadit      bool                              `json:"isusecreadit" bson:"isusecreadit"` // ขายเชื่อได้
	BusinessType      int8                              `json:"businesstype" bson:"businesstype"` // ประเภทธุรกิจ
	PaymentType       int8                              `json:"paymenttype" bson:"paymenttype"`   // ประเภทการชำระเงิน ex. กินก่อนจ่าย จ่ายก่อนกิน
	IsPOSActive       bool                              `json:"isposactive" bson:"isposactive"`   // ใช้งาน POS
	SaleChanels       *[]salechannel_models.SaleChannel `json:"salechanels" bson:"salechanels"`
	ServiceCharge     float64                           `json:"servicecharge" bson:"servicecharge"` // ค่าบริการ
	IsCopyReceipt     bool                              `json:"iscopyreceipt" bson:"iscopyreceipt"` // ภาษี
	IsVatRegister     bool                              `json:"isvatregister" bson:"isvatregister"` // จดทะเบียนภาษี
	VatType           int8                              `json:"vattype" bson:"vattype"`             // ราคารวมภาษี หรือ ราคาไม่รวมภาษี
	VatRate           float64                           `json:"vatrate" bson:"vatrate"`             // อัตราภาษี
}

type PosSettingTimeForSale struct {
	Names *[]models.NameX `json:"names" bson:"names"`
	From  string          `json:"from" bson:"from"`
	To    string          `json:"to" bson:"to"`
}

type PosSettingSlip struct {
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

type PosSettingBranch struct {
	models.DocIdentity `bson:"inline"`
	Code               string          `json:"code" bson:"code"`
	Names              *[]models.NameX `json:"names" bson:"names"`
}

type PosSettingWarehouse struct {
	models.DocIdentity `bson:"inline"`
	Code               string               `json:"code" bson:"code"`
	Names              *[]models.NameNormal `json:"names" bson:"names"`
}

type PosSettingLocation struct {
	Code  string               `json:"code" bson:"code"`
	Names *[]models.NameNormal `json:"names" bson:"names"`
}

type SettingInfo struct {
	models.DocIdentity `bson:"inline"`
	Setting            `bson:"inline"`
}

func (SettingInfo) CollectionName() string {
	return settingCollectionName
}

type SettingData struct {
	models.HoldingCodeentity `bson:"inline"`
	SettingInfo              `bson:"inline"`
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
