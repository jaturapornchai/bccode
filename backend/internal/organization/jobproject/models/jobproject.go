package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const jobProjectCollectionName = "organizationJobProjects"

type JobProject struct {
	models.PartitionIdentity `bson:"inline"`
	Code string          `json:"code" bson:"code"`
	Names *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	ParentCode string          `json:"parentcode,omitempty" bson:"parentcode,omitempty"`
}

type JobProjectInfo struct {
	models.DocIdentity `bson:"inline"`
	JobProject  `bson:"inline"`
}

func (JobProjectInfo) CollectionName() string {
	return jobProjectCollectionName
}

type JobProjectData struct {
	models.ShopIdentity `bson:"inline"`
	JobProjectInfo  `bson:"inline"`
}

type JobProjectDoc struct {
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	JobProjectData  `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (JobProjectDoc) CollectionName() string {
	return jobProjectCollectionName
}

type JobProjectItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (JobProjectItemGuid) CollectionName() string {
	return jobProjectCollectionName
}

type JobProjectActivity struct {
	JobProjectData  `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (JobProjectActivity) CollectionName() string {
	return jobProjectCollectionName
}

type JobProjectDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (JobProjectDeleteActivity) CollectionName() string {
	return jobProjectCollectionName
}
