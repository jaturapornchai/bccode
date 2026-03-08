package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const sectiondepartmentCollectionName = "sectionDepartment"

type SectionDepartment struct {
	models.PartitionIdentity `bson:"inline"`
	BranchCode               string    `json:"branchcode" bson:"branchcode"`
	DepartmentCode           string    `json:"departmentcode" bson:"departmentcode"`
	ProductCodes             *[]string `json:"productcodes" bson:"productcodes"`
}

type SectionDepartmentInfo struct {
	models.DocIdentity `bson:"inline"`
	SectionDepartment  `bson:"inline"`
}

func (SectionDepartmentInfo) CollectionName() string {
	return sectiondepartmentCollectionName
}

type SectionDepartmentData struct {
	models.ShopIdentity   `bson:"inline"`
	SectionDepartmentInfo `bson:"inline"`
}

type SectionDepartmentDoc struct {
	ID                    primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	SectionDepartmentData `bson:"inline"`
	models.ActivityDoc    `bson:"inline"`
}

func (SectionDepartmentDoc) CollectionName() string {
	return sectiondepartmentCollectionName
}

type SectionDepartmentItemGuid struct {
	DepartmentCode string `json:"departmentcode" bson:"departmentcode"`
}

func (SectionDepartmentItemGuid) CollectionName() string {
	return sectiondepartmentCollectionName
}

type SectionDepartmentActivity struct {
	SectionDepartmentData `bson:"inline"`
	models.ActivityTime   `bson:"inline"`
}

func (SectionDepartmentActivity) CollectionName() string {
	return sectiondepartmentCollectionName
}

type SectionDepartmentDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (SectionDepartmentDeleteActivity) CollectionName() string {
	return sectiondepartmentCollectionName
}
