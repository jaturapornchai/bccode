// patternproduct_http.go
package patternproduct

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/internal/smlaiproduct/patternproduct/models"
	patternrepo "smlcloudplatform/internal/smlaiproduct/patternproduct/repositories"
	patternsvc "smlcloudplatform/internal/smlaiproduct/patternproduct/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IPatternProductHttp interface{}

type PatternProductHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc patternsvc.IPatternProductHttpService
}

func NewPatternProductHttp(ms *microservice.Microservice, cfg config.IConfig) PatternProductHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := patternrepo.NewPatternProductRepository(pst)
	repoProductBarcode := productbarcode_repositories.NewProductBarcodeRepository(pst, cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)

	svc := patternsvc.NewPatternProductHttpService(repo, repoProductBarcode, masterSyncCacheRepo, 15*time.Second)

	return PatternProductHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h PatternProductHttp) RegisterHttp() {
	h.ms.POST("/aicloud/pattern/bulk", h.SaveBulk)
	h.ms.GET("/aicloud/pattern", h.SearchPatternProductPage)
	h.ms.GET("/aicloud/pattern/list", h.SearchPatternProductStep)
	h.ms.POST("/aicloud/pattern", h.CreatePatternProduct)
	h.ms.GET("/aicloud/pattern/:id", h.InfoPatternProduct)
	h.ms.GET("/aicloud/pattern/code/:code", h.InfoPatternProductByCode)
	h.ms.PUT("/aicloud/pattern/:id", h.UpdatePatternProduct)
	h.ms.DELETE("/aicloud/pattern/:id", h.DeletePatternProduct)
	h.ms.DELETE("/aicloud/pattern", h.DeletePatternProductByGUIDs)
}

// Create PatternProduct godoc
// @Description Create PatternProduct
// @Tags		AICloudProductPattern
// @Param		PatternProduct	body	models.PatternProduct	true	"PatternProduct"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401	{object}	common.AuthResponseFailed
// @Security		AccessToken
// @Router /aicloud/pattern [post]
func (h PatternProductHttp) CreatePatternProduct(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()
	docReq := &models.PatternProduct{}
	if err := json.Unmarshal([]byte(input), docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err := ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreatePatternProduct(holdingCode, authUsername, *docReq)
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

// Update PatternProduct godoc
// @Description Update PatternProduct
// @Tags		AICloudProductPattern
// @Param		id	path	string	true	"PatternProduct ID"
// @Param		PatternProduct	body	models.PatternProduct	true	"PatternProduct"
// @Accept		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401	{object}	common.AuthResponseFailed
// @Security		AccessToken
// @Router /aicloud/pattern/{id} [put]
func (h PatternProductHttp) UpdatePatternProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")
	input := ctx.ReadInput()
	docReq := &models.PatternProduct{}
	if err := json.Unmarshal([]byte(input), docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err := ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err := h.svc.UpdatePatternProduct(holdingCode, id, authUsername, *docReq)
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

// Delete PatternProduct godoc
// @Description Delete PatternProduct
// @Tags		AICloudProductPattern
// @Param		id  path      string  true  "PatternProduct ID"
// @Accept 		json
// @Success		200	{object} common.ResponseSuccessWithID
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/pattern/{id} [delete]
func (h PatternProductHttp) DeletePatternProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeletePatternProduct(holdingCode, id, authUsername)
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

// Delete PatternProduct godoc
// @Description Delete PatternProduct
// @Tags		AICloudProductPattern
// @Param		PatternProduct  body      []string  true  "PatternProduct GUIDs"
// @Accept 		json
// @Success		200	{object} common.ResponseSuccessWithID
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/pattern [delete]
func (h PatternProductHttp) DeletePatternProductByGUIDs(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	input := ctx.ReadInput()
	docReq := []string{}
	if err := json.Unmarshal([]byte(input), &docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err := h.svc.DeletePatternProductByGUIDs(holdingCode, authUsername, docReq)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})
	return nil
}

// Get PatternProduct godoc
// @Description get PatternProduct info by guidfixed
// @Tags		AICloudProductPattern
// @Param		id  path      string  true  "PatternProduct guidfixed"
// @Accept 		json
// @Success		200	{object} common.ApiResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/pattern/{id} [get]
func (h PatternProductHttp) InfoPatternProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get PatternProduct %v", id)
	doc, err := h.svc.InfoPatternProduct(holdingCode, id)
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

// Get PatternProduct By Code godoc
// @Description get PatternProduct info by Code
// @Tags		AICloudProductPattern
// @Param		code  path      string  true  "PatternProduct Code"
// @Accept 		json
// @Success		200	{object} common.ApiResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/pattern/code/{code} [get]
func (h PatternProductHttp) InfoPatternProductByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	code := ctx.Param("code")
	doc, err := h.svc.InfoPatternProductByCode(holdingCode, code)
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

// List PatternProduct step godoc
// @Description get list step
// @Tags		AICloudProductPattern
// @Param		q	query	string	false  "Search Value"
// @Param		page	query	integer	false  "Page"
// @Param		limit	query	integer	false  "Limit"
// @Accept 		json
// @Success		200	{array} common.ApiResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/pattern [get]
func (h PatternProductHttp) SearchPatternProductPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageable := utils.GetPageable(ctx.QueryParam)
	docList, pagination, err := h.svc.SearchPatternProduct(holdingCode, map[string]interface{}{}, pageable)
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

// List PatternProduct godoc
// @Description search limit offset
// @Tags		AICloudProductPattern
// @Param		q	query	string	false  "Search Value"
// @Param		offset	query	integer	false  "offset"
// @Param		limit	query	integer	false  "limit"
// @Param		lang	query	string	false  "lang"
// @Accept 		json
// @Success		200	{array} common.ApiResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/pattern/list [get]
func (h PatternProductHttp) SearchPatternProductStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageableStep := utils.GetPageableStep(ctx.QueryParam)
	lang := ctx.QueryParam("lang")
	docList, total, err := h.svc.SearchPatternProductStep(holdingCode, lang, pageableStep)
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

// Create PatternProduct Bulk godoc
// @Description Create PatternProduct
// @Tags		AICloudProductPattern
// @Param		PatternProduct  body      []models.PatternProduct  true  "PatternProduct"
// @Accept 		json
// @Success		201	{object} common.BulkResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/pattern/bulk [post]
func (h PatternProductHttp) SaveBulk(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()
	dataReq := []models.PatternProduct{}
	if err := json.Unmarshal([]byte(input), &dataReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	bulkResponse, err := h.svc.SaveInBatch(holdingCode, authUsername, dataReq)
	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.BulkResponse{
		Success:    true,
		BulkImport: bulkResponse,
	})
	return nil
}
