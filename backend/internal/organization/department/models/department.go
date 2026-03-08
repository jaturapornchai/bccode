package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const departmentCollectionName = "organizationDepartments"

type Department struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string          `json:"code" bson:"code"`
	Names                    *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
}

type DepartmentInfo struct {
	models.DocIdentity `bson:"inline"`
	Department         `bson:"inline"`
}

func (DepartmentInfo) CollectionName() string {
	return departmentCollectionName
}

type DepartmentData struct {
	models.ShopIdentity `bson:"inline"`
	DepartmentInfo      `bson:"inline"`
}

type DepartmentDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	DepartmentData     `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (DepartmentDoc) CollectionName() string {
	return departmentCollectionName
}

type DepartmentItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (DepartmentItemGuid) CollectionName() string {
	return departmentCollectionName
}

type DepartmentActivity struct {
	DepartmentData      `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (DepartmentActivity) CollectionName() string {
	return departmentCollectionName
}

type DepartmentDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (DepartmentDeleteActivity) CollectionName() string {
	return departmentCollectionName
}
