package stockbalancedetail

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	trancache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/transaction/stockbalancedetail/models"
	"smlcloudplatform/internal/transaction/stockbalancedetail/repositories"
	"smlcloudplatform/internal/transaction/stockbalancedetail/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
)

type IStockBalanceDetailHttp interface{}

type StockBalanceDetailHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IStockBalanceDetailService
}

func NewStockBalanceDetailHttp(ms *microservice.Microservice, cfg config.IConfig) StockBalanceDetailHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	producer := ms.Producer(cfg.MQConfig())

	repo := repositories.NewStockBalanceDetailRepository(pst)
	repoMq := repositories.NewStockBalanceDetailMessageQueueRepository(producer)

	productBarcodeRepo := productbarcode_repositories.NewProductBarcodeRepository(pst, cache)

	transCacheRepo := trancache.NewCacheRepository(cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewStockBalanceDetailService(repo, transCacheRepo, productBarcodeRepo, repoMq, masterSyncCacheRepo, services.StockBalanceDetailParser{})

	return StockBalanceDetailHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h StockBalanceDetailHttp) RegisterHttp() {

	h.ms.GET("/transaction/stock-balance-detail/doc/:docno", h.SearchStockBalanceDetailPage)
	h.ms.GET("/transaction/stock-balance-detail/list/doc/:docno", h.SearchStockBalanceDetailStep)
	h.ms.POST("/transaction/stock-balance-detail", h.CreateStockBalanceDetail)
	h.ms.GET("/transaction/stock-balance-detail/:id", h.InfoStockBalanceDetail)
	h.ms.PUT("/transaction/stock-balance-detail/:id", h.UpdateStockBalanceDetail)
	h.ms.DELETE("/transaction/stock-balance-detail/:id", h.DeleteStockBalanceDetail)
	h.ms.DELETE("/transaction/stock-balance-detail", h.DeleteStockBalanceDetailByGUIDs)
	h.ms.DELETE("/transaction/stock-balance-detail/doc/:docno", h.DeleteStockBalanceDetailByDocNo)
}

// Create StockBalanceDetail godoc
// @Description Create StockBalanceDetail
// @Tags		StockBalanceDetail
// @Param		StockBalanceDetail  body      models.StockBalanceDetail  true  "StockBalanceDetail"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-balance-detail [post]
func (h StockBalanceDetailHttp) CreateStockBalanceDetail(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := []models.StockBalanceDetail{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.CreateStockBalanceDetail(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
	})
	return nil
}

// Update StockBalanceDetail godoc
// @Description Update StockBalanceDetail
// @Tags		StockBalanceDetail
// @Param		id  path      string  true  "StockBalanceDetail ID"
// @Param		StockBalanceDetail  body      models.StockBalanceDetail  true  "StockBalanceDetail"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-balance-detail/{id} [put]
func (h StockBalanceDetailHttp) UpdateStockBalanceDetail(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.StockBalanceDetail{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateStockBalanceDetail(shopID, id, authUsername, *docReq)

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

// Delete StockBalanceDetail godoc
// @Description Delete StockBalanceDetail
// @Tags		StockBalanceDetail
// @Param		id  path      string  true  "StockBalanceDetail ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-balance-detail/{id} [delete]
func (h StockBalanceDetailHttp) DeleteStockBalanceDetail(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteStockBalanceDetail(shopID, id, authUsername)

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

// Delete StockBalanceDetail godoc
// @Description Delete StockBalanceDetail
// @Tags		StockBalanceDetail
// @Param		StockBalanceDetail  body      []string  true  "StockBalanceDetail GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-balance-detail [delete]
func (h StockBalanceDetailHttp) DeleteStockBalanceDetailByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteStockBalanceDetailByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get StockBalanceDetail godoc
// @Description get StockBalanceDetail info by guidfixed
// @Tags		StockBalanceDetail
// @Param		id  path      string  true  "StockBalanceDetail guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-balance-detail/{id} [get]
func (h StockBalanceDetailHttp) InfoStockBalanceDetail(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get StockBalanceDetail %v", id)
	doc, err := h.svc.InfoStockBalanceDetail(shopID, id)

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

// List StockBalanceDetail step godoc
// @Description get list step
// @Tags		StockBalanceDetail
// @Param		docno  path      string  true  "DocNo"
// @Param		custcode	query	string		false  "customer code"
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-balance-detail/doc/{docno} [get]
func (h StockBalanceDetailHttp) SearchStockBalanceDetailPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	docNo := ctx.Param("docno")
	filters := map[string]interface{}{
		"docno": docNo,
	}

	docList, pagination, err := h.svc.SearchStockBalanceDetail(shopID, filters, pageable)

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

// List StockBalanceDetail godoc
// @Description search limit offset
// @Tags		StockBalanceDetail
// @Param		docno  path      string  true  "DocNo"
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-balance-detail/list/doc/{docno} [get]
func (h StockBalanceDetailHttp) SearchStockBalanceDetailStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	docNo := ctx.Param("docno")
	filters := map[string]interface{}{
		"docno": docNo,
	}

	docList, total, err := h.svc.SearchStockBalanceDetailStep(shopID, lang, filters, pageableStep)

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

// Delete StockBalanceDetail godoc
// @Description Delete StockBalanceDetail
// @Tags		StockBalanceDetail
// @Param		docno  path      string  true  "DocNo"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/stock-balance-detail/doc/{docno} [delete]
func (h StockBalanceDetailHttp) DeleteStockBalanceDetailByDocNo(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	docNo := ctx.Param("docno")

	err := h.svc.DeleteStockBalanceDetailByDocNo(shopID, authUsername, docNo)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}
