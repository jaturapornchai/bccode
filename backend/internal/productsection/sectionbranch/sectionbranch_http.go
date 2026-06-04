package sectionbranch

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/productsection/sectionbranch/models"
	"smlcloudplatform/internal/productsection/sectionbranch/repositories"
	"smlcloudplatform/internal/productsection/sectionbranch/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
)

type ISectionBranchHttp interface{}

type SectionBranchHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.ISectionBranchHttpService
}

func NewSectionBranchHttp(ms *microservice.Microservice, cfg config.IConfig) SectionBranchHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewSectionBranchRepository(pst)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewSectionBranchHttpService(repo, utils.NewGUID, masterSyncCacheRepo)

	return SectionBranchHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h SectionBranchHttp) RegisterHttp() {

	h.ms.POST("/product-section/branch/bulk", h.SaveBulk)

	h.ms.GET("/product-section/branch", h.SearchSectionBranchPage)
	h.ms.GET("/product-section/branch/list", h.SearchSectionBranchStep)
	h.ms.GET("/product-section/branch/:id", h.InfoSectionBranch)
	h.ms.GET("/product-section/branch/code/:code", h.InfoSectionBranchByCode)
	h.ms.PUT("/product-section/branch", h.SaveSectionBranch)
	h.ms.DELETE("/product-section/branch/:id", h.DeleteSectionBranch)
	h.ms.DELETE("/product-section/branch", h.DeleteSectionBranchByGUIDs)
}

// Save SectionBranch godoc
// @Description Save SectionBranch
// @Tags		SectionBranch
// @Param		SectionBranch  body      models.SectionBranch  true  "SectionBranch"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product-section/branch [put]
func (h SectionBranchHttp) SaveSectionBranch(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()

	docReq := &models.SectionBranch{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	id, err := h.svc.SaveSectionBranch(holdingCode, authUsername, *docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      id,
	})

	return nil
}

// Delete SectionBranch godoc
// @Description Delete SectionBranch
// @Tags		SectionBranch
// @Param		id  path      string  true  "SectionBranch ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product-section/branch/{id} [delete]
func (h SectionBranchHttp) DeleteSectionBranch(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteSectionBranch(holdingCode, id, authUsername)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      id,
	})

	return nil
}

// Delete SectionBranch godoc
// @Description Delete SectionBranch
// @Tags		SectionBranch
// @Param		SectionBranch  body      []string  true  "SectionBranch GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product-section/branch [delete]
func (h SectionBranchHttp) DeleteSectionBranchByGUIDs(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	input := ctx.ReadInput()

	docReq := []string{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.DeleteSectionBranchByGUIDs(holdingCode, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get SectionBranch godoc
// @Description get SectionBranch info by guidfixed
// @Tags		SectionBranch
// @Param		id  path      string  true  "SectionBranch guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product-section/branch/{id} [get]
func (h SectionBranchHttp) InfoSectionBranch(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get SectionBranch %v", id)
	doc, err := h.svc.InfoSectionBranch(holdingCode, id)

	if err != nil {
		h.ms.Logger.Errorf("Error getting document %v: %v", id, err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    doc,
	})
	return nil
}

// Get SectionBranch By Code godoc
// @Description get SectionBranch info by Code
// @Tags		SectionBranch
// @Param		code  path      string  true  "SectionBranch Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product-section/branch/code/{code} [get]
func (h SectionBranchHttp) InfoSectionBranchByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	code := ctx.Param("code")

	doc, err := h.svc.InfoSectionBranchByBranchCode(holdingCode, code)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    doc,
	})
	return nil
}

// List SectionBranch step godoc
// @Description get list step
// @Tags		SectionBranch
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product-section/branch [get]
func (h SectionBranchHttp) SearchSectionBranchPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.svc.SearchSectionBranch(holdingCode, map[string]interface{}{}, pageable)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success:    true,
		Data:       docList,
		Pagination: pagination,
	})
	return nil
}

// List SectionBranch godoc
// @Description search limit offset
// @Tags		SectionBranch
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product-section/branch/list [get]
func (h SectionBranchHttp) SearchSectionBranchStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	docList, total, err := h.svc.SearchSectionBranchStep(holdingCode, lang, pageableStep)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    docList,
		Total:   total,
	})
	return nil
}

// Create SectionBranch Bulk godoc
// @Description Create SectionBranch
// @Tags		SectionBranch
// @Param		SectionBranch  body      []models.SectionBranch  true  "SectionBranch"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product-section/branch/bulk [post]
func (h SectionBranchHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()

	dataReq := []models.SectionBranch{}
	err := json.Unmarshal([]byte(input), &dataReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	bulkResponse, err := h.svc.SaveInBatch(holdingCode, authUsername, dataReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	ctx.Response(
		http.StatusCreated,
		common.BulkResponse{
			Success:    true,
			BulkImport: bulkResponse,
		},
	)

	return nil
}
