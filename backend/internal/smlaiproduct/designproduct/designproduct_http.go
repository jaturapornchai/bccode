// designproduct_http.go
package designproduct

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/internal/smlaiproduct/designproduct/models"
	designrepo "smlcloudplatform/internal/smlaiproduct/designproduct/repositories"
	designsvc "smlcloudplatform/internal/smlaiproduct/designproduct/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IDesignProductHttp interface{}

type DesignProductHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc designsvc.IDesignProductHttpService
}

func NewDesignProductHttp(ms *microservice.Microservice, cfg config.IConfig) DesignProductHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := designrepo.NewDesignProductRepository(pst)
	repoProductBarcode := productbarcode_repositories.NewProductBarcodeRepository(pst, cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)

	svc := designsvc.NewDesignProductHttpService(repo, repoProductBarcode, masterSyncCacheRepo, 15*time.Second)

	return DesignProductHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h DesignProductHttp) RegisterHttp() {
	h.ms.POST("/aicloud/design/bulk", h.SaveBulk)
	h.ms.GET("/aicloud/design", h.SearchDesignProductPage)
	h.ms.GET("/aicloud/design/list", h.SearchDesignProductStep)
	h.ms.POST("/aicloud/design", h.CreateDesignProduct)
	h.ms.GET("/aicloud/design/:id", h.InfoDesignProduct)
	h.ms.GET("/aicloud/design/code/:code", h.InfoDesignProductByCode)
	h.ms.PUT("/aicloud/design/:id", h.UpdateDesignProduct)
	h.ms.DELETE("/aicloud/design/:id", h.DeleteDesignProduct)
	h.ms.DELETE("/aicloud/design", h.DeleteDesignProductByGUIDs)
}

// Create DesignProduct godoc
// @Description Create DesignProduct
// @Tags		AICloudProductDesign
// @Param		DesignProduct	body	models.DesignProduct	true	"DesignProduct"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401	{object}	common.AuthResponseFailed
// @Security		AccessToken
// @Router /aicloud/design [post]
func (h DesignProductHttp) CreateDesignProduct(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()
	docReq := &models.DesignProduct{}
	if err := json.Unmarshal([]byte(input), docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err := ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateDesignProduct(holdingCode, authUsername, *docReq)
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

// Update DesignProduct godoc
// @Description Update DesignProduct
// @Tags		AICloudProductDesign
// @Param		id	path	string	true	"DesignProduct ID"
// @Param		DesignProduct	body	models.DesignProduct	true	"DesignProduct"
// @Accept		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401	{object}	common.AuthResponseFailed
// @Security		AccessToken
// @Router /aicloud/design/{id} [put]
func (h DesignProductHttp) UpdateDesignProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")
	input := ctx.ReadInput()
	docReq := &models.DesignProduct{}
	if err := json.Unmarshal([]byte(input), docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err := ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err := h.svc.UpdateDesignProduct(holdingCode, id, authUsername, *docReq)
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

// Delete DesignProduct godoc
// @Description Delete DesignProduct
// @Tags		AICloudProductDesign
// @Param		id  path      string  true  "DesignProduct ID"
// @Accept 		json
// @Success		200	{object} common.ResponseSuccessWithID
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/design/{id} [delete]
func (h DesignProductHttp) DeleteDesignProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteDesignProduct(holdingCode, id, authUsername)
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

// Delete DesignProduct godoc
// @Description Delete DesignProduct
// @Tags		AICloudProductDesign
// @Param		DesignProduct  body      []string  true  "DesignProduct GUIDs"
// @Accept 		json
// @Success		200	{object} common.ResponseSuccessWithID
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/design [delete]
func (h DesignProductHttp) DeleteDesignProductByGUIDs(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	input := ctx.ReadInput()
	docReq := []string{}
	if err := json.Unmarshal([]byte(input), &docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err := h.svc.DeleteDesignProductByGUIDs(holdingCode, authUsername, docReq)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})
	return nil
}

// Get DesignProduct godoc
// @Description get DesignProduct info by guidfixed
// @Tags		AICloudProductDesign
// @Param		id  path      string  true  "DesignProduct guidfixed"
// @Accept 		json
// @Success		200	{object} common.ApiResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/design/{id} [get]
func (h DesignProductHttp) InfoDesignProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get DesignProduct %v", id)
	doc, err := h.svc.InfoDesignProduct(holdingCode, id)
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

// Get DesignProduct By Code godoc
// @Description get DesignProduct info by Code
// @Tags		AICloudProductDesign
// @Param		code  path      string  true  "DesignProduct Code"
// @Accept 		json
// @Success		200	{object} common.ApiResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/design/code/{code} [get]
func (h DesignProductHttp) InfoDesignProductByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	code := ctx.Param("code")
	doc, err := h.svc.InfoDesignProductByCode(holdingCode, code)
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

// List DesignProduct step godoc
// @Description get list step
// @Tags		AICloudProductDesign
// @Param		q	query	string	false  "Search Value"
// @Param		page	query	integer	false  "Page"
// @Param		limit	query	integer	false  "Limit"
// @Accept 		json
// @Success		200	{array} common.ApiResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/design [get]
func (h DesignProductHttp) SearchDesignProductPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageable := utils.GetPageable(ctx.QueryParam)
	docList, pagination, err := h.svc.SearchDesignProduct(holdingCode, map[string]interface{}{}, pageable)
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

// List DesignProduct godoc
// @Description search limit offset
// @Tags		AICloudProductDesign
// @Param		q	query	string	false  "Search Value"
// @Param		offset	query	integer	false  "offset"
// @Param		limit	query	integer	false  "limit"
// @Param		lang	query	string	false  "lang"
// @Accept 		json
// @Success		200	{array} common.ApiResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/design/list [get]
func (h DesignProductHttp) SearchDesignProductStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageableStep := utils.GetPageableStep(ctx.QueryParam)
	lang := ctx.QueryParam("lang")
	docList, total, err := h.svc.SearchDesignProductStep(holdingCode, lang, pageableStep)
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

// Create DesignProduct Bulk godoc
// @Description Create DesignProduct
// @Tags		AICloudProductDesign
// @Param		DesignProduct  body      []models.DesignProduct  true  "DesignProduct"
// @Accept 		json
// @Success		201	{object} common.BulkResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/design/bulk [post]
func (h DesignProductHttp) SaveBulk(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()
	dataReq := []models.DesignProduct{}
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
