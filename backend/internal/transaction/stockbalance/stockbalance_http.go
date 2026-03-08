package stockbalance

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	trancache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/transaction/stockbalance/models"
	"smlcloudplatform/internal/transaction/stockbalance/repositories"
	"smlcloudplatform/internal/transaction/stockbalance/services"
	stockbalancedetail_repositories "smlcloudplatform/internal/transaction/stockbalancedetail/repositories"
	stockbalancedetail_services "smlcloudplatform/internal/transaction/stockbalancedetail/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

type IStockBalanceHttp interface{}

type StockBalanceHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IStockBalanceHttpService
}

func NewStockBalanceHttp(ms *microservice.Microservice, cfg config.IConfig) StockBalanceHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	producer := ms.Producer(cfg.MQConfig())

	repo := repositories.NewStockBalanceRepository(pst)
	repoMq := repositories.NewStockBalanceMessageQueueRepository(producer)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	transCacheRepo := trancache.NewCacheRepository(cache)

	productBarcodeRepo := productbarcode_repositories.NewProductBarcodeRepository(pst, cache)

	repoStockBalanceDetail := stockbalancedetail_repositories.NewStockBalanceDetailRepository(pst)
	repoMqStockBalanceDetail := stockbalancedetail_repositories.NewStockBalanceDetailMessageQueueRepository(producer)

	svcStockBalanceDetail := stockbalancedetail_services.NewStockBalanceDetailService(
		repoStockBalanceDetail,
		transCacheRepo,
		productBarcodeRepo,
		repoMqStockBalanceDetail,
		masterSyncCacheRepo,
		stockbalancedetail_services.StockBalanceDetailParser{},
	)

	svc := services.NewStockBalanceHttpService(svcStockBalanceDetail, repo, transCacheRepo, repoMq, masterSyncCacheRepo)

	return StockBalanceHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h StockBalanceHttp) RegisterHttp() {

	h.ms.POST("/transaction/stock-balance/bulk", h.SaveBulk)

	h.ms.GET("/transaction/stock-balance", h.SearchStockBalancePage)
	h.ms.GET("/transaction/stock-balance/list", h.SearchStockBalanceStep)
	h.ms.POST("/transaction/stock-balance", h.CreateStockBalance)
	h.ms.GET("/transaction/stock-balance/:id", h.InfoStockBalance)
	h.ms.GET("/transaction/stock-balance/code/:code", h.InfoStockBalanceByCode)
	h.ms.PUT("/transaction/stock-balance/:id", h.UpdateStockBalance)
	h.ms.DELETE("/transaction/stock-balance/:id", h.DeleteStockBalance)
	h.ms.DELETE("/transaction/stock-balance", h.DeleteStockBalanceByGUIDs)
}

// Create StockBalance godoc
// @Description Create StockBalance
// @Tags		StockBalance
// @Param		StockBalance  body      models.StockBalance  true  "StockBalance"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-balance [post]
func (h StockBalanceHttp) CreateStockBalance(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.StockBalance{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	_, idx, docNo, err := h.svc.CreateStockBalance(shopID, authUsername, *docReq)

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

// Update StockBalance godoc
// @Description Update StockBalance
// @Tags		StockBalance
// @Param		id  path      string  true  "StockBalance ID"
// @Param		StockBalance  body      models.StockBalance  true  "StockBalance"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-balance/{id} [put]
func (h StockBalanceHttp) UpdateStockBalance(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.StockBalance{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateStockBalance(shopID, id, authUsername, *docReq)

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

// Delete StockBalance godoc
// @Description Delete StockBalance
// @Tags		StockBalance
// @Param		id  path      string  true  "StockBalance ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-balance/{id} [delete]
func (h StockBalanceHttp) DeleteStockBalance(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteStockBalance(shopID, id, authUsername)

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

// Delete StockBalance godoc
// @Description Delete StockBalance
// @Tags		StockBalance
// @Param		StockBalance  body      []string  true  "StockBalance GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-balance [delete]
func (h StockBalanceHttp) DeleteStockBalanceByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteStockBalanceByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get StockBalance godoc
// @Description get StockBalance info by guidfixed
// @Tags		StockBalance
// @Param		id  path      string  true  "StockBalance guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-balance/{id} [get]
func (h StockBalanceHttp) InfoStockBalance(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get StockBalance %v", id)
	doc, err := h.svc.InfoStockBalance(shopID, id)

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

// Get StockBalance By Code godoc
// @Description get StockBalance info by Code
// @Tags		StockBalance
// @Param		code  path      string  true  "StockBalance Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-balance/code/{code} [get]
func (h StockBalanceHttp) InfoStockBalanceByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")

	doc, err := h.svc.InfoStockBalanceByCode(shopID, code)

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

// List StockBalance step godoc
// @Description get list step
// @Tags		StockBalance
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
// @Router /transaction/stock-balance [get]
func (h StockBalanceHttp) SearchStockBalancePage(ctx microservice.IContext) error {
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

	docList, pagination, err := h.svc.SearchStockBalance(shopID, filters, pageable)

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

// List StockBalance godoc
// @Description search limit offset
// @Tags		StockBalance
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
// @Router /transaction/stock-balance/list [get]
func (h StockBalanceHttp) SearchStockBalanceStep(ctx microservice.IContext) error {
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

	docList, total, err := h.svc.SearchStockBalanceStep(shopID, lang, filters, pageableStep)

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

// Create StockBalance Bulk godoc
// @Description Create StockBalance
// @Tags		StockBalance
// @Param		StockBalance  body      []models.StockBalance  true  "StockBalance"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-balance/bulk [post]
func (h StockBalanceHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.StockBalance{}
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
