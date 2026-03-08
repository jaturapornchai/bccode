package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const ordertypeCollectionName = "productOrderTypes"

type OrderType struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string            `json:"code" bson:"code"`
	Names                    *[]models.NameX   `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	Prices                   *[]OrderTypePrice `json:"prices" bson:"prices" validate:"required,min=1,unique=Type,dive"`
	Remarks                  []*[]models.NameX `json:"remarks" bson:"remarks"`
}

type OrderTypePrice struct {
	Type  int8    `json:"type" bson:"type"`
	Price float64 `json:"price" bson:"price"`
}

type OrderTypeInfo struct {
	models.DocIdentity `bson:"inline"`
	OrderType          `bson:"inline"`
}

func (OrderTypeInfo) CollectionName() string {
	return ordertypeCollectionName
}

type OrderTypeData struct {
	models.ShopIdentity `bson:"inline"`
	OrderTypeInfo       `bson:"inline"`
}

type OrderTypeDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	OrderTypeData      `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (OrderTypeDoc) CollectionName() string {
	return ordertypeCollectionName
}

type OrderTypeItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (OrderTypeItemGuid) CollectionName() string {
	return ordertypeCollectionName
}

type OrderTypeActivity struct {
	OrderTypeData       `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (OrderTypeActivity) CollectionName() string {
	return ordertypeCollectionName
}

type OrderTypeDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (OrderTypeDeleteActivity) CollectionName() string {
	return ordertypeCollectionName
}
