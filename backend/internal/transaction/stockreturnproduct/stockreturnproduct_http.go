package stockreturnproduct

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	trans_cache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/transaction/stockreturnproduct/models"
	"smlcloudplatform/internal/transaction/stockreturnproduct/repositories"
	"smlcloudplatform/internal/transaction/stockreturnproduct/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

type IStockReturnProductHttp interface{}

type StockReturnProductHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IStockReturnProductService
}

func NewStockReturnProductHttp(ms *microservice.Microservice, cfg config.IConfig) StockReturnProductHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	producer := ms.Producer(cfg.MQConfig())

	repo := repositories.NewStockReturnProductRepository(pst)
	repoMq := repositories.NewStockReturnProductMessageQueueRepository(producer)

	productBarcodeRepo := productbarcode_repositories.NewProductBarcodeRepository(pst, cache)

	transCacheRepo := trans_cache.NewCacheRepository(cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewStockReturnProductService(repo, transCacheRepo, productBarcodeRepo, repoMq, masterSyncCacheRepo, services.StockReturnProductParser{})

	return StockReturnProductHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h StockReturnProductHttp) RegisterHttp() {

	h.ms.POST("/transaction/stock-return-product/bulk", h.SaveBulk)

	h.ms.GET("/transaction/stock-return-product", h.SearchStockReturnProductPage)
	h.ms.GET("/transaction/stock-return-product/list", h.SearchStockReturnProductStep)
	h.ms.POST("/transaction/stock-return-product", h.CreateStockReturnProduct)
	h.ms.GET("/transaction/stock-return-product/:id", h.InfoStockReturnProduct)
	h.ms.GET("/transaction/stock-return-product/code/:code", h.InfoStockReturnProductByCode)
	h.ms.PUT("/transaction/stock-return-product/:id", h.UpdateStockReturnProduct)
	h.ms.DELETE("/transaction/stock-return-product/:id", h.DeleteStockReturnProduct)
	h.ms.DELETE("/transaction/stock-return-product", h.DeleteStockReturnProductByGUIDs)
}

// Create StockReturnProduct godoc
// @Description Create StockReturnProduct
// @Tags		StockReturnProduct
// @Param		StockReturnProduct  body      models.StockReturnProduct  true  "StockReturnProduct"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-return-product [post]
func (h StockReturnProductHttp) CreateStockReturnProduct(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.StockReturnProduct{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, docNo, err := h.svc.CreateStockReturnProduct(shopID, authUsername, *docReq)

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

// Update StockReturnProduct godoc
// @Description Update StockReturnProduct
// @Tags		StockReturnProduct
// @Param		id  path      string  true  "StockReturnProduct ID"
// @Param		StockReturnProduct  body      models.StockReturnProduct  true  "StockReturnProduct"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-return-product/{id} [put]
func (h StockReturnProductHttp) UpdateStockReturnProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.StockReturnProduct{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateStockReturnProduct(shopID, id, authUsername, *docReq)

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

// Delete StockReturnProduct godoc
// @Description Delete StockReturnProduct
// @Tags		StockReturnProduct
// @Param		id  path      string  true  "StockReturnProduct ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-return-product/{id} [delete]
func (h StockReturnProductHttp) DeleteStockReturnProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteStockReturnProduct(shopID, id, authUsername)

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

// Delete StockReturnProduct godoc
// @Description Delete StockReturnProduct
// @Tags		StockReturnProduct
// @Param		StockReturnProduct  body      []string  true  "StockReturnProduct GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-return-product [delete]
func (h StockReturnProductHttp) DeleteStockReturnProductByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteStockReturnProductByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get StockReturnProduct godoc
// @Description get StockReturnProduct info by guidfixed
// @Tags		StockReturnProduct
// @Param		id  path      string  true  "StockReturnProduct guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-return-product/{id} [get]
func (h StockReturnProductHttp) InfoStockReturnProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get StockReturnProduct %v", id)
	doc, err := h.svc.InfoStockReturnProduct(shopID, id)

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

// Get StockReturnProduct By Code godoc
// @Description get StockReturnProduct info by Code
// @Tags		StockReturnProduct
// @Param		code  path      string  true  "StockReturnProduct Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-return-product/code/{code} [get]
func (h StockReturnProductHttp) InfoStockReturnProductByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")

	doc, err := h.svc.InfoStockReturnProductByCode(shopID, code)

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

// List StockReturnProduct step godoc
// @Description get list step
// @Tags		StockReturnProduct
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
// @Router /transaction/stock-return-product [get]
func (h StockReturnProductHttp) SearchStockReturnProductPage(ctx microservice.IContext) error {
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

	docList, pagination, err := h.svc.SearchStockReturnProduct(shopID, filters, pageable)

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

// List StockReturnProduct godoc
// @Description search limit offset
// @Tags		StockReturnProduct
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
// @Router /transaction/stock-return-product/list [get]
func (h StockReturnProductHttp) SearchStockReturnProductStep(ctx microservice.IContext) error {
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

	docList, total, err := h.svc.SearchStockReturnProductStep(shopID, lang, filters, pageableStep)

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

// Create StockReturnProduct Bulk godoc
// @Description Create StockReturnProduct
// @Tags		StockReturnProduct
// @Param		StockReturnProduct  body      []models.StockReturnProduct  true  "StockReturnProduct"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-return-product/bulk [post]
func (h StockReturnProductHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.StockReturnProduct{}
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
