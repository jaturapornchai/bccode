package models

import (
	"gorm.io/gorm"
	pkgModels "smlcloudplatform/internal/models"
	"time"
)

type CompanyPg struct {
	HoldingCode string          `json:"holdingcode" gorm:"column:holdingcode;index"`
	GuidFixed   string          `json:"guidfixed" gorm:"column:guidfixed;primaryKey"`
	Code        string          `json:"code" gorm:"column:code;index:idxcompanycode,unique"`
	Names       pkgModels.JSONB `json:"names" gorm:"column:names;type:jsonb"`
	TaxID       string          `json:"taxid" gorm:"column:taxid"`
	IsActive    bool            `json:"isactive" gorm:"column:isactive;default:true"`
	CreatedAt   time.Time       `json:"createdat"`
	UpdatedAt   time.Time       `json:"updatedat"`
	DeletedAt   gorm.DeletedAt  `json:"deletedat" gorm:"index"`
}

func (CompanyPg) TableName() string {
	return "organizationcompanies"
}
