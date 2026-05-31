package company

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	companyModels "smlcloudplatform/internal/organization/company/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type CompanyHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
}

func NewCompanyHttp(ms *microservice.Microservice, cfg config.IConfig) CompanyHttp {
	return CompanyHttp{
		ms:  ms,
		cfg: cfg,
	}
}

func (h CompanyHttp) RegisterHttp() {
	h.ms.POST("/organization/company", h.CreateCompany)
	h.ms.GET("/organization/company", h.SearchCompany)
	h.ms.GET("/organization/company/:id", h.InfoCompany)
	h.ms.PUT("/organization/company/:id", h.UpdateCompany)
	h.ms.DELETE("/organization/company/:id", h.DeleteCompany)
}

func (h CompanyHttp) CreateCompany(ctx microservice.IContext) error {
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	var req companyModels.CompanyPg
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	req.ShopID = shopID
	if req.GuidFixed == "" {
		req.GuidFixed = utils.NewGUID()
	}
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()
	req.IsActive = true

	pst := h.ms.PersisterTenant(h.cfg.PersisterConfig(), shopID)
	db := pst.DBClient()
	if err := db.Create(&req).Error; err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      req.GuidFixed,
	})
	return nil
}

func (h CompanyHttp) SearchCompany(ctx microservice.IContext) error {
	shopID := ctx.UserInfo().ShopID

	pst := h.ms.PersisterTenant(h.cfg.PersisterConfig(), shopID)
	db := pst.DBClient()
	var list []companyModels.CompanyPg
	if err := db.Where("shopid = ?", shopID).Find(&list).Error; err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    list,
	})
	return nil
}

func (h CompanyHttp) InfoCompany(ctx microservice.IContext) error {
	shopID := ctx.UserInfo().ShopID
	id := ctx.Param("id")

	pst := h.ms.PersisterTenant(h.cfg.PersisterConfig(), shopID)
	db := pst.DBClient()
	var data companyModels.CompanyPg
	if err := db.Where("shopid = ? AND guid_fixed = ?", shopID, id).First(&data).Error; err != nil {
		ctx.ResponseError(http.StatusNotFound, "Company not found")
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    data,
	})
	return nil
}

func (h CompanyHttp) UpdateCompany(ctx microservice.IContext) error {
	shopID := ctx.UserInfo().ShopID
	id := ctx.Param("id")
	input := ctx.ReadInput()

	pst := h.ms.PersisterTenant(h.cfg.PersisterConfig(), shopID)
	db := pst.DBClient()
	var existing companyModels.CompanyPg
	if err := db.Where("shopid = ? AND guid_fixed = ?", shopID, id).First(&existing).Error; err != nil {
		ctx.ResponseError(http.StatusNotFound, "Company not found")
		return err
	}

	var req companyModels.CompanyPg
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	existing.Names = req.Names
	existing.TaxID = req.TaxID
	existing.Code = req.Code
	existing.IsActive = req.IsActive
	existing.UpdatedAt = time.Now()

	if err := db.Save(&existing).Error; err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      id,
	})
	return nil
}

func (h CompanyHttp) DeleteCompany(ctx microservice.IContext) error {
	shopID := ctx.UserInfo().ShopID
	id := ctx.Param("id")

	pst := h.ms.PersisterTenant(h.cfg.PersisterConfig(), shopID)
	db := pst.DBClient()
	var data companyModels.CompanyPg
	if err := db.Where("shopid = ? AND guid_fixed = ?", shopID, id).First(&data).Error; err != nil {
		ctx.ResponseError(http.StatusNotFound, "Company not found")
		return err
	}

	// Soft delete via GORM
	if err := db.Delete(&data).Error; err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      id,
	})
	return nil
}
