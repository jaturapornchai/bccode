package models

import (
	pkgModels "smlcloudplatform/internal/models"
)

type CostCenterPg struct {
	HoldingCode string          `json:"holding_code" gorm:"column:holding_code;primaryKey"`
	GuidFixed   string          `json:"guid_fixed" gorm:"column:guid_fixed;uniqueIndex"`
	Code        string          `json:"code" gorm:"column:code"`
	Names       pkgModels.JSONB `json:"names" gorm:"column:names;type:jsonb"`
}

func (CostCenterPg) TableName() string {
	return "organization_cost_center"
}
