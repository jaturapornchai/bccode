package branch

import (
	"encoding/json"
	"errors"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	branchModels "smlcloudplatform/internal/organization/branch/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type BranchHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
}

func NewBranchHttp(ms *microservice.Microservice, cfg config.IConfig) BranchHttp {
	return BranchHttp{
		ms:  ms,
		cfg: cfg,
	}
}

func (h BranchHttp) RegisterHttp() {
	h.ms.POST("/organization/branch", h.CreateBranch)
	h.ms.GET("/organization/branch", h.SearchBranch)
	h.ms.GET("/organization/branch/list", h.SearchBranchStep)
	h.ms.GET("/organization/branch/:id", h.InfoBranch)
	h.ms.PUT("/organization/branch/:id", h.UpdateBranch)
	h.ms.DELETE("/organization/branch/:id", h.DeleteBranch)
}

func (h BranchHttp) CreateBranch(ctx microservice.IContext) error {
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	var req branchModels.BranchPg
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	normalizedCode, err := branchModels.NormalizeThaiTaxBranchCode(req.Code)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	req.Code = normalizedCode
	req.ShopID = shopID
	if req.GuidFixed == "" {
		req.GuidFixed = utils.NewGUID()
	}
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()
	req.IsActive = true

	pst := h.ms.PersisterTenant(h.cfg.PersisterConfig(), shopID)
	db := pst.DBClient()
	var existing branchModels.BranchPg
	err = db.Where("shopid = ? AND code = ?", shopID, req.Code).First(&existing).Error
	if err == nil {
		ctx.ResponseError(http.StatusConflict, "branch code is exists")
		return errors.New("branch code is exists")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}
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

func (h BranchHttp) SearchBranch(ctx microservice.IContext) error {
	shopID := ctx.UserInfo().ShopID
	companyGuid := ctx.QueryParam("company_guid")

	pst := h.ms.PersisterTenant(h.cfg.PersisterConfig(), shopID)
	db := pst.DBClient()
	var list []branchModels.BranchPg

	query := db.Where("shopid = ?", shopID)
	if companyGuid != "" {
		query = query.Where("company_guid = ?", companyGuid)
	}

	if err := query.Find(&list).Error; err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    list,
	})
	return nil
}

func (h BranchHttp) InfoBranch(ctx microservice.IContext) error {
	shopID := ctx.UserInfo().ShopID
	id := ctx.Param("id")

	pst := h.ms.PersisterTenant(h.cfg.PersisterConfig(), shopID)
	db := pst.DBClient()
	var data branchModels.BranchPg
	if err := db.Where("shopid = ? AND guid_fixed = ?", shopID, id).First(&data).Error; err != nil {
		ctx.ResponseError(http.StatusNotFound, "Branch not found")
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    data,
	})
	return nil
}

func (h BranchHttp) UpdateBranch(ctx microservice.IContext) error {
	shopID := ctx.UserInfo().ShopID
	id := ctx.Param("id")
	input := ctx.ReadInput()

	pst := h.ms.PersisterTenant(h.cfg.PersisterConfig(), shopID)
	db := pst.DBClient()
	var existing branchModels.BranchPg
	if err := db.Where("shopid = ? AND guid_fixed = ?", shopID, id).First(&existing).Error; err != nil {
		ctx.ResponseError(http.StatusNotFound, "Branch not found")
		return err
	}

	var req branchModels.BranchPg
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	normalizedCode, err := branchModels.NormalizeThaiTaxBranchCode(req.Code)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	req.Code = normalizedCode
	if req.Code != existing.Code {
		var duplicate branchModels.BranchPg
		err = db.Where("shopid = ? AND code = ? AND guid_fixed <> ?", shopID, req.Code, id).First(&duplicate).Error
		if err == nil {
			ctx.ResponseError(http.StatusConflict, "branch code is exists")
			return errors.New("branch code is exists")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.ResponseError(http.StatusInternalServerError, err.Error())
			return err
		}
	}
	existing.Names = req.Names
	existing.Code = req.Code
	existing.CompanyGuid = req.CompanyGuid
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

func (h BranchHttp) DeleteBranch(ctx microservice.IContext) error {
	shopID := ctx.UserInfo().ShopID
	id := ctx.Param("id")

	pst := h.ms.PersisterTenant(h.cfg.PersisterConfig(), shopID)
	db := pst.DBClient()
	var data branchModels.BranchPg
	if err := db.Where("shopid = ? AND guid_fixed = ?", shopID, id).First(&data).Error; err != nil {
		ctx.ResponseError(http.StatusNotFound, "Branch not found")
		return err
	}

	if branchModels.IsThaiHeadOfficeBranchCode(data.Code) {
		ctx.ResponseError(http.StatusBadRequest, "head office branch cannot be deleted")
		return errors.New("head office branch cannot be deleted")
	}

	var total int64
	if err := db.Model(&branchModels.BranchPg{}).Where("shopid = ?", shopID).Count(&total).Error; err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}
	if total <= 1 {
		ctx.ResponseError(http.StatusBadRequest, "company must have at least one branch")
		return errors.New("company must have at least one branch")
	}

	// Soft delete
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

func (h BranchHttp) SearchBranchStep(ctx microservice.IContext) error {
	shopID := ctx.UserInfo().ShopID
	q := ctx.QueryParam("q")
	offsetStr := ctx.QueryParam("offset")
	limitStr := ctx.QueryParam("limit")

	offset := 0
	limit := 100
	if offsetStr != "" {
		if val, err := strconv.Atoi(offsetStr); err == nil {
			offset = val
		}
	}
	if limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil {
			limit = val
		}
	}

	pst := h.ms.PersisterTenant(h.cfg.PersisterConfig(), shopID)
	db := pst.DBClient()
	var list []branchModels.BranchPg

	query := db.Where("shopid = ?", shopID)
	if q != "" {
		query = query.Where("code ILIKE ? OR names::text ILIKE ?", "%"+q+"%", "%"+q+"%")
	}

	var total int64
	if err := query.Model(&branchModels.BranchPg{}).Count(&total).Error; err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	if err := query.Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    list,
		Total:   total,
	})
	return nil
}
