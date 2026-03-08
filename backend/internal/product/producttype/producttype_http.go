package producttype

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/internal/product/producttype/models"
	"smlcloudplatform/internal/product/producttype/repositories"
	"smlcloudplatform/internal/product/producttype/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IProductTypeHttp interface{}

type ProductTypeHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IProductTypeHttpService
}

func NewProductTypeHttp(ms *microservice.Microservice, cfg config.IConfig) ProductTypeHttp {
	prod := ms.Producer(cfg.MQConfig())
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewProductTypeRepository(pst)
	repoProductBarcode := productbarcode_repositories.NewProductBarcodeRepository(pst, cache)
	repoMessageQueue := repositories.NewProductTypeMessageQueueRepository(prod)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewProductTypeHttpService(repo, repoProductBarcode, repoMessageQueue, masterSyncCacheRepo, 15*time.Second)

	return ProductTypeHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h ProductTypeHttp) RegisterHttp() {

	h.ms.POST("/product/type/bulk", h.SaveBulk)

	h.ms.GET("/product/type", h.SearchProductTypePage)
	h.ms.GET("/product/type/list", h.SearchProductTypeStep)
	h.ms.POST("/product/type", h.CreateProductType)
	h.ms.GET("/product/type/:id", h.InfoProductType)
	h.ms.GET("/product/type/code/:code", h.InfoProductTypeByCode)
	h.ms.PUT("/product/type/:id", h.UpdateProductType)
	h.ms.DELETE("/product/type/:id", h.DeleteProductType)
	h.ms.DELETE("/product/type", h.DeleteProductTypeByGUIDs)
}

// Create ProductType godoc
// @Description Create ProductType
// @Tags		ProductType
// @Param		ProductType  body      models.ProductType  true  "ProductType"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/type [post]
func (h ProductTypeHttp) CreateProductType(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.ProductType{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateProductType(shopID, authUsername, *docReq)

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

// Update ProductType godoc
// @Description Update ProductType
// @Tags		ProductType
// @Param		id  path      string  true  "ProductType ID"
// @Param		ProductType  body      models.ProductType  true  "ProductType"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/type/{id} [put]
func (h ProductTypeHttp) UpdateProductType(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.ProductType{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateProductType(shopID, id, authUsername, *docReq)

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

// Delete ProductType godoc
// @Description Delete ProductType
// @Tags		ProductType
// @Param		id  path      string  true  "ProductType ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/type/{id} [delete]
func (h ProductTypeHttp) DeleteProductType(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteProductType(shopID, id, authUsername)

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

// Delete ProductType godoc
// @Description Delete ProductType
// @Tags		ProductType
// @Param		ProductType  body      []string  true  "ProductType GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/type [delete]
func (h ProductTypeHttp) DeleteProductTypeByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteProductTypeByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get ProductType godoc
// @Description get ProductType info by guidfixed
// @Tags		ProductType
// @Param		id  path      string  true  "ProductType guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/type/{id} [get]
func (h ProductTypeHttp) InfoProductType(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get ProductType %v", id)
	doc, err := h.svc.InfoProductType(shopID, id)

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

// Get ProductType By Code godoc
// @Description get ProductType info by Code
// @Tags		ProductType
// @Param		code  path      string  true  "ProductType Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/type/code/{code} [get]
func (h ProductTypeHttp) InfoProductTypeByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")

	doc, err := h.svc.InfoProductTypeByCode(shopID, code)

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

// List ProductType step godoc
// @Description get list step
// @Tags		ProductType
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/type [get]
func (h ProductTypeHttp) SearchProductTypePage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.svc.SearchProductType(shopID, map[string]interface{}{}, pageable)

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

// List ProductType godoc
// @Description search limit offset
// @Tags		ProductType
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/type/list [get]
func (h ProductTypeHttp) SearchProductTypeStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	docList, total, err := h.svc.SearchProductTypeStep(shopID, lang, pageableStep)

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

// Create ProductType Bulk godoc
// @Description Create ProductType
// @Tags		ProductType
// @Param		ProductType  body      []models.ProductType  true  "ProductType"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/type/bulk [post]
func (h ProductTypeHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.ProductType{}
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
