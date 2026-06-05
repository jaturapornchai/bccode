package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const formTemplateCollectionName = "formTemplates"

type FormTemplate struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string                 `json:"code" bson:"code"`
	Names                    *[]models.NameX        `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	DocType                  string                 `json:"doctype" bson:"doctype"`
	IsDefault                bool                   `json:"isdefault" bson:"isdefault"`
	TemplateData             map[string]interface{} `json:"templatedata" bson:"templatedata"`
}

type FormTemplateInfo struct {
	models.DocIdentity `bson:"inline"`
	FormTemplate       `bson:"inline"`
}

func (FormTemplateInfo) CollectionName() string {
	return formTemplateCollectionName
}

type FormTemplateData struct {
	models.HoldingCodeentity `bson:"inline"`
	FormTemplateInfo         `bson:"inline"`
}

type FormTemplateDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"id,omitempty"`
	FormTemplateData   `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (FormTemplateDoc) CollectionName() string {
	return formTemplateCollectionName
}

type FormTemplateItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (FormTemplateItemGuid) CollectionName() string {
	return formTemplateCollectionName
}

type FormTemplateActivity struct {
	FormTemplateData    `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (FormTemplateActivity) CollectionName() string {
	return formTemplateCollectionName
}

type FormTemplateDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (FormTemplateDeleteActivity) CollectionName() string {
	return formTemplateCollectionName
}
