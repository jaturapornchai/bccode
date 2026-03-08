package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const shopSettingsCollectionName = "restaurantSettings"

type RestaurantSettings struct {
	Code string `json:"code" bson:"code"`
	Body string `json:"body" bson:"body"`
}

type RestaurantSettingsInfo struct {
	models.DocIdentity `bson:"inline"`
	RestaurantSettings `bson:"inline"`
}

func (RestaurantSettingsInfo) CollectionName() string {
	return shopSettingsCollectionName
}

type RestaurantSettingsData struct {
	models.ShopIdentity    `bson:"inline"`
	RestaurantSettingsInfo `bson:"inline"`
}

type RestaurantSettingsDoc struct {
	ID                     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	RestaurantSettingsData `bson:"inline"`
	models.ActivityDoc     `bson:"inline"`
	models.LastUpdate      `bson:"inline"`
}

func (RestaurantSettingsDoc) CollectionName() string {
	return shopSettingsCollectionName
}

type RestaurantSettingsItemGuid struct {
	Code string `json:"code" bson:"code" gorm:"code"`
}

func (RestaurantSettingsItemGuid) CollectionName() string {
	return shopSettingsCollectionName
}

type RestaurantSettingsActivity struct {
	RestaurantSettingsData `bson:"inline"`
	models.ActivityTime    `bson:"inline"`
}

func (RestaurantSettingsActivity) CollectionName() string {
	return shopSettingsCollectionName
}

type RestaurantSettingsDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (RestaurantSettingsDeleteActivity) CollectionName() string {
	return shopSettingsCollectionName
}

type RestaurantSettingsInfoResponse struct {
	Success bool                   `json:"success"`
	Data    RestaurantSettingsInfo `json:"data,omitempty"`
}

type RestaurantSettingsPageResponse struct {
	Success    bool                          `json:"success"`
	Data       []RestaurantSettingsInfo      `json:"data,omitempty"`
	Pagination models.PaginationDataResponse `json:"pagination,omitempty"`
}

type RestaurantSettingsLastActivityResponse struct {
	New    []RestaurantSettingsActivity       `json:"new" `
	Remove []RestaurantSettingsDeleteActivity `json:"remove"`
}

type RestaurantSettingsFetchUpdateResponse struct {
	Success    bool                                   `json:"success"`
	Data       RestaurantSettingsLastActivityResponse `json:"data,omitempty"`
	Pagination models.PaginationDataResponse          `json:"pagination,omitempty"`
}
