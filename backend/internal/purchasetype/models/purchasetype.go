package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const purchaseTypeCollectionName = "purchaseTypes"

// PurchaseType - ประเภทการจัดซื้อ (รองรับหลายภาษา)
type PurchaseType struct {
	models.PartitionIdentity `bson:"inline"`
	Code string          `json:"code" bson:"code"`
	Names *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	Descriptions *[]models.NameX `json:"descriptions" bson:"descriptions"`
}

type PurchaseTypeInfo struct {
	models.DocIdentity `bson:"inline"`
	PurchaseType  `bson:"inline"`
}

func (PurchaseTypeInfo) CollectionName() string {
	return purchaseTypeCollectionName
}

type PurchaseTypeData struct {
	models.ShopIdentity `bson:"inline"`
	PurchaseTypeInfo  `bson:"inline"`
}

type PurchaseTypeDoc struct {
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	PurchaseTypeData  `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (PurchaseTypeDoc) CollectionName() string {
	return purchaseTypeCollectionName
}

type PurchaseTypeItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (PurchaseTypeItemGuid) CollectionName() string {
	return purchaseTypeCollectionName
}

type PurchaseTypeActivity struct {
	PurchaseTypeData  `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (PurchaseTypeActivity) CollectionName() string {
	return purchaseTypeCollectionName
}

type PurchaseTypeDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (PurchaseTypeDeleteActivity) CollectionName() string {
	return purchaseTypeCollectionName
}
