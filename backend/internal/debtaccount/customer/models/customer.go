package models

import (
	groupModels "smlcloudplatform/internal/debtaccount/customergroup/models"
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const customerCollectionName = "customers"

type Customer struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string          `json:"code" bson:"code"`
	PersonalType             int8            `json:"personaltype" bson:"personaltype"`
	Images                   *[]Image        `json:"images" bson:"images"`
	Names                    *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`

	AddressForBilling  Address    `json:"addressforbilling" bson:"addressforbilling"`
	AddressForShipping *[]Address `json:"addressforshipping" bson:"addressforshipping"`
	TaxId              string     `json:"taxid" bson:"taxid"`
	Email              string     `json:"email" bson:"email"`
	CustomerType       int        `json:"customertype" bson:"customertype"`
	BranchNumber       string     `json:"branchnumber" bson:"branchnumber"`
	IsCreditor         bool       `json:"iscreditor" bson:"iscreditor"`
	IsDebtor           bool       `json:"isdebtor" bson:"isdebtor"`

	FundCode   string    `json:"fundcode" bson:"fundcode"`
	CreditDay  int       `json:"creditday" bson:"creditday"`
	GroupGUIDs *[]string `json:"-" bson:"groups"`
}

type Address struct {
	GUID            string          `json:"guid" bson:"guid"`
	Address         *[]string       `json:"address" bson:"address"`
	CountryCode     string          `json:"countrycode" bson:"countrycode"`
	ProvinceCode    string          `json:"provincecode" bson:"provincecode"`
	DistrictCode    string          `json:"districtcode" bson:"districtcode"`
	SubDistrictCode string          `json:"subdistrictcode" bson:"subdistrictcode"`
	ZipCode         string          `json:"zipcode" bson:"zipcode"`
	ContactNames    *[]models.NameX `json:"contactnames" bson:"contactnames"`
	PhonePrimary    string          `json:"phoneprimary" bson:"phoneprimary"`
	PhoneSecondary  string          `json:"phonesecondary" bson:"phonesecondary"`
	Latitude        float64         `json:"latitude" bson:"latitude"`
	Longitude       float64         `json:"longitude" bson:"longitude"`
}

type Image struct {
	XOrder int    `json:"xorder" bson:"xorder"`
	URI    string `json:"uri" bson:"uri"`
}

type CustomerRequest struct {
	Customer
	Groups []string `json:"groups"`
}

type CustomerInfo struct {
	models.DocIdentity `bson:"inline"`
	Customer           `bson:"inline"`
	Groups             *[]groupModels.CustomerGroupInfo `json:"groups" bson:"-"`
}

func (CustomerInfo) CollectionName() string {
	return customerCollectionName
}

type CustomerData struct {
	models.ShopIdentity `bson:"inline"`
	CustomerInfo        `bson:"inline"`
}

type CustomerDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CustomerData       `bson:"inline"`
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
	CustomerData        `bson:"inline"`
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
