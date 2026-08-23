package models

import (
	"strings"
	"time"

	common "smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const companyCollectionName = "organizationcompanies"

type Company struct {
	Code     string       `json:"code" bson:"code"`
	Names    common.JSONB `json:"names" bson:"names"`
	TaxID    string       `json:"taxid" bson:"taxid"`
	LogoURI  string       `json:"logouri" bson:"logouri"`
	IsActive bool         `json:"isactive" bson:"isactive"`
}

type CompanyDoc struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Version     int64              `json:"__v" bson:"__v"`
	HoldingCode string             `json:"holdingcode" bson:"holdingcode"`
	HoldingUID  string             `json:"holdinguid" bson:"holdinguid"`
	GuidFixed   string             `json:"guidfixed" bson:"guidfixed"`
	CompanyUID  string             `json:"companyuid" bson:"companyuid"`
	IsDeleted   bool               `json:"isdeleted" bson:"isdeleted"`
	Company     `bson:"inline"`
	CreatedAt   time.Time  `json:"createdat" bson:"createdat"`
	UpdatedAt   time.Time  `json:"updatedat" bson:"updatedat"`
	DeletedAt   *time.Time `json:"deletedat,omitempty" bson:"deletedat,omitempty"`
	CreatedBy   string     `json:"createdby,omitempty" bson:"createdby,omitempty"`
	UpdatedBy   string     `json:"updatedby,omitempty" bson:"updatedby,omitempty"`
	DeletedBy   string     `json:"deletedby,omitempty" bson:"deletedby,omitempty"`
}

func (CompanyDoc) CollectionName() string {
	return companyCollectionName
}

func NormalizeCompanyCode(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}
