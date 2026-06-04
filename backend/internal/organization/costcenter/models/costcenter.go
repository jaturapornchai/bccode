package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const costCenterCollectionName = "organizationCostCenters"

type CostCenter struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string          `json:"code" bson:"code"`
	Names                    *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
}

type CostCenterInfo struct {
	models.DocIdentity `bson:"inline"`
	CostCenter         `bson:"inline"`
}

func (CostCenterInfo) CollectionName() string {
	return costCenterCollectionName
}

type CostCenterData struct {
	models.HoldingCodeentity `bson:"inline"`
	CostCenterInfo           `bson:"inline"`
}

type CostCenterDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CostCenterData     `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (CostCenterDoc) CollectionName() string {
	return costCenterCollectionName
}

type CostCenterItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (CostCenterItemGuid) CollectionName() string {
	return costCenterCollectionName
}

type CostCenterActivity struct {
	CostCenterData      `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (CostCenterActivity) CollectionName() string {
	return costCenterCollectionName
}

type CostCenterDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (CostCenterDeleteActivity) CollectionName() string {
	return costCenterCollectionName
}
