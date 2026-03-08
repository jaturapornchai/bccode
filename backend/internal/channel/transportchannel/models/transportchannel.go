package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const transportchannelCollectionName = "transportChannel"

type TransportChannel struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string `json:"code" bson:"code" validate:"required,min=1"`
	Name                     string `json:"name" bson:"name" validate:"required,min=1"`
	ImageUri                 string `json:"imageuri" bson:"imageuri"`
}

type TransportChannelInfo struct {
	models.DocIdentity `bson:"inline"`
	TransportChannel   `bson:"inline"`
}

func (TransportChannelInfo) CollectionName() string {
	return transportchannelCollectionName
}

type TransportChannelData struct {
	models.ShopIdentity  `bson:"inline"`
	TransportChannelInfo `bson:"inline"`
}

type TransportChannelDoc struct {
	ID                   primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TransportChannelData `bson:"inline"`
	models.ActivityDoc   `bson:"inline"`
}

func (TransportChannelDoc) CollectionName() string {
	return transportchannelCollectionName
}

type TransportChannelItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (TransportChannelItemGuid) CollectionName() string {
	return transportchannelCollectionName
}

type TransportChannelActivity struct {
	TransportChannelData `bson:"inline"`
	models.ActivityTime  `bson:"inline"`
}

func (TransportChannelActivity) CollectionName() string {
	return transportchannelCollectionName
}

type TransportChannelDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (TransportChannelDeleteActivity) CollectionName() string {
	return transportchannelCollectionName
}
