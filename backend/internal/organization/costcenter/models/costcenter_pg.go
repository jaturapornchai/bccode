package models

import (
	pkgModels "smlcloudplatform/internal/models"
)

type CostCenterPg struct {
	ShopID    string          `json:"shopid" gorm:"column:shopid;primaryKey"`
	GuidFixed string          `json:"guidfixed" gorm:"column:guidfixed;uniqueIndex"`
	Code      string          `json:"code" gorm:"column:code"`
	Names     pkgModels.JSONB `json:"names" gorm:"column:names;type:jsonb"`
}

func (CostCenterPg) TableName() string {
	return "organization_cost_center"
}
