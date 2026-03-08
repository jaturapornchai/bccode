package sectionbusinesstype

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/productsection/sectionbusinesstype/models"
	"smlcloudplatform/internal/productsection/sectionbusinesstype/repositories"
	"smlcloudplatform/internal/productsection/sectionbusinesstype/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
)

type ISectionBusinessTypeHttp interface{}

type SectionBusinessTypeHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.ISectionBusinessTypeHttpService
}

func NewSectionBusinessTypeHttp(ms *microservice.Microservice, cfg config.IConfig) SectionBusinessTypeHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewSectionBusinessTypeRepository(pst)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewSectionBusinessTypeHttpService(repo, masterSyncCacheRepo)

	return SectionBusinessTypeHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h SectionBusinessTypeHttp) RegisterHttp() {

	h.ms.POST("/product-section/business-type/bulk", h.SaveBulk)

	h.ms.GET("/product-section/business-type", h.SearchSectionBusinessTypePage)
	h.ms.GET("/product-section/business-type/list", h.SearchSectionBusinessTypeStep)
	h.ms.GET("/product-section/business-type/:id", h.InfoSectionBusinessType)
	h.ms.GET("/product-section/business-type/code/:code", h.InfoSectionBusinessTypeByCode)
	h.ms.PUT("/product-section/business-type", h.SaveSectionBusinessType)
	h.ms.DELETE("/product-section/business-type/:id", h.DeleteSectionBusinessType)
	h.ms.DELETE("/product-section/business-type", h.DeleteSectionBusinessTypeByGUIDs)
}

// Save SectionBusinessType godoc
// @Description Save SectionBusinessType
// @Tags		SectionBusinessType
// @Param		SectionBusinessType  body      models.SectionBusinessType  true  "SectionBusinessType"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product-section/business-type [put]
func (h SectionBusinessTypeHttp) SaveSectionBusinessType(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	docReq := &models.SectionBusinessType{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	id, err := h.svc.SaveSectionBusinessType(shopID, authUsername, *docReq)

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

// Delete SectionBusinessType godoc
// @Description Delete SectionBusinessType
// @Tags		SectionBusinessType
// @Param		id  path      string  true  "SectionBusinessType ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product-section/business-type/{id} [delete]
func (h SectionBusinessTypeHttp) DeleteSectionBusinessType(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteSectionBusinessType(shopID, id, authUsername)

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

// Delete SectionBusinessType godoc
// @Description Delete SectionBusinessType
// @Tags		SectionBusinessType
// @Param		SectionBusinessType  body      []string  true  "SectionBusinessType GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product-section/business-type [delete]
func (h SectionBusinessTypeHttp) DeleteSectionBusinessTypeByGUIDs(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	input := ctx.ReadInput()

	docReq := []string{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.DeleteSectionBusinessTypeByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get SectionBusinessType godoc
// @Description get SectionBusinessType info by guidfixed
// @Tags		SectionBusinessType
// @Param		id  path      string  true  "SectionBusinessType guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product-section/business-type/{id} [get]
func (h SectionBusinessTypeHttp) InfoSectionBusinessType(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get SectionBusinessType %v", id)
	doc, err := h.svc.InfoSectionBusinessType(shopID, id)

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

// Get SectionBusinessType By Code godoc
// @Description get SectionBusinessType info by Code
// @Tags		SectionBusinessType
// @Param		code  path      string  true  "SectionBusinessType Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product-section/business-type/code/{code} [get]
func (h SectionBusinessTypeHttp) InfoSectionBusinessTypeByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")

	doc, err := h.svc.InfoSectionBusinessTypeByBusinessTypeCode(shopID, code)

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

// List SectionBusinessType step godoc
// @Description get list step
// @Tags		SectionBusinessType
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product-section/business-type [get]
func (h SectionBusinessTypeHttp) SearchSectionBusinessTypePage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.svc.SearchSectionBusinessType(shopID, map[string]interface{}{}, pageable)

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

// List SectionBusinessType godoc
// @Description search limit offset
// @Tags		SectionBusinessType
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product-section/business-type/list [get]
func (h SectionBusinessTypeHttp) SearchSectionBusinessTypeStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	docList, total, err := h.svc.SearchSectionBusinessTypeStep(shopID, lang, pageableStep)

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

// Create SectionBusinessType Bulk godoc
// @Description Create SectionBusinessType
// @Tags		SectionBusinessType
// @Param		SectionBusinessType  body      []models.SectionBusinessType  true  "SectionBusinessType"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product-section/business-type/bulk [post]
func (h SectionBusinessTypeHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.SectionBusinessType{}
	err := json.Unmarshal([]byte(input), &dataReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	bulkResponse, err := h.svc.SaveInBatch(shopID, authUsername, dataReq)

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
