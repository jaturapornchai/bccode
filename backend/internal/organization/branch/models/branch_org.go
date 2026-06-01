package models

import (
	"time"

	common "smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BranchOrgDoc struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ShopID      string             `json:"shopid" bson:"shopid"`
	GuidFixed   string             `json:"guid_fixed" bson:"guid_fixed"`
	CompanyGuid string             `json:"company_guid" bson:"company_guid"`
	Code        string             `json:"code" bson:"code"`
	Names       common.JSONB       `json:"names" bson:"names"`
	IsActive    bool               `json:"is_active" bson:"is_active"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
	DeletedAt   *time.Time         `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	CreatedBy   string             `json:"createdby,omitempty" bson:"createdby,omitempty"`
	UpdatedBy   string             `json:"updatedby,omitempty" bson:"updatedby,omitempty"`
	DeletedBy   string             `json:"deletedby,omitempty" bson:"deletedby,omitempty"`
}

func (BranchOrgDoc) CollectionName() string {
	return branchCollectionName
}
