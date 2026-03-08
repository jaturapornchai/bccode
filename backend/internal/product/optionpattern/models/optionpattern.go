package models

import (
	"smlcloudplatform/internal/models"

	optionModel "smlcloudplatform/internal/product/option/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const optionpatternCollectionName = "optionPattern"

type OptionPattern struct {
	models.PartitionIdentity `bson:"inline"`

	PatternCode          string `json:"patterncode" bson:"patterncode"`
	models.Name          `bson:"inline"`
	Names                *[]models.NameX        `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	OptionPatternDetails *[]OptionPatternDetail `json:"optionpatterndetails" bson:"optionpatterndetails"`
	ColorCode            string                 `json:"colorcode" bson:"colorcode"`
}

type OptionPatternDetail struct {
	XOrder              int8   `json:"xorder" bson:"xorder"`
	OptionCode          string `json:"optioncode" bson:"optioncode"`
	*optionModel.Option `bson:"inline"`
}

type OptionPatternInfo struct {
	models.DocIdentity `bson:"inline"`
	OptionPattern      `bson:"inline"`
}

func (OptionPatternInfo) CollectionName() string {
	return optionpatternCollectionName
}

type OptionPatternData struct {
	models.ShopIdentity `bson:"inline"`
	OptionPatternInfo   `bson:"inline"`
}

type OptionPatternDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	OptionPatternData  `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (OptionPatternDoc) CollectionName() string {
	return optionpatternCollectionName
}

type OptionPatternItemGuid struct {
	PatternCode string `json:"patterncode" bson:"patterncode"`
}

func (OptionPatternItemGuid) CollectionName() string {
	return optionpatternCollectionName
}

type OptionPatternActivity struct {
	OptionPatternData   `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (OptionPatternActivity) CollectionName() string {
	return optionpatternCollectionName
}

type OptionPatternDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (OptionPatternDeleteActivity) CollectionName() string {
	return optionpatternCollectionName
}
