package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const storefrontCollectionName = "storefront"

type Storefront struct {
	models.PartitionIdentity `bson:"inline"`
	models.Name              `bson:"inline"`
	Devices                  *[]Device `json:"devices" bson:"devices"`
}

type Device struct {
	OS   string `json:"os" bson:"os"`
	UUID string `json:"uuid" bson:"uuid"`
}

type StorefrontInfo struct {
	models.DocIdentity `bson:"inline"`
	Storefront         `bson:"inline"`
}

func (StorefrontInfo) CollectionName() string {
	return storefrontCollectionName
}

type StorefrontData struct {
	models.ShopIdentity `bson:"inline"`
	StorefrontInfo      `bson:"inline"`
}

type StorefrontDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	StorefrontData     `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (StorefrontDoc) CollectionName() string {
	return storefrontCollectionName
}

type StorefrontItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (StorefrontItemGuid) CollectionName() string {
	return storefrontCollectionName
}

type StorefrontActivity struct {
	StorefrontData      `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (StorefrontActivity) CollectionName() string {
	return storefrontCollectionName
}

type StorefrontDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (StorefrontDeleteActivity) CollectionName() string {
	return storefrontCollectionName
}
