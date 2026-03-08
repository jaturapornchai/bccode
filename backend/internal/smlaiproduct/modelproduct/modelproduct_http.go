// modelproduct_http.go
package modelproduct

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/internal/smlaiproduct/modelproduct/models"
	modelrepo "smlcloudplatform/internal/smlaiproduct/modelproduct/repositories"
	modelsvc "smlcloudplatform/internal/smlaiproduct/modelproduct/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IModelProductHttp interface{}

type ModelProductHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc modelsvc.IModelProductHttpService
}

func NewModelProductHttp(ms *microservice.Microservice, cfg config.IConfig) ModelProductHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := modelrepo.NewModelProductRepository(pst)
	repoProductBarcode := productbarcode_repositories.NewProductBarcodeRepository(pst, cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)

	svc := modelsvc.NewModelProductHttpService(repo, repoProductBarcode, masterSyncCacheRepo, 15*time.Second)

	return ModelProductHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h ModelProductHttp) RegisterHttp() {
	h.ms.POST("/aicloud/model/bulk", h.SaveBulk)
	h.ms.GET("/aicloud/model", h.SearchModelProductPage)
	h.ms.GET("/aicloud/model/list", h.SearchModelProductStep)
	h.ms.POST("/aicloud/model", h.CreateModelProduct)
	h.ms.GET("/aicloud/model/:id", h.InfoModelProduct)
	h.ms.GET("/aicloud/model/code/:code", h.InfoModelProductByCode)
	h.ms.PUT("/aicloud/model/:id", h.UpdateModelProduct)
	h.ms.DELETE("/aicloud/model/:id", h.DeleteModelProduct)
	h.ms.DELETE("/aicloud/model", h.DeleteModelProductByGUIDs)
}

// Create ModelProduct godoc
// @Description Create ModelProduct
// @Tags		AICloudProductModel
// @Param		ModelProduct	body	models.ModelProduct	true	"ModelProduct"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401	{object}	common.AuthResponseFailed
// @Security		AccessToken
// @Router /aicloud/model [post]
func (h ModelProductHttp) CreateModelProduct(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()
	docReq := &models.ModelProduct{}
	if err := json.Unmarshal([]byte(input), docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err := ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateModelProduct(shopID, authUsername, *docReq)
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

// Update ModelProduct godoc
// @Description Update ModelProduct
// @Tags		AICloudProductModel
// @Param		id	path	string	true	"ModelProduct ID"
// @Param		ModelProduct	body	models.ModelProduct	true	"ModelProduct"
// @Accept		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401	{object}	common.AuthResponseFailed
// @Security		AccessToken
// @Router /aicloud/model/{id} [put]
func (h ModelProductHttp) UpdateModelProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()
	docReq := &models.ModelProduct{}
	if err := json.Unmarshal([]byte(input), docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err := ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err := h.svc.UpdateModelProduct(shopID, id, authUsername, *docReq)
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

// Delete ModelProduct godoc
// @Description Delete ModelProduct
// @Tags		AICloudProductModel
// @Param		id  path      string  true  "ModelProduct ID"
// @Accept 		json
// @Success		200	{object} common.ResponseSuccessWithID
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/model/{id} [delete]
func (h ModelProductHttp) DeleteModelProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteModelProduct(shopID, id, authUsername)
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

// Delete ModelProduct godoc
// @Description Delete ModelProduct
// @Tags		AICloudProductModel
// @Param		ModelProduct  body      []string  true  "ModelProduct GUIDs"
// @Accept 		json
// @Success		200	{object} common.ResponseSuccessWithID
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/model [delete]
func (h ModelProductHttp) DeleteModelProductByGUIDs(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	input := ctx.ReadInput()
	docReq := []string{}
	if err := json.Unmarshal([]byte(input), &docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err := h.svc.DeleteModelProductByGUIDs(shopID, authUsername, docReq)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})
	return nil
}

// Get ModelProduct godoc
// @Description get ModelProduct info by guidfixed
// @Tags		AICloudProductModel
// @Param		id  path      string  true  "ModelProduct guidfixed"
// @Accept 		json
// @Success		200	{object} common.ApiResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/model/{id} [get]
func (h ModelProductHttp) InfoModelProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get ModelProduct %v", id)
	doc, err := h.svc.InfoModelProduct(shopID, id)
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

// Get ModelProduct By Code godoc
// @Description get ModelProduct info by Code
// @Tags		AICloudProductModel
// @Param		code  path      string  true  "ModelProduct Code"
// @Accept 		json
// @Success		200	{object} common.ApiResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/model/code/{code} [get]
func (h ModelProductHttp) InfoModelProductByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")
	doc, err := h.svc.InfoModelProductByCode(shopID, code)
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

// List ModelProduct step godoc
// @Description get list step
// @Tags		AICloudProductModel
// @Param		q	query	string	false  "Search Value"
// @Param		page	query	integer	false  "Page"
// @Param		limit	query	integer	false  "Limit"
// @Accept 		json
// @Success		200	{array} common.ApiResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/model [get]
func (h ModelProductHttp) SearchModelProductPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)
	docList, pagination, err := h.svc.SearchModelProduct(shopID, map[string]interface{}{}, pageable)
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

// List ModelProduct godoc
// @Description search limit offset
// @Tags		AICloudProductModel
// @Param		q	query	string	false  "Search Value"
// @Param		offset	query	integer	false  "offset"
// @Param		limit	query	integer	false  "limit"
// @Param		lang	query	string	false  "lang"
// @Accept 		json
// @Success		200	{array} common.ApiResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/model/list [get]
func (h ModelProductHttp) SearchModelProductStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)
	lang := ctx.QueryParam("lang")
	docList, total, err := h.svc.SearchModelProductStep(shopID, lang, pageableStep)
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

// Create ModelProduct Bulk godoc
// @Description Create ModelProduct
// @Tags		AICloudProductModel
// @Param		ModelProduct  body      []models.ModelProduct  true  "ModelProduct"
// @Accept 		json
// @Success		201	{object} common.BulkResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/model/bulk [post]
func (h ModelProductHttp) SaveBulk(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()
	dataReq := []models.ModelProduct{}
	if err := json.Unmarshal([]byte(input), &dataReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	bulkResponse, err := h.svc.SaveInBatch(shopID, authUsername, dataReq)
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
