package models

import (
	groupModels "smlcloudplatform/internal/debtaccount/creditorgroup/models"
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const creditorCollectionName = "creditors"

type Creditor struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string          `json:"code" bson:"code"`
	PersonalType             int8            `json:"personaltype" bson:"personaltype"`
	Images                   *[]Image        `json:"images" bson:"images"`
	Names                    *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`

	AddressForBilling  Address      `json:"addressforbilling" bson:"addressforbilling"`
	AddressForShipping *[]Address   `json:"addressforshipping" bson:"addressforshipping"`
	TaxId              string       `json:"taxid" bson:"taxid"`
	Email              string       `json:"email" bson:"email"`
	CustomerType       int          `json:"customertype" bson:"customertype"`
	BranchNumber       string       `json:"branchnumber" bson:"branchnumber"`
	FundCode           string       `json:"fundcode" bson:"fundcode"`
	CreditDay          int          `json:"creditday" bson:"creditday"`
	IsMember           bool         `json:"ismember" bson:"ismember"`
	GroupGUIDs         *[]string    `json:"-" bson:"groups"`
	Auth               CreditorAuth `json:"auth" bson:"auth"`
}

type CreditorAuth struct {
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
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

type CreditorRequest struct {
	Creditor
	Groups []string `json:"groups"`
}

type CreditorInfo struct {
	models.DocIdentity `bson:"inline"`
	Creditor           `bson:"inline"`
	Groups             *[]groupModels.CreditorGroupInfo `json:"groups" bson:"-"`
}

func (CreditorInfo) CollectionName() string {
	return creditorCollectionName
}

type CreditorData struct {
	models.ShopIdentity `bson:"inline"`
	CreditorInfo        `bson:"inline"`
}

type CreditorDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CreditorData       `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (CreditorDoc) CollectionName() string {
	return creditorCollectionName
}

type CreditorItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (CreditorItemGuid) CollectionName() string {
	return creditorCollectionName
}

type CreditorActivity struct {
	CreditorData        `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (CreditorActivity) CollectionName() string {
	return creditorCollectionName
}

type CreditorDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (CreditorDeleteActivity) CollectionName() string {
	return creditorCollectionName
}
