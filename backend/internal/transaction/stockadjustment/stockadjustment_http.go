package stockadjustment

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	trancache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/transaction/stockadjustment/models"
	"smlcloudplatform/internal/transaction/stockadjustment/repositories"
	"smlcloudplatform/internal/transaction/stockadjustment/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

type IStockAdjustmentHttp interface{}

type StockAdjustmentHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IStockAdjustmentService
}

func NewStockAdjustmentHttp(ms *microservice.Microservice, cfg config.IConfig) StockAdjustmentHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	producer := ms.Producer(cfg.MQConfig())

	repo := repositories.NewStockAdjustmentRepository(pst)
	repoMq := repositories.NewStockAdjustmentMessageQueueRepository(producer)

	productBarcodeRepo := productbarcode_repositories.NewProductBarcodeRepository(pst, cache)

	transRepo := trancache.NewCacheRepository(cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewStockAdjustmentService(repo, transRepo, productBarcodeRepo, repoMq, masterSyncCacheRepo, services.StockAdjustmenParser{})

	return StockAdjustmentHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h StockAdjustmentHttp) RegisterHttp() {

	h.ms.POST("/transaction/stock-adjustment/bulk", h.SaveBulk)

	h.ms.GET("/transaction/stock-adjustment", h.SearchStockAdjustmentPage)
	h.ms.GET("/transaction/stock-adjustment/list", h.SearchStockAdjustmentStep)
	h.ms.POST("/transaction/stock-adjustment", h.CreateStockAdjustment)
	h.ms.GET("/transaction/stock-adjustment/:id", h.InfoStockAdjustment)
	h.ms.GET("/transaction/stock-adjustment/code/:code", h.InfoStockAdjustmentByCode)
	h.ms.PUT("/transaction/stock-adjustment/:id", h.UpdateStockAdjustment)
	h.ms.DELETE("/transaction/stock-adjustment/:id", h.DeleteStockAdjustment)
	h.ms.DELETE("/transaction/stock-adjustment", h.DeleteStockAdjustmentByGUIDs)
}

// Create StockAdjustment godoc
// @Description Create StockAdjustment
// @Tags		StockAdjustment
// @Param		StockAdjustment  body      models.StockAdjustment  true  "StockAdjustment"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-adjustment [post]
func (h StockAdjustmentHttp) CreateStockAdjustment(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.StockAdjustment{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, docNo, err := h.svc.CreateStockAdjustment(shopID, authUsername, *docReq)

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

// Update StockAdjustment godoc
// @Description Update StockAdjustment
// @Tags		StockAdjustment
// @Param		id  path      string  true  "StockAdjustment ID"
// @Param		StockAdjustment  body      models.StockAdjustment  true  "StockAdjustment"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-adjustment/{id} [put]
func (h StockAdjustmentHttp) UpdateStockAdjustment(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.StockAdjustment{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateStockAdjustment(shopID, id, authUsername, *docReq)

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

// Delete StockAdjustment godoc
// @Description Delete StockAdjustment
// @Tags		StockAdjustment
// @Param		id  path      string  true  "StockAdjustment ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-adjustment/{id} [delete]
func (h StockAdjustmentHttp) DeleteStockAdjustment(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteStockAdjustment(shopID, id, authUsername)

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

// Delete StockAdjustment godoc
// @Description Delete StockAdjustment
// @Tags		StockAdjustment
// @Param		StockAdjustment  body      []string  true  "StockAdjustment GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-adjustment [delete]
func (h StockAdjustmentHttp) DeleteStockAdjustmentByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteStockAdjustmentByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get StockAdjustment godoc
// @Description get StockAdjustment info by guidfixed
// @Tags		StockAdjustment
// @Param		id  path      string  true  "StockAdjustment guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-adjustment/{id} [get]
func (h StockAdjustmentHttp) InfoStockAdjustment(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get StockAdjustment %v", id)
	doc, err := h.svc.InfoStockAdjustment(shopID, id)

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

// Get StockAdjustment By Code godoc
// @Description get StockAdjustment info by Code
// @Tags		StockAdjustment
// @Param		code  path      string  true  "StockAdjustment Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-adjustment/code/{code} [get]
func (h StockAdjustmentHttp) InfoStockAdjustmentByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")

	doc, err := h.svc.InfoStockAdjustmentByCode(shopID, code)

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

// List StockAdjustment step godoc
// @Description get list step
// @Tags		StockAdjustment
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
// @Router /transaction/stock-adjustment [get]
func (h StockAdjustmentHttp) SearchStockAdjustmentPage(ctx microservice.IContext) error {
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

	docList, pagination, err := h.svc.SearchStockAdjustment(shopID, filters, pageable)

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

// List StockAdjustment godoc
// @Description search limit offset
// @Tags		StockAdjustment
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
// @Router /transaction/stock-adjustment/list [get]
func (h StockAdjustmentHttp) SearchStockAdjustmentStep(ctx microservice.IContext) error {
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

	docList, total, err := h.svc.SearchStockAdjustmentStep(shopID, lang, filters, pageableStep)

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

// Create StockAdjustment Bulk godoc
// @Description Create StockAdjustment
// @Tags		StockAdjustment
// @Param		StockAdjustment  body      []models.StockAdjustment  true  "StockAdjustment"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-adjustment/bulk [post]
func (h StockAdjustmentHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.StockAdjustment{}
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
