package models

import (
	"smlcloudplatform/internal/models"
)

type BusinessType struct {
	models.PartitionIdentity
	Code      string          `json:"code"`
	Names     *[]models.NameX `json:"names" validate:"required,min=1,unique=Code,dive"`
	IsDefault bool            `json:"isdefault"`
}

type BusinessTypeInfo struct {
	models.DocIdentity
	BusinessType
}
