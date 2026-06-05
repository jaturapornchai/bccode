package models

import (
	pkgModels "smlcloudplatform/internal/models"
)

type JobProjectPg struct {
	HoldingCode string          `json:"holdingcode" gorm:"column:holdingcode;primaryKey"`
	GuidFixed   string          `json:"guidfixed" gorm:"column:guidfixed;uniqueIndex"`
	Code        string          `json:"code" gorm:"column:code"`
	Names       pkgModels.JSONB `json:"names" gorm:"column:names;type:jsonb"`
	ParentCode  string          `json:"parentcode" gorm:"column:parentcode"`
}

func (JobProjectPg) TableName() string {
	return "organization_job_project"
}
