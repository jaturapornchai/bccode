package models

import "smlcloudplatform/internal/models"

type TableDesign struct {
	Index        int `json:"index" bson:"index" validate:"required"`
	ObjectDesign `bson:"inline"`
	Number       string `json:"number" bson:"number"`
	Seates       int    `json:"seates" bson:"seates" validate:"required"`
	models.Name  `bson:"inline"`
}
