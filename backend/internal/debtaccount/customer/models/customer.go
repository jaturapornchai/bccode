package models

import (
	groupModels "smlcloudplatform/internal/debtaccount/customergroup/models"
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const customerCollectionName = "customers"

type Customer struct {
	models.PartitionIdentity `bson:"inline"`
	Code string          `json:"code" bson:"code"`
	PersonalType int8            `json:"personal_type" bson:"personal_type"`
	Images *[]Image        `json:"images" bson:"images"`
	Names *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`

	AddressForBilling Address    `json:"addressforbilling" bson:"addressforbilling"`
	AddressForShipping *[]Address `json:"addressforshipping" bson:"addressforshipping"`
	TaxId string     `json:"tax_id" bson:"tax_id"`
	Email string     `json:"email" bson:"email"`
	CustomerType int        `json:"customer_type" bson:"customer_type"`
	BranchNumber string     `json:"branch_number" bson:"branch_number"`
	IsCreditor bool       `json:"iscreditor" bson:"iscreditor"`
	IsDebtor bool       `json:"isdebtor" bson:"isdebtor"`

	FundCode string    `json:"fund_code" bson:"fund_code"`
	CreditDay int       `json:"creditday" bson:"creditday"`
	GroupGUIDs *[]string `json:"-" bson:"groups"`
}

type Address struct {
	GUID string          `json:"guid" bson:"guid"`
	Address *[]string       `json:"address" bson:"address"`
	CountryCode string          `json:"country_code" bson:"country_code"`
	ProvinceCode string          `json:"province_code" bson:"province_code"`
	DistrictCode string          `json:"district_code" bson:"district_code"`
	SubDistrictCode string          `json:"sub_district_code" bson:"sub_district_code"`
	ZipCode string          `json:"zip_code" bson:"zip_code"`
	ContactNames *[]models.NameX `json:"contactnames" bson:"contactnames"`
	PhonePrimary string          `json:"phone_primary" bson:"phone_primary"`
	PhoneSecondary string          `json:"phone_secondary" bson:"phone_secondary"`
	Latitude float64         `json:"latitude" bson:"latitude"`
	Longitude float64         `json:"longitude" bson:"longitude"`
}

type Image struct {
	XOrder int    `json:"xorder" bson:"xorder"`
	URI string `json:"uri" bson:"uri"`
}

type CustomerRequest struct {
	Customer
	Groups []string `json:"groups"`
}

type CustomerInfo struct {
	models.DocIdentity `bson:"inline"`
	Customer  `bson:"inline"`
	Groups *[]groupModels.CustomerGroupInfo `json:"groups" bson:"-"`
}

func (CustomerInfo) CollectionName() string {
	return customerCollectionName
}

type CustomerData struct {
	models.ShopIdentity `bson:"inline"`
	CustomerInfo  `bson:"inline"`
}

type CustomerDoc struct {
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CustomerData  `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (CustomerDoc) CollectionName() string {
	return customerCollectionName
}

type CustomerItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (CustomerItemGuid) CollectionName() string {
	return customerCollectionName
}

type CustomerActivity struct {
	CustomerData  `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (CustomerActivity) CollectionName() string {
	return customerCollectionName
}

type CustomerDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (CustomerDeleteActivity) CollectionName() string {
	return customerCollectionName
}
