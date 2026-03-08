package productgroup

import (
	"encoding/json"
	"net/http"
	"net/url"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/internal/product/productgroup/models"
	"smlcloudplatform/internal/product/productgroup/repositories"
	"smlcloudplatform/internal/product/productgroup/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
)

type IProductGroupHttp interface{}

type ProductGroupHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IProductGroupHttpService
}

func NewProductGroupHttp(ms *microservice.Microservice, cfg config.IConfig) ProductGroupHttp {
	prod := ms.Producer(cfg.MQConfig())
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewProductGroupRepository(pst)
	repoProductBarcode := productbarcode_repositories.NewProductBarcodeRepository(pst, cache)
	repoMessageQueue := repositories.NewProductGroupMessageQueueRepository(prod)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewProductGroupHttpService(repo, repoMessageQueue, repoProductBarcode, cfg.ProductGroupServiceConfig(), masterSyncCacheRepo)

	return ProductGroupHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h ProductGroupHttp) RegisterHttp() {

	h.ms.POST("/product/group/bulk", h.SaveBulk)

	h.ms.GET("/product/group", h.SearchProductGroupPage)
	h.ms.GET("/product/group/list", h.SearchProductGroupStep)
	h.ms.POST("/product/group", h.CreateProductGroup)
	h.ms.POST("/product/group/save", h.SaveProductGroup)
	h.ms.GET("/product/group/:id", h.InfoProductGroup)
	h.ms.GET("/product/group/by-code", h.InfoArray)
	h.ms.PUT("/product/group/:id", h.UpdateProductGroup)
	h.ms.DELETE("/product/group/:id", h.DeleteProductGroup)
	h.ms.DELETE("/product/group", h.DeleteProductGroupByGUIDs)
}

// Create ProductGroup godoc
// @Description Create ProductGroup
// @Tags		ProductGroup
// @Param		ProductGroup  body      models.ProductGroup  true  "ProductGroup"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/group [post]
func (h ProductGroupHttp) CreateProductGroup(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.ProductGroup{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateProductGroup(shopID, authUsername, *docReq)

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

// Save ProductGroup godoc
// @Description Save ProductGroup
// @Tags		ProductGroup
// @Param		ProductGroup  body      models.ProductGroup  true  "ProductGroup"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/group/save [post]
func (h ProductGroupHttp) SaveProductGroup(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.ProductGroup{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.SaveProductGroup(shopID, authUsername, *docReq)

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

// Update ProductGroup godoc
// @Description Update ProductGroup
// @Tags		ProductGroup
// @Param		id  path      string  true  "ProductGroup ID"
// @Param		ProductGroup  body      models.ProductGroup  true  "ProductGroup"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/group/{id} [put]
func (h ProductGroupHttp) UpdateProductGroup(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.ProductGroup{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateProductGroup(shopID, id, authUsername, *docReq)

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

// Delete ProductGroup godoc
// @Description Delete ProductGroup
// @Tags		ProductGroup
// @Param		id  path      string  true  "ProductGroup ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/group/{id} [delete]
func (h ProductGroupHttp) DeleteProductGroup(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteProductGroup(shopID, id, authUsername)

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

// Delete ProductGroup godoc
// @Description Delete ProductGroup
// @Tags		ProductGroup
// @Param		ProductGroup  body      []string  true  "ProductGroup GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/group [delete]
func (h ProductGroupHttp) DeleteProductGroupByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteProductGroupByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get ProductGroup godoc
// @Description get struct array by ID
// @Tags		ProductGroup
// @Param		id  path      string  true  "ProductGroup ID"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/group/{id} [get]
func (h ProductGroupHttp) InfoProductGroup(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get ProductGroup %v", id)
	doc, err := h.svc.InfoProductGroup(shopID, id)

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

// Get Product Group By code array godoc
// @Description get Product Group by code array
// @Tags		Unit
// @Param		codes	query	string		false  "Code filter, json array encode "
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/group/by-code [get]
func (h ProductGroupHttp) InfoArray(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	codesReq, err := url.QueryUnescape(ctx.QueryParam("codes"))

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	docReq := []string{}
	err = json.Unmarshal([]byte(codesReq), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}
	// where to filter array
	doc, err := h.svc.InfoWTFArray(shopID, docReq)

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

// List ProductGroup godoc
// @Description get struct array by ID
// @Tags		ProductGroup
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "page"
// @Param		limit	query	integer		false  "limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/group [get]
func (h ProductGroupHttp) SearchProductGroupPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.svc.SearchProductGroup(shopID, map[string]interface{}{}, pageable)

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

// List ProductGroup godoc
// @Description search limit offset
// @Tags		ProductGroup
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/group/list [get]
func (h ProductGroupHttp) SearchProductGroupStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	docList, total, err := h.svc.SearchProductGroupStep(shopID, lang, pageableStep)

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

// Create ProductGroup Bulk godoc
// @Description Create ProductGroup
// @Tags		ProductGroup
// @Param		ProductGroup  body      []models.ProductGroup  true  "ProductGroup"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/group/bulk [post]
func (h ProductGroupHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.ProductGroup{}
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
