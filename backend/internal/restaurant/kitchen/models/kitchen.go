package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const kitchenCollectionName = "restaurantKitchens"

type Kitchen struct {
	GroupNumber int             `json:"groupnumber" bson:"groupnumber"`
	Code        string          `json:"code" bson:"code"`
	Names       *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	Printers    *[]string       `json:"printers" bson:"printers"`
	Products    *[]string       `json:"products" bson:"products"`
	Zones       *[]string       `json:"zones" bson:"zones"`
}

type KitchenInfo struct {
	models.DocIdentity `bson:"inline"`
	Kitchen            `bson:"inline"`
}

func (KitchenInfo) CollectionName() string {
	return kitchenCollectionName
}

type KitchenData struct {
	models.ShopIdentity `bson:"inline"`
	KitchenInfo         `bson:"inline"`
}

type KitchenDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	KitchenData        `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
	models.LastUpdate  `bson:"inline"`
}

func (KitchenDoc) CollectionName() string {
	return kitchenCollectionName
}

type KitchenItemGuid struct {
	Code string `json:"code" bson:"code" gorm:"code"`
}

func (KitchenItemGuid) CollectionName() string {
	return kitchenCollectionName
}

type KitchenActivity struct {
	KitchenData         `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (KitchenActivity) CollectionName() string {
	return kitchenCollectionName
}

type KitchenDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (KitchenDeleteActivity) CollectionName() string {
	return kitchenCollectionName
}

type KitchenInfoResponse struct {
	Success bool        `json:"success"`
	Data    KitchenInfo `json:"data,omitempty"`
}

type KitchenPageResponse struct {
	Success    bool                          `json:"success"`
	Data       []KitchenInfo                 `json:"data,omitempty"`
	Pagination models.PaginationDataResponse `json:"pagination,omitempty"`
}

type KitchenLastActivityResponse struct {
	New    []KitchenActivity       `json:"new" `
	Remove []KitchenDeleteActivity `json:"remove"`
}

type KitchenFetchUpdateResponse struct {
	Success    bool                          `json:"success"`
	Data       KitchenLastActivityResponse   `json:"data,omitempty"`
	Pagination models.PaginationDataResponse `json:"pagination,omitempty"`
}
