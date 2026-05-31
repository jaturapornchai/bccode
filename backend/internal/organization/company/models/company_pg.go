package models

import (
	"time"
	pkgModels "smlcloudplatform/internal/models"
	"gorm.io/gorm"
)

type CompanyPg struct {
	ShopID    string          `json:"shopid" gorm:"column:shopid;index"`
	GuidFixed string          `json:"guid_fixed" gorm:"column:guid_fixed;primaryKey"`
	Code      string          `json:"code" gorm:"column:code;index:idx_company_code,unique"`
	Names     pkgModels.JSONB `json:"names" gorm:"column:names;type:jsonb"`
	TaxID     string          `json:"tax_id" gorm:"column:tax_id"`
	IsActive  bool            `json:"is_active" gorm:"column:is_active;default:true"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt gorm.DeletedAt  `json:"deleted_at" gorm:"index"`
}

func (CompanyPg) TableName() string {
	return "organization_companies"
}
