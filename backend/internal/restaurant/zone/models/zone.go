package models

import (
	"smlcloudplatform/internal/models"
	printerModel "smlcloudplatform/internal/restaurant/printer/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const shopZoneCollectionName = "restaurantZones"

type Zone struct {
	GroupNumber int                   `json:"groupnumber" bson:"groupnumber"`
	Code        string                `json:"code" bson:"code"`
	Names       *[]models.NameX       `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	Printer     *printerModel.Printer `json:"printer" bson:"printer"`
}

type ZoneInfo struct {
	models.DocIdentity `bson:"inline"`
	Zone               `bson:"inline"`
}

func (ZoneInfo) CollectionName() string {
	return shopZoneCollectionName
}

type ZoneData struct {
	models.ShopIdentity `bson:"inline"`
	ZoneInfo            `bson:"inline"`
}

type ZoneDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ZoneData           `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
	models.LastUpdate  `bson:"inline"`
}

func (ZoneDoc) CollectionName() string {
	return shopZoneCollectionName
}

type ZoneItemGuid struct {
	Code string `json:"code" bson:"code" gorm:"code"`
}

func (ZoneItemGuid) CollectionName() string {
	return shopZoneCollectionName
}

type ZoneActivity struct {
	ZoneData            `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ZoneActivity) CollectionName() string {
	return shopZoneCollectionName
}

type ZoneDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ZoneDeleteActivity) CollectionName() string {
	return shopZoneCollectionName
}

type ZoneInfoResponse struct {
	Success bool     `json:"success"`
	Data    ZoneInfo `json:"data,omitempty"`
}

type ZonePageResponse struct {
	Success    bool                          `json:"success"`
	Data       []ZoneInfo                    `json:"data,omitempty"`
	Pagination models.PaginationDataResponse `json:"pagination,omitempty"`
}

type ZoneLastActivityResponse struct {
	New    []ZoneActivity       `json:"new" `
	Remove []ZoneDeleteActivity `json:"remove"`
}

type ZoneFetchUpdateResponse struct {
	Success    bool                          `json:"success"`
	Data       ZoneLastActivityResponse      `json:"data,omitempty"`
	Pagination models.PaginationDataResponse `json:"pagination,omitempty"`
}
