package models

import (
	groupModels "smlcloudplatform/internal/debtaccount/debtorgroup/models"
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const debtorCollectionName = "debtors"

type Debtor struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string `json:"code" bson:"code"`

	PersonalType       int8            `json:"personal_type" bson:"personal_type"`
	Images             *[]Image        `json:"images" bson:"images"`
	Names              *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	PointBalance       float64         `json:"pointbalance" bson:"pointbalance"`
	PointsCode         string          `json:"points_code" bson:"points_code"`
	AddressForBilling  Address         `json:"addressforbilling" bson:"addressforbilling"`
	AddressForShipping *[]Address      `json:"addressforshipping" bson:"addressforshipping"`
	TaxId              string          `json:"tax_id" bson:"tax_id"`
	Email              string          `json:"email" bson:"email"`
	CustomerType       int             `json:"customer_type" bson:"customer_type"`
	BranchNumber       string          `json:"branch_number" bson:"branch_number"`
	FundCode           string          `json:"fund_code" bson:"fund_code"`
	CreditDay          int             `json:"creditday" bson:"creditday"`
	IsMember           bool            `json:"ismember" bson:"ismember"`
	GroupGUIDs         *[]string       `json:"groups" bson:"groups"`
	Auth               DebtorAuth      `json:"auth" bson:"auth"`
	DebtorLine         DebtorLine      `json:"line" bson:"line"`
	PriceLevel         string          `json:"pricelevel" bson:"pricelevel"`
}

type DebtorLine struct {
	LineID    string    `json:"line_id" bson:"line_id"`
	LineUID   string    `json:"line_uid" bson:"line_uid"`
	ClientIDs *[]string `json:"clientids" bson:"clientids"`
}

type DebtorAuth struct {
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
}

type Address struct {
	GUID            string          `json:"guid" bson:"guid"`
	Address         *[]string       `json:"address" bson:"address"`
	CountryCode     string          `json:"country_code" bson:"country_code"`
	ProvinceCode    string          `json:"province_code" bson:"province_code"`
	DistrictCode    string          `json:"district_code" bson:"district_code"`
	SubDistrictCode string          `json:"sub_district_code" bson:"sub_district_code"`
	ZipCode         string          `json:"zip_code" bson:"zip_code"`
	ContactNames    *[]models.NameX `json:"contactnames" bson:"contactnames"`
	PhonePrimary    string          `json:"phone_primary" bson:"phone_primary"`
	PhoneSecondary  string          `json:"phone_secondary" bson:"phone_secondary"`
	Latitude        float64         `json:"latitude" bson:"latitude"`
	Longitude       float64         `json:"longitude" bson:"longitude"`
}

type Image struct {
	XOrder int    `json:"xorder" bson:"xorder"`
	URI    string `json:"uri" bson:"uri"`
}

type DebtorRequest struct {
	Debtor
	Groups []string `json:"groups"`
}

type DebtorInfo struct {
	models.HoldingCodeentity `bson:"inline"`
	models.DocIdentity       `bson:"inline"`
	Debtor                   `bson:"inline"`
	Groups                   *[]groupModels.DebtorGroupInfo `json:"groups" bson:"-"`
}

func (DebtorInfo) CollectionName() string {
	return debtorCollectionName
}

type DebtorData struct {
	models.HoldingCodeentity `bson:"inline"`
	DebtorInfo               `bson:"inline"`
}

type DebtorDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	DebtorData         `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (DebtorDoc) CollectionName() string {
	return debtorCollectionName
}

type DebtorItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (DebtorItemGuid) CollectionName() string {
	return debtorCollectionName
}

type DebtorActivity struct {
	DebtorData          `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (DebtorActivity) CollectionName() string {
	return debtorCollectionName
}

type DebtorDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (DebtorDeleteActivity) CollectionName() string {
	return debtorCollectionName
}
