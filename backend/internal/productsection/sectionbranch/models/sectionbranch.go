package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const sectionbranchCollectionName = "productSectionBranch"

type SectionBranch struct {
	models.PartitionIdentity `bson:"inline"`
	BranchCode               string    `json:"branchcode" bson:"branchcode"`
	ProductCodes             *[]string `json:"productcodes" bson:"productcodes"`
}

type SectionBranchInfo struct {
	models.DocIdentity `bson:"inline"`
	SectionBranch      `bson:"inline"`
}

func (SectionBranchInfo) CollectionName() string {
	return sectionbranchCollectionName
}

type SectionBranchData struct {
	models.ShopIdentity `bson:"inline"`
	SectionBranchInfo   `bson:"inline"`
}

type SectionBranchDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	SectionBranchData  `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (SectionBranchDoc) CollectionName() string {
	return sectionbranchCollectionName
}

type SectionBranchItemGuid struct {
	BranchCode string `json:"branchcode" bson:"branchcode"`
}

func (SectionBranchItemGuid) CollectionName() string {
	return sectionbranchCollectionName
}

type SectionBranchActivity struct {
	SectionBranchData   `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (SectionBranchActivity) CollectionName() string {
	return sectionbranchCollectionName
}

type SectionBranchDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (SectionBranchDeleteActivity) CollectionName() string {
	return sectionbranchCollectionName
}
