package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const colorCollectionName = "colors"

type Color struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string          `json:"code" bson:"code"`
	Names                    *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	ColorSelect              string          `json:"colorselect" bson:"colorselect"`
	ColorSystem              string          `json:"colorsystem" bson:"colorsystem"`
	ColorHex                 string          `json:"colorhex" bson:"colorhex"`
	ColorSelectHex           string          `json:"colorselecthex" bson:"colorselecthex"`
	ColorSystemHex           string          `json:"colorsystemhex" bson:"colorsystemhex"`
}

type ColorInfo struct {
	models.DocIdentity `bson:"inline"`
	Color              `bson:"inline"`
}

func (ColorInfo) CollectionName() string {
	return colorCollectionName
}

type ColorData struct {
	models.ShopIdentity `bson:"inline"`
	ColorInfo           `bson:"inline"`
}

type ColorDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ColorData          `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (ColorDoc) CollectionName() string {
	return colorCollectionName
}

type ColorItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (ColorItemGuid) CollectionName() string {
	return colorCollectionName
}

type ColorActivity struct {
	ColorData           `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ColorActivity) CollectionName() string {
	return colorCollectionName
}

type ColorDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ColorDeleteActivity) CollectionName() string {
	return colorCollectionName
}
