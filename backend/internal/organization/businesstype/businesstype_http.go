package businesstype

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/organization/businesstype/models"
	"smlcloudplatform/internal/organization/businesstype/repositories"
	"smlcloudplatform/internal/organization/businesstype/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
)

type IBusinessTypeHttp interface{}

type BusinessTypeHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IBusinessTypeHttpService
}

func NewBusinessTypeHttp(ms *microservice.Microservice, cfg config.IConfig) BusinessTypeHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewBusinessTypeRepository(pst)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewBusinessTypeHttpService(repo, masterSyncCacheRepo)

	return BusinessTypeHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h BusinessTypeHttp) RegisterHttp() {

	h.ms.POST("/organization/business-type/bulk", h.SaveBulk)

	h.ms.GET("/organization/business-type", h.SearchBusinessTypePage)
	h.ms.GET("/organization/business-type/list", h.SearchBusinessTypeStep)
	h.ms.POST("/organization/business-type", h.CreateBusinessType)
	h.ms.GET("/organization/business-type/:id", h.InfoBusinessType)
	h.ms.GET("/organization/business-type/default", h.InfoBusinessTypeDefault)
	h.ms.GET("/organization/business-type/code/:code", h.InfoBusinessTypeByCode)
	h.ms.PUT("/organization/business-type/:id", h.UpdateBusinessType)
	h.ms.DELETE("/organization/business-type/:id", h.DeleteBusinessType)
	h.ms.DELETE("/organization/business-type", h.DeleteBusinessTypeByGUIDs)
}

// Create BusinessType godoc
// @Description Create BusinessType
// @Tags		BusinessType
// @Param		BusinessType  body      models.BusinessType  true  "BusinessType"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/business-type [post]
func (h BusinessTypeHttp) CreateBusinessType(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.BusinessType{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateBusinessType(shopID, authUsername, *docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      idx,
	})
	return nil
}

// Update BusinessType godoc
// @Description Update BusinessType
// @Tags		BusinessType
// @Param		id  path      string  true  "BusinessType ID"
// @Param		BusinessType  body      models.BusinessType  true  "BusinessType"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/business-type/{id} [put]
func (h BusinessTypeHttp) UpdateBusinessType(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.BusinessType{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateBusinessType(shopID, id, authUsername, *docReq)

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

// Delete BusinessType godoc
// @Description Delete BusinessType
// @Tags		BusinessType
// @Param		id  path      string  true  "BusinessType ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/business-type/{id} [delete]
func (h BusinessTypeHttp) DeleteBusinessType(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteBusinessType(shopID, id, authUsername)

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

// Delete BusinessType godoc
// @Description Delete BusinessType
// @Tags		BusinessType
// @Param		BusinessType  body      []string  true  "BusinessType GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/business-type [delete]
func (h BusinessTypeHttp) DeleteBusinessTypeByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteBusinessTypeByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get BusinessType godoc
// @Description get BusinessType info by guidfixed
// @Tags		BusinessType
// @Param		id  path      string  true  "BusinessType guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/business-type/{id} [get]
func (h BusinessTypeHttp) InfoBusinessType(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get BusinessType %v", id)
	doc, err := h.svc.InfoBusinessType(shopID, id)

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

// Get BusinessType default godoc
// @Description get BusinessType info default
// @Tags		BusinessType
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/business-type/default [get]
func (h BusinessTypeHttp) InfoBusinessTypeDefault(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	doc, err := h.svc.InfoBusinessTypeDefault(shopID)

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

// Get BusinessType By Code godoc
// @Description get BusinessType info by Code
// @Tags		BusinessType
// @Param		code  path      string  true  "BusinessType Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/business-type/code/{code} [get]
func (h BusinessTypeHttp) InfoBusinessTypeByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")

	doc, err := h.svc.InfoBusinessTypeByCode(shopID, code)

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

// List BusinessType step godoc
// @Description get list step
// @Tags		BusinessType
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/business-type [get]
func (h BusinessTypeHttp) SearchBusinessTypePage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.svc.SearchBusinessType(shopID, map[string]interface{}{}, pageable)

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

// List BusinessType godoc
// @Description search limit offset
// @Tags		BusinessType
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/business-type/list [get]
func (h BusinessTypeHttp) SearchBusinessTypeStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	docList, total, err := h.svc.SearchBusinessTypeStep(shopID, lang, pageableStep)

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

// Create BusinessType Bulk godoc
// @Description Create BusinessType
// @Tags		BusinessType
// @Param		BusinessType  body      []models.BusinessType  true  "BusinessType"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/business-type/bulk [post]
func (h BusinessTypeHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.BusinessType{}
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
