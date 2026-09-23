package models

import (
	"smlcloudplatform/internal/models"
)

type ShopRequest struct {
	Shop
	BusinessType ShopBusinessType `json:"businesstype"`
}

type ShopBusinessType struct {
	Code  string          `json:"code"`
	Names *[]models.NameX `json:"names"`
}

type Shop struct {
	HoldingCode          string         `json:"holdingcode"`
	IsActive             bool           `json:"isactive"`
	ProfilePicture       string         `json:"profilepicture"`
	Name1                string         `json:"name1"`
	Names                []models.NameX `json:"names"`
	Telephone            string         `json:"telephone"`
	BranchCode           string         `json:"branchcode"`
	IsMainShop           bool           `json:"ismainshop"`
	PosProductCenterType int8           `json:"posproductcentertype"`
	ProductCenterType    int8           `json:"productcentertype"`
	DebtorCenterType     int8           `json:"debtorcentertype"`
	MainHoldingCode      string         `json:"mainholdingcode"`
	Address              []models.NameX `json:"address"`
	Images               []ShopImage    `json:"images"`
	Logo                 string         `json:"logo"`
	Settings             ShopSettings   `json:"settings"`
	IsBcMember           bool           `json:"isbcmember"`
	ApiKey               string         `json:"apikey"`
	PromptShopInfo       string         `json:"promptshopinfo"`
}

type ShopImage struct {
	XOrder int    `json:"xorder"`
	URI    string `json:"uri"`
}

type ShopSettings struct {
	TaxID                 string           `json:"taxid"`
	CompanyRegistrationNo string           `json:"companyregistrationno"`
	CountryCode           string           `json:"countrycode"`
	Language              string           `json:"language"`
	EmailOwners           []string         `json:"emailowners"`
	EmailStaffs           []string         `json:"emailstaffs"`
	Latitude              float64          `json:"latitude"`
	Longitude             float64          `json:"longitude"`
	IsUseBranch           bool             `json:"isusebranch"`
	IsUseDepartment       bool             `json:"isusedepartment"`
	UseBuddhistCalendar   bool             `json:"usebuddhistcalendar"` // true = พ.ศ., false = ค.ศ.
	IsVatRegistered       bool             `json:"isvatregistered"`
	VatRate               float64          `json:"vatrate"`
	VatTypeSale           float64          `json:"vattypesale"`
	VateTypePurchase      int              `json:"vattypepurchase"`
	InquiryTypeSale       int              `json:"inquirytypesale"`
	InquiryTypePurchase   int              `json:"inquirytypepurchase"`
	LanguageConfigs       []LanguageConfig `json:"languageconfigs"`
	BaseCurrency          string           `json:"basecurrency"` // สกุลเงินหลักของบริษัท (เช่น THB, USD)
	CurrencyCodes         []string         `json:"currencycodes"`
	Timezone              string           `json:"timezone"`
	TimezoneLabel         string           `json:"timezonelabel"`
	TimezoneOffset        string           `json:"timezoneoffset"`
	DateFormat            string           `json:"dateformat"`
	DecimalQuantity       int8             `json:"decimalquantity"`
	DecimalPrice          int8             `json:"decimalprice"`
	DecimalDocument       int8             `json:"decimaldocument"`
}

type LanguageConfig struct {
	Code           string `json:"code"`
	CodeTranslator string `json:"codetranslator"`
	Name           string `json:"name"`
	IsUse          bool   `json:"isuse"`
	IsDefault      bool   `json:"isdefault"`
}

type ShopInfo struct {
	models.DocIdentity
	Shop
}

type ShopDoc struct {
	HoldingUID string `json:"holdinguid"`
	IsDeleted  bool   `json:"isdeleted"`
	ShopInfo
	models.ActivityDoc
}
