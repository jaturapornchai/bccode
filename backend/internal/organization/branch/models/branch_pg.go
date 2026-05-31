package models

import (
	"time"

	pkgModels "smlcloudplatform/internal/models"

	"gorm.io/gorm"
)

type BranchPg struct {
	ShopID      string          `json:"shopid" gorm:"column:shopid;index;uniqueIndex:idx_branch_shop_company_code,where:deleted_at IS NULL"`
	GuidFixed   string          `json:"guid_fixed" gorm:"column:guid_fixed;primaryKey"`
	CompanyGuid string          `json:"company_guid" gorm:"column:company_guid;index;uniqueIndex:idx_branch_shop_company_code,where:deleted_at IS NULL"`
	Code        string          `json:"code" gorm:"column:code;index;uniqueIndex:idx_branch_shop_company_code,where:deleted_at IS NULL"`
	Names       pkgModels.JSONB `json:"names" gorm:"column:names;type:jsonb"`
	IsActive    bool            `json:"is_active" gorm:"column:is_active;default:true"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `json:"deleted_at" gorm:"index"`
}

func (BranchPg) TableName() string {
	return "organization_branches"
}
