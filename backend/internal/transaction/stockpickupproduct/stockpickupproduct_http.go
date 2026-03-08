package stockpickupproduct

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	trans_cache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/transaction/stockpickupproduct/models"
	"smlcloudplatform/internal/transaction/stockpickupproduct/repositories"
	"smlcloudplatform/internal/transaction/stockpickupproduct/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

type IStockPickupProductHttp interface{}

type StockPickupProductHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IStockPickupProductService
}

func NewStockPickupProductHttp(ms *microservice.Microservice, cfg config.IConfig) StockPickupProductHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	producer := ms.Producer(cfg.MQConfig())

	repo := repositories.NewStockPickupProductRepository(pst)
	repoMq := repositories.NewStockPickupProductMessageQueueRepository(producer)

	productBarcodeRepo := productbarcode_repositories.NewProductBarcodeRepository(pst, cache)

	transCacheRepo := trans_cache.NewCacheRepository(cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewStockPickupProductService(repo, transCacheRepo, productBarcodeRepo, repoMq, masterSyncCacheRepo, services.StockPickupProductParser{})

	return StockPickupProductHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h StockPickupProductHttp) RegisterHttp() {

	h.ms.POST("/transaction/stock-prickup-product/bulk", h.SaveBulk)

	h.ms.GET("/transaction/stock-prickup-product", h.SearchStockPickupProductPage)
	h.ms.GET("/transaction/stock-prickup-product/list", h.SearchStockPickupProductStep)
	h.ms.POST("/transaction/stock-prickup-product", h.CreateStockPickupProduct)
	h.ms.GET("/transaction/stock-prickup-product/:id", h.InfoStockPickupProduct)
	h.ms.GET("/transaction/stock-prickup-product/code/:code", h.InfoStockPickupProductByCode)
	h.ms.PUT("/transaction/stock-prickup-product/:id", h.UpdateStockPickupProduct)
	h.ms.DELETE("/transaction/stock-prickup-product/:id", h.DeleteStockPickupProduct)
	h.ms.DELETE("/transaction/stock-prickup-product", h.DeleteStockPickupProductByGUIDs)
}

// Create StockPickupProduct godoc
// @Description Create StockPickupProduct
// @Tags		StockPickupProduct
// @Param		StockPickupProduct  body      models.StockPickupProduct  true  "StockPickupProduct"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-prickup-product [post]
func (h StockPickupProductHttp) CreateStockPickupProduct(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.StockPickupProduct{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, docNo, err := h.svc.CreateStockPickupProduct(shopID, authUsername, *docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      idx,
		Data:    docNo,
	})
	return nil
}

// Update StockPickupProduct godoc
// @Description Update StockPickupProduct
// @Tags		StockPickupProduct
// @Param		id  path      string  true  "StockPickupProduct ID"
// @Param		StockPickupProduct  body      models.StockPickupProduct  true  "StockPickupProduct"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-prickup-product/{id} [put]
func (h StockPickupProductHttp) UpdateStockPickupProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.StockPickupProduct{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateStockPickupProduct(shopID, id, authUsername, *docReq)

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

// Delete StockPickupProduct godoc
// @Description Delete StockPickupProduct
// @Tags		StockPickupProduct
// @Param		id  path      string  true  "StockPickupProduct ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-prickup-product/{id} [delete]
func (h StockPickupProductHttp) DeleteStockPickupProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteStockPickupProduct(shopID, id, authUsername)

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

// Delete StockPickupProduct godoc
// @Description Delete StockPickupProduct
// @Tags		StockPickupProduct
// @Param		StockPickupProduct  body      []string  true  "StockPickupProduct GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-prickup-product [delete]
func (h StockPickupProductHttp) DeleteStockPickupProductByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteStockPickupProductByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get StockPickupProduct godoc
// @Description get StockPickupProduct info by guidfixed
// @Tags		StockPickupProduct
// @Param		id  path      string  true  "StockPickupProduct guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-prickup-product/{id} [get]
func (h StockPickupProductHttp) InfoStockPickupProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get StockPickupProduct %v", id)
	doc, err := h.svc.InfoStockPickupProduct(shopID, id)

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

// Get StockPickupProduct By Code godoc
// @Description get StockPickupProduct info by Code
// @Tags		StockPickupProduct
// @Param		code  path      string  true  "StockPickupProduct Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-prickup-product/code/{code} [get]
func (h StockPickupProductHttp) InfoStockPickupProductByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")

	doc, err := h.svc.InfoStockPickupProductByCode(shopID, code)

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

// List StockPickupProduct step godoc
// @Description get list step
// @Tags		StockPickupProduct
// @Param		q		query	string		false  "Search Value"
// @Param		custcode	query	string		false  "cust code"
// @Param		branchcode	query	string		false  "branch code"
// @Param		fromdate	query	string		false  "from date"
// @Param		todate	query	string		false  "to date"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-prickup-product [get]
func (h StockPickupProductHttp) SearchStockPickupProductPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "custcode",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "-",
			Field: "docdatetime",
			Type:  requestfilter.FieldTypeRangeDate,
		},
		{
			Param: "branchcode",
			Field: "branch.code",
			Type:  requestfilter.FieldTypeString,
		},
	})

	docList, pagination, err := h.svc.SearchStockPickupProduct(shopID, filters, pageable)

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

// List StockPickupProduct godoc
// @Description search limit offset
// @Tags		StockPickupProduct
// @Param		q		query	string		false  "Search Value"
// @Param		custcode	query	string		false  "cust code"
// @Param		branchcode	query	string		false  "branch code"
// @Param		fromdate	query	string		false  "from date"
// @Param		todate	query	string		false  "to date"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-prickup-product/list [get]
func (h StockPickupProductHttp) SearchStockPickupProductStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "custcode",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "-",
			Field: "docdatetime",
			Type:  requestfilter.FieldTypeRangeDate,
		},
		{
			Param: "branchcode",
			Field: "branch.code",
			Type:  requestfilter.FieldTypeString,
		},
	})

	docList, total, err := h.svc.SearchStockPickupProductStep(shopID, lang, filters, pageableStep)

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

// Create StockPickupProduct Bulk godoc
// @Description Create StockPickupProduct
// @Tags		StockPickupProduct
// @Param		StockPickupProduct  body      []models.StockPickupProduct  true  "StockPickupProduct"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-prickup-product/bulk [post]
func (h StockPickupProductHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.StockPickupProduct{}
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
