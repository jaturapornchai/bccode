package models

import (
	"time"

	pkgModels "smlcloudplatform/internal/models"

	"gorm.io/gorm"
)

type BranchPg struct {
	HoldingCode string          `json:"holdingcode" gorm:"column:holdingcode;index;uniqueIndex:idxbranchshopcompanycode,where:deletedat IS NULL"`
	GuidFixed   string          `json:"guidfixed" gorm:"column:guidfixed;primaryKey"`
	CompanyGuid string          `json:"companyguid" gorm:"column:companyguid;index;uniqueIndex:idxbranchshopcompanycode,where:deletedat IS NULL"`
	Code        string          `json:"code" gorm:"column:code;index;uniqueIndex:idxbranchshopcompanycode,where:deletedat IS NULL"`
	Names       pkgModels.JSONB `json:"names" gorm:"column:names;type:jsonb"`
	IsActive    bool            `json:"isactive" gorm:"column:isactive;default:true"`
	CreatedAt   time.Time       `json:"createdat"`
	UpdatedAt   time.Time       `json:"updatedat"`
	DeletedAt   gorm.DeletedAt  `json:"deletedat" gorm:"index"`
}

func (BranchPg) TableName() string {
	return "organizationbranches"
}
