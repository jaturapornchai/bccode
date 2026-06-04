package models

import (
	pkgModels "smlcloudplatform/internal/models"
)

type JobProjectPg struct {
	HoldingCode string          `json:"holding_code" gorm:"column:holding_code;primaryKey"`
	GuidFixed   string          `json:"guid_fixed" gorm:"column:guid_fixed;uniqueIndex"`
	Code        string          `json:"code" gorm:"column:code"`
	Names       pkgModels.JSONB `json:"names" gorm:"column:names;type:jsonb"`
	ParentCode  string          `json:"parentcode" gorm:"column:parentcode"`
}

func (JobProjectPg) TableName() string {
	return "organization_job_project"
}
