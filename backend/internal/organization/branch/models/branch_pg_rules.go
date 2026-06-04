package models

import (
	"errors"
	"strings"
	"time"

	companyModels "smlcloudplatform/internal/organization/company/models"

	"gorm.io/gorm"
)

var (
	ErrBranchCompanyGuidRequired = errors.New("company_guid is required")
	ErrBranchCompanyNotFound     = errors.New("company not found")
	ErrBranchCodeExists          = errors.New("branch code is exists")
)

func PrepareCreateBranch(req *BranchPg, holdingCode string, now time.Time, newGUID func() string) error {
	if err := PrepareUpdateBranch(req); err != nil {
		return err
	}
	req.HoldingCode = holdingCode
	if req.GuidFixed == "" {
		req.GuidFixed = newGUID()
	}
	req.CreatedAt = now
	req.UpdatedAt = now
	req.IsActive = true
	return nil
}

func PrepareUpdateBranch(req *BranchPg) error {
	req.CompanyGuid = strings.TrimSpace(req.CompanyGuid)
	if req.CompanyGuid == "" {
		return ErrBranchCompanyGuidRequired
	}

	normalizedCode, err := NormalizeThaiTaxBranchCode(req.Code)
	if err != nil {
		return err
	}
	req.Code = normalizedCode
	return nil
}

func BranchCompanyLookup(db *gorm.DB, holdingCode string, companyGuid string) *gorm.DB {
	return db.Where("holding_code = ? AND guid_fixed = ?", holdingCode, companyGuid)
}

func BranchDuplicateLookup(db *gorm.DB, holdingCode string, companyGuid string, code string, excludeGuid string) *gorm.DB {
	query := db.Where("holding_code = ? AND company_guid = ? AND code = ?", holdingCode, companyGuid, code)
	if excludeGuid != "" {
		query = query.Where("guid_fixed <> ?", excludeGuid)
	}
	return query
}

func CompanyBranchCountLookup(db *gorm.DB, holdingCode string, companyGuid string) *gorm.DB {
	return db.Model(&BranchPg{}).Where("holding_code = ? AND company_guid = ?", holdingCode, companyGuid)
}

func EnsureCompanyExists(db *gorm.DB, holdingCode string, companyGuid string) error {
	var company companyModels.CompanyPg
	err := BranchCompanyLookup(db, holdingCode, companyGuid).First(&company).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrBranchCompanyNotFound
	}
	return err
}

func EnsureBranchCodeAvailable(db *gorm.DB, holdingCode string, companyGuid string, code string, excludeGuid string) error {
	var existing BranchPg
	err := BranchDuplicateLookup(db, holdingCode, companyGuid, code, excludeGuid).First(&existing).Error
	if err == nil {
		return ErrBranchCodeExists
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}
