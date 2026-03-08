package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const shopCollectionName = "shops"

type ShopRequest struct {
	Shop
	BusinessType ShopBusinessType `json:"businesstype"`
}

type ShopBusinessType struct {
	Code  string          `json:"code" bson:"code"`
	Names *[]models.NameX `json:"names" bson:"names"`
}

type Shop struct {
	ProfilePicture       string         `json:"profilepicture" bson:"profilepicture"`
	Name1                string         `json:"name1" bson:"name1"`
	Names                []models.NameX `json:"names" bson:"names"`
	Telephone            string         `json:"telephone" bson:"telephone"`
	BranchCode           string         `json:"branchcode" bson:"branchcode"`
	IsMainShop           bool           `json:"ismainshop" bson:"ismainshop"`
	PosProductCenterType int8           `json:"posproductcentertype" bson:"posproductcentertype"`
	ProductCenterType    int8           `json:"productcentertype" bson:"productcentertype"`
	DebtorCenterType     int8           `json:"debtorcentertype" bson:"debtorcentertype"`
	MainShopId           string         `json:"mainshopid" bson:"mainshopid"`
	Address              []models.NameX `json:"address" bson:"address"`
	Images               []ShopImage    `json:"images" bson:"images"`
	Logo                 string         `json:"logo" bson:"logo"`
	Settings             ShopSettings   `json:"settings" bson:"settings"`
	IsBcMember		  bool           `json:"isbcmember" bson:"isbcmember"`
	ApiKey               string         `json:"apikey" bson:"apikey"`
	PromptShopInfo       string         `json:"promptshopinfo" bson:"promptshopinfo"`
}

type ShopImage struct {
	XOrder int    `json:"xorder" bson:"xorder"`
	URI    string `json:"uri" bson:"uri"`
}

type ShopSettings struct {
	TaxID               string           `json:"taxid" bson:"taxid"`
	EmailOwners         []string         `json:"emailowners" bson:"emailowners"`
	EmailStaffs         []string         `json:"emailstaffs" bson:"emailstaffs"`
	Latitude            float64          `json:"latitude" bson:"latitude"`
	Longitude           float64          `json:"longitude" bson:"longitude"`
	IsUseBranch         bool             `json:"isusebranch" bson:"isusebranch"`
	IsUseDepartment     bool             `json:"isusedepartment" bson:"isusedepartment"`
	UseBuddhistCalendar bool             `json:"usebuddhistcalendar" bson:"usebuddhistcalendar"` // true = พ.ศ., false = ค.ศ.
	VatTypeSale         float64          `json:"vattypesale" bson:"vattypesale"`
	VateTypePurchase    int              `json:"vattypepurchase" bson:"vattypepurchase"`
	InquiryTypeSale     int              `json:"inquirytypesale" bson:"inquirytypesale"`
	InquiryTypePurchase int              `json:"inquirytypepurchase" bson:"inquirytypepurchase"`
	LanguageConfigs     []LanguageConfig `json:"languageconfigs" bson:"languageconfigs"`
	BaseCurrency        string           `json:"base_currency" bson:"base_currency"` // สกุลเงินหลักของบริษัท (เช่น THB, USD)
}

type LanguageConfig struct {
	Code           string `json:"code" bson:"code"`
	CodeTranslator string `json:"codetranslator" bson:"codetranslator"`
	Name           string `json:"name" bson:"name"`
	IsUse          bool   `json:"isuse" bson:"isuse"`
	IsDefault      bool   `json:"isdefault" bson:"isdefault"`
}

type ShopInfo struct {
	models.DocIdentity `bson:"inline"`
	Shop               `bson:"inline"`
}

func (ShopInfo) CollectionName() string {
	return shopCollectionName
}

type ShopDoc struct {
	ID                 primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	ShopInfo           `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (ShopDoc) CollectionName() string {
	return shopCollectionName
}
