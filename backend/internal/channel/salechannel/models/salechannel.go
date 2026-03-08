package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const salechannelCollectionName = "saleChannel"

type SaleChannel struct {
	Code        string  `json:"code" bson:"code" validate:"required,min=1"`
	Name        string  `json:"name" bson:"name" validate:"required,min=1"`
	GP          float64 `json:"gp" bson:"gp"`
	GPType      int8    `json:"gptype" bson:"gptype"`
	PriceNumber int8    `json:"price" bson:"price"`
	ImageUri    string  `json:"imageuri" bson:"imageuri"`
}

type SaleChannelInfo struct {
	models.DocIdentity `bson:"inline"`
	SaleChannel        `bson:"inline"`
}

func (SaleChannelInfo) CollectionName() string {
	return salechannelCollectionName
}

type SaleChannelData struct {
	models.PartitionIdentity `bson:"inline"`
	models.ShopIdentity      `bson:"inline"`
	SaleChannelInfo          `bson:"inline"`
}

type SaleChannelDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	SaleChannelData    `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (SaleChannelDoc) CollectionName() string {
	return salechannelCollectionName
}

type SaleChannelItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (SaleChannelItemGuid) CollectionName() string {
	return salechannelCollectionName
}

type SaleChannelActivity struct {
	SaleChannelData     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (SaleChannelActivity) CollectionName() string {
	return salechannelCollectionName
}

type SaleChannelDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (SaleChannelDeleteActivity) CollectionName() string {
	return salechannelCollectionName
}
