package models

import (
	"smlcloudplatform/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const notifierCollectionName = "notifier"

type Notifier struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string    `json:"code" bson:"code" validate:"required,min=1,max=255"`
	Title                    string    `json:"title" bson:"title" validate:"required,min=1,max=100"`
	Message                  string    `json:"message" bson:"message" validate:"required,min=1,max=255"`
	NotifiedAt               time.Time `json:"notifiedat" bson:"notifiedat"`
	AccceptedAt              time.Time `json:"accceptedat" bson:"accceptedat"`
	AccceptedBy              string    `json:"accceptedby" bson:"accceptedby"`
	RejectedAt               time.Time `json:"rejectedat" bson:"rejectedat"`
	RejectedBy               string    `json:"rejectedby" bson:"rejectedby"`
}

type NotifierInfo struct {
	models.DocIdentity `bson:"inline"`
	Notifier           `bson:"inline"`
}

func (NotifierInfo) CollectionName() string {
	return notifierCollectionName
}

type NotifierData struct {
	models.ShopIdentity `bson:"inline"`
	NotifierInfo        `bson:"inline"`
}

type NotifierDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	NotifierData       `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (NotifierDoc) CollectionName() string {
	return notifierCollectionName
}

type NotifierItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (NotifierItemGuid) CollectionName() string {
	return notifierCollectionName
}
