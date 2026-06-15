package models

import (
	"time"

	common "smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BranchOrgDoc struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	HoldingCode string             `json:"holdingcode" bson:"holdingcode"`
	GuidFixed   string             `json:"guidfixed" bson:"guidfixed"`
	CompanyGuid string             `json:"companyguid" bson:"companyguid"`
	Code        string             `json:"code" bson:"code"`
	Names       common.JSONB       `json:"names" bson:"names"`
	IsActive    bool               `json:"isactive" bson:"isactive"`
	CreatedAt   time.Time          `json:"createdat" bson:"createdat"`
	UpdatedAt   time.Time          `json:"updatedat" bson:"updatedat"`
	DeletedAt   *time.Time         `json:"deletedat,omitempty" bson:"deletedat,omitempty"`
	CreatedBy   string             `json:"createdby,omitempty" bson:"createdby,omitempty"`
	UpdatedBy   string             `json:"updatedby,omitempty" bson:"updatedby,omitempty"`
	DeletedBy   string             `json:"deletedby,omitempty" bson:"deletedby,omitempty"`
}

func (BranchOrgDoc) CollectionName() string {
	return branchCollectionName
}
