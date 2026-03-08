package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const sectionbusinesstypeCollectionName = "sectionBusinessType"

type SectionBusinessType struct {
	models.PartitionIdentity `bson:"inline"`
	BusinessTypeCode         string    `json:"businesstypecode" bson:"businesstypecode"`
	ProductCodes             *[]string `json:"productcodes" bson:"productcodes"`
}

type SectionBusinessTypeInfo struct {
	models.DocIdentity  `bson:"inline"`
	SectionBusinessType `bson:"inline"`
}

func (SectionBusinessTypeInfo) CollectionName() string {
	return sectionbusinesstypeCollectionName
}

type SectionBusinessTypeData struct {
	models.ShopIdentity     `bson:"inline"`
	SectionBusinessTypeInfo `bson:"inline"`
}

type SectionBusinessTypeDoc struct {
	ID                      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	SectionBusinessTypeData `bson:"inline"`
	models.ActivityDoc      `bson:"inline"`
}

func (SectionBusinessTypeDoc) CollectionName() string {
	return sectionbusinesstypeCollectionName
}

type SectionBusinessTypeItemGuid struct {
	BusinessTypeCode string `json:"businesstypecode" bson:"businesstypecode"`
}

func (SectionBusinessTypeItemGuid) CollectionName() string {
	return sectionbusinesstypeCollectionName
}

type SectionBusinessTypeActivity struct {
	SectionBusinessTypeData `bson:"inline"`
	models.ActivityTime     `bson:"inline"`
}

func (SectionBusinessTypeActivity) CollectionName() string {
	return sectionbusinesstypeCollectionName
}

type SectionBusinessTypeDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (SectionBusinessTypeDeleteActivity) CollectionName() string {
	return sectionbusinesstypeCollectionName
}
