package stockbalanceimport

import (
	"encoding/json"
	"errors"
	"net/http"
	"path/filepath"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_repo "smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/internal/stockbalanceimport/models"
	"smlcloudplatform/internal/stockbalanceimport/repositories"
	"smlcloudplatform/internal/stockbalanceimport/services"
	trancache "smlcloudplatform/internal/transaction/repositories"
	stockbalance_models "smlcloudplatform/internal/transaction/stockbalance/models"
	stockbalance_repositories "smlcloudplatform/internal/transaction/stockbalance/repositories"
	stockbalance_serrvices "smlcloudplatform/internal/transaction/stockbalance/services"
	stockbalancedetail_serrvices "smlcloudplatform/internal/transaction/stockbalancedetail/services"
	"smlcloudplatform/pkg/microservice"
	"time"

	stockbalancedetail_repositories "smlcloudplatform/internal/transaction/stockbalancedetail/repositories"
	"smlcloudplatform/internal/utils"
)

type IStockBalanceImportHttp interface{}

type StockBalanceImportHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IStockBalanceImportService
}

func NewStockBalanceImportHttp(ms *microservice.Microservice, cfg config.IConfig) StockBalanceImportHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	producer := ms.Producer(cfg.MQConfig())
	pstClickHouse := ms.ClickHousePersister(cfg.ClickHouseConfig())

	repo := stockbalance_repositories.NewStockBalanceRepository(pst)
	repoMq := stockbalance_repositories.NewStockBalanceMessageQueueRepository(producer)

	transCacheRepo := trancache.NewCacheRepository(cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	stockbalanceDetailRepo := stockbalancedetail_repositories.NewStockBalanceDetailRepository(pst)
	stockbalanceDetailMqRepo := stockbalancedetail_repositories.NewStockBalanceDetailMessageQueueRepository(producer)

	productBarcodeRepo := productbarcode_repo.NewProductBarcodeRepository(pst, cache)

	stockBalanceDetailSvc := stockbalancedetail_serrvices.NewStockBalanceDetailService(
		stockbalanceDetailRepo,
		transCacheRepo,
		productBarcodeRepo,
		stockbalanceDetailMqRepo,
		masterSyncCacheRepo,
		stockbalancedetail_serrvices.StockBalanceDetailParser{},
	)

	stockBalanceSvc := stockbalance_serrvices.NewStockBalanceHttpService(stockBalanceDetailSvc, repo, transCacheRepo, repoMq, masterSyncCacheRepo)

	chRepo := repositories.NewStockBalanceImportClickHouseRepository(pstClickHouse)

	svc := services.NewStockBalanceImportService(chRepo, productBarcodeRepo, stockBalanceSvc, stockBalanceDetailSvc, utils.RandStringBytesMaskImprSrcUnsafe, utils.NewGUID, time.Now)

	return StockBalanceImportHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h StockBalanceImportHttp) RegisterHttp() {
	h.ms.POST("/stockbalanceimport/upload", h.UploadExcel)
	h.ms.POST("/stockbalanceimport", h.Create)
	h.ms.GET("/stockbalanceimport/:task-id", h.List)
	h.ms.DELETE("/stockbalanceimport/:task-id", h.DeleteByTask)
	h.ms.POST("/stockbalanceimport/:task-id", h.SaveTask)
	h.ms.GET("/stockbalanceimport/:task-id/meta", h.Meta)
	h.ms.POST("/stockbalanceimport/:task-id/verify", h.Verify)
	h.ms.PUT("/stockbalanceimport/item/:guid", h.Update)
	h.ms.DELETE("/stockbalanceimport/item/:guid", h.Delete)
}

// Create StockBalanceImport godoc
// @Description Create StockBalanceImport
// @Tags		StockBalanceImport
// @Param		file  formData      file  true  "excel file"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /stockbalanceimport/upload [post]
func (h StockBalanceImportHttp) UploadExcel(ctx microservice.IContext) error {
	shopID := ctx.UserInfo().ShopID
	authUsername := ctx.UserInfo().Username
	tempFile, err := ctx.FormFile("file")

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	// Check if the file is an Excel file
	if filepath.Ext(tempFile.Filename) != ".xlsx" {
		ctx.ResponseError(400, "Invalid file xlsx type.")
		return errors.New("invalid file type")
	}

	file, err := tempFile.Open()

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}
	defer file.Close()

	taskID, err := h.svc.ImportFromFile(shopID, authUsername, file)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      taskID,
	})

	return nil
}

// Create StockBalanceImport godoc
// @Description Create StockBalanceImport
// @Tags		StockBalanceImport
// @Param		StockBalanceImport  body      models.StockBalanceImport  true  "StockBalanceImport"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /stockbalanceimport [post]
func (h StockBalanceImportHttp) Create(ctx microservice.IContext) error {
	shopID := ctx.UserInfo().ShopID
	authUsername := ctx.UserInfo().Username

	input := ctx.ReadInput()

	docReq := models.StockBalanceImport{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.Create(shopID, authUsername, &docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
	})
	return nil
}

// List StockBalanceImport godoc
// @Description List StockBalanceImport
// @Tags		StockBalanceImport
// @Param		task-id		path		string		true		"task id"
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /stockbalanceimport/{task-id} [get]
func (h StockBalanceImportHttp) List(ctx microservice.IContext) error {
	shopID := ctx.UserInfo().ShopID

	taskID := ctx.Param("task-id")

	if taskID == "" {
		ctx.Response(http.StatusCreated, common.ApiResponse{
			Success: true,
			Data:    []string{},
		})
		return nil
	}

	pageable := utils.GetPageable(ctx.QueryParam)

	filters := map[string]interface{}{}

	isExistRaw := ctx.QueryParam("exist")
	isExist := false
	if isExistRaw == "true" {
		isExist = true
	}

	if isExistRaw != "" {
		filters["exist"] = isExist
	}

	results, page, err := h.svc.List(shopID, taskID, filters, pageable)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success:    true,
		Pagination: page,
		Data:       results,
	})
	return nil
}

// Get StockBalanceImport Part godoc
// @Description Get StockBalanceImport Part
// @Tags		StockBalanceImport
// @Param		guid		path		string		true		"guid"
// @Accept 		json
// @Success		201	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /stockbalanceimport/item/{guid} [put]
func (h StockBalanceImportHttp) Update(ctx microservice.IContext) error {
	shopID := ctx.UserInfo().ShopID

	input := ctx.ReadInput()

	guid := ctx.Param("guid")

	docReq := models.StockBalanceImportRaw{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.Update(shopID, guid, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
	})
	return nil
}

// Delete StockBalanceImport By GUID godoc
// @Description Delete StockBalanceImport By GUID
// @Tags		StockBalanceImport
// @Param		guid		path		string		true		"guid"
// @Accept 		json
// @Success		201	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /stockbalanceimport/item/{guid} [delete]
func (h StockBalanceImportHttp) Delete(ctx microservice.IContext) error {
	shopID := ctx.UserInfo().ShopID

	guid := ctx.Param("guid")

	err := h.svc.Delete(shopID, guid)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
	})
	return nil
}

// Delete StockBalanceImport By Task ID godoc
// @Description Delete StockBalanceImport By Task ID
// @Tags		StockBalanceImport
// @Param		task-id		path		string		true		"task id"
// @Accept 		json
// @Success		201	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /stockbalanceimport/{task-id} [delete]
func (h StockBalanceImportHttp) DeleteByTask(ctx microservice.IContext) error {
	shopID := ctx.UserInfo().ShopID

	taskID := ctx.Param("task-id")

	err := h.svc.DeleteTask(shopID, taskID)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
	})
	return nil
}

// Delete StockBalanceImport By Task ID godoc
// @Description Delete StockBalanceImport By Task ID
// @Tags		StockBalanceImport
// @Param		task-id		path		string		true		"task id"
// @Param		StockBalanceHeader  body      models.StockBalanceHeader  true  "Stock Balance Header"
// @Accept 		json
// @Success		201	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /stockbalanceimport/{task-id} [post]
func (h StockBalanceImportHttp) SaveTask(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	taskID := ctx.Param("task-id")

	input := ctx.ReadInput()
	docReq := stockbalance_models.StockBalanceHeader{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	docNo, err := h.svc.SaveTask(shopID, authUsername, taskID, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		DocNo:   docNo,
	})
	return nil
}

// Get StockBalanceImport Meta godoc
// @Description Get StockBalanceImport Meta
// @Tags		StockBalanceImport
// @Param		task-id		path		string		true		"task id"
// @Accept 		json
// @Success		201	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /stockbalanceimport/{task-id}/meta [get]
func (h StockBalanceImportHttp) Meta(ctx microservice.IContext) error {
	shopID := ctx.UserInfo().ShopID

	taskID := ctx.Param("task-id")

	result, err := h.svc.Meta(shopID, taskID)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		Data:    result,
	})
	return nil
}

// Verify StockBalanceImport By Task ID godoc
// @Description Verify StockBalanceImport By Task ID
// @Tags		StockBalanceImport
// @Param		task-id		path		string		true		"task id"
// @Accept 		json
// @Success		201	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /stockbalanceimport/{task-id}/verify [post]
func (h StockBalanceImportHttp) Verify(ctx microservice.IContext) error {
	shopID := ctx.UserInfo().ShopID

	guid := ctx.Param("task-id")

	err := h.svc.Verify(shopID, guid)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
	})
	return nil
}
