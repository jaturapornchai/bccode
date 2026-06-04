package models

import (
	"strings"
	"time"

	common "smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const companyCollectionName = "organizationCompanies"

type Company struct {
	Code     string       `json:"code" bson:"code"`
	Names    common.JSONB `json:"names" bson:"names"`
	TaxID    string       `json:"tax_id" bson:"tax_id"`
	IsActive bool         `json:"is_active" bson:"is_active"`
}

type CompanyDoc struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	HoldingCode string             `json:"holding_code" bson:"holding_code"`
	GuidFixed   string             `json:"guid_fixed" bson:"guid_fixed"`
	Company     `bson:"inline"`
	CreatedAt   time.Time  `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" bson:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
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
