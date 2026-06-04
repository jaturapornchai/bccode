package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const mediaCollectionName = "posMedia"

type Media struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string           `json:"code" bson:"code"`
	Description              *[]models.NameX  `json:"description" bson:"description"`
	Resources                *[]MediaResource `json:"resources" bson:"resources"`
}

type MediaResource struct {
	MediaType   int8            `json:"media_type" bson:"media_type"`
	Uri         string          `json:"uri" bson:"uri"`
	DaysOfWeek  []int8          `json:"daysofweek" bson:"daysofweek"`
	FromDate    string          `json:"from_date" bson:"from_date"`
	ToDate      string          `json:"to_date" bson:"to_date"`
	FromTime    string          `json:"from_time" bson:"from_time"`
	ToTime      string          `json:"to_time" bson:"to_time"`
	Description *[]models.NameX `json:"description" bson:"description"`
	DisplayTime int             `json:"displaytime" bson:"displaytime"`
}

type MediaInfo struct {
	models.DocIdentity `bson:"inline"`
	Media              `bson:"inline"`
}

func (MediaInfo) CollectionName() string {
	return mediaCollectionName
}

type MediaData struct {
	models.HoldingCodeentity `bson:"inline"`
	MediaInfo                `bson:"inline"`
}

type MediaDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	MediaData          `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (MediaDoc) CollectionName() string {
	return mediaCollectionName
}

type MediaItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (MediaItemGuid) CollectionName() string {
	return mediaCollectionName
}

type MediaActivity struct {
	MediaData           `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (MediaActivity) CollectionName() string {
	return mediaCollectionName
}

type MediaDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (MediaDeleteActivity) CollectionName() string {
	return mediaCollectionName
}
