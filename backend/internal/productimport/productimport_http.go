package productimport

import (
	"encoding/json"
	"errors"
	"net/http"
	"path/filepath"
	"smlcloudplatform/internal/config"
	creditorRepo "smlcloudplatform/internal/debtaccount/creditor/repositories"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	branch_repositories "smlcloudplatform/internal/organization/branch/repositories"
	businesstype_repositories "smlcloudplatform/internal/organization/businesstype/repositories"
	productmaster "smlcloudplatform/internal/product/product/repositories"
	product_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	product_serrvices "smlcloudplatform/internal/product/productbarcode/services"
	productcategory_repositories "smlcloudplatform/internal/product/productcategory/repositories"
	productcategory_services "smlcloudplatform/internal/product/productcategory/services"
	productunit_repo "smlcloudplatform/internal/product/unit/repositories"
	unit_repositories "smlcloudplatform/internal/product/unit/repositories"
	unitmaster "smlcloudplatform/internal/product/unit/repositories"
	unit_services "smlcloudplatform/internal/product/unit/services"
	"smlcloudplatform/internal/productimport/models"
	"smlcloudplatform/internal/productimport/repositories"
	"smlcloudplatform/internal/productimport/services"
	brandproduct_repositories "smlcloudplatform/internal/smlaiproduct/brandproduct/repositories"
	categoryproduct_repositories "smlcloudplatform/internal/smlaiproduct/categoryproduct/repositories"
	classproduct_repositories "smlcloudplatform/internal/smlaiproduct/classproduct/repositories"
	designproduct_repositories "smlcloudplatform/internal/smlaiproduct/designproduct/repositories"
	gradeproduct_repositories "smlcloudplatform/internal/smlaiproduct/gradeproduct/repositories"
	groupproduct_repositories "smlcloudplatform/internal/smlaiproduct/groupproduct/repositories"
	groupsuboneproduct_repositories "smlcloudplatform/internal/smlaiproduct/groupsuboneproduct/repositories"
	groupsubtwoproduct_repositories "smlcloudplatform/internal/smlaiproduct/groupsubtwoproduct/repositories"
	modelproduct_repositories "smlcloudplatform/internal/smlaiproduct/modelproduct/repositories"
	patternproduct_repositories "smlcloudplatform/internal/smlaiproduct/patternproduct/repositories"
	"smlcloudplatform/internal/utils"
	warehouse_repositories "smlcloudplatform/internal/warehouse/repositories"
	"smlcloudplatform/pkg/microservice"
	"strconv"
	"time"
)

type IProductImportHttp interface{}

// SaveTaskRequest - Request structure for SaveTask endpoint
type SaveTaskRequest struct {
	models.ProductImportHeader
	ImportMode string `json:"import_mode"`  // "INSERT_ONLY" | "UPDATE_ONLY" | "BOTH" | "AUTO"
	Preview bool   `json:"preview"`      // true = แสดงผลเฉพาะ, false = บันทึกจริง
	ForceUpdate bool   `json:"force_update"` // true = บังคับ update แม้ไม่มีการเปลี่ยนแปลง
}

// ApplyChangesRequest - Request structure for ApplyChanges endpoint
type ApplyChangesRequest struct {
	ImportMode string `json:"import_mode"`  // "INSERT_ONLY" | "UPDATE_ONLY" | "BOTH" | "AUTO"
	ForceUpdate bool   `json:"force_update"` // Force update even if no changes detected
}

type ProductImportHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IProductImportService
}

func NewProductImportHttp(ms *microservice.Microservice, cfg config.IConfig) ProductImportHttp {
	time.Local = time.UTC
	safeTimeNow := func() time.Time {
		return time.Now().UTC()
	}
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	producer := ms.Producer(cfg.MQConfig())
	pstClickHouse := ms.ClickHousePersister(cfg.ClickHouseConfig())

	repo := product_repositories.NewProductBarcodeRepository(pst, cache)
	repoMq := product_repositories.NewProductBarcodeMessageQueueRepository(producer)
	repoCh := product_repositories.NewProductBarcodeClickhouseRepository(pstClickHouse)
	creditorRepo := creditorRepo.NewCreditorRepository(pst)
	repoMaster := productmaster.NewProductRepository(pst)
	unitmaster := unitmaster.NewUnitRepository(pst)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)

	productcategoryRepo := productcategory_repositories.NewProductCategoryRepository(pst)
	productcategorySvc := productcategory_services.NewProductCategoryHttpService(productcategoryRepo, masterSyncCacheRepo, repo)

	unitRepo := productunit_repo.NewUnitRepository(pst)

	groupProductRepo := groupproduct_repositories.NewGroupProductRepository(pst)
	groupsuboneProductRepo := groupsuboneproduct_repositories.NewGroupsuboneProductRepository(pst)
	groupsubtwoproductRepo := groupsubtwoproduct_repositories.NewGroupsubtwoProductRepository(pst)
	brandProductRepo := brandproduct_repositories.NewBrandProductRepository(pst)
	designProductRepo := designproduct_repositories.NewDesignProductRepository(pst)
	modelProductRepo := modelproduct_repositories.NewModelProductRepository(pst)
	patternProductRepo := patternproduct_repositories.NewPatternProductRepository(pst)
	gradeProductRepo := gradeproduct_repositories.NewGradeProductRepository(pst)
	categoryProductRepo := categoryproduct_repositories.NewCategoryProductRepository(pst)
	classProductRepo := classproduct_repositories.NewClassProductRepository(pst)

	// Branch and BusinessType repositories
	branchRepo := branch_repositories.NewBranchRepository(pst)
	businessTypeRepo := businesstype_repositories.NewBusinessTypeRepository(pst)

	chRepo := repositories.NewProductImportClickHouseRepository(pstClickHouse)
	taskStatusRepo := repositories.NewTaskStatusClickHouseRepository(pstClickHouse)

	// Price History Service
	priceHistoryRepo := product_repositories.NewProductPriceHistoryRepository(pst)
	priceHistorySvc := product_serrvices.NewProductPriceHistoryService(priceHistoryRepo, utils.NewGUID, safeTimeNow)

	// Warehouse Repository
	warehouseRepo := warehouse_repositories.NewWarehouseRepository(pst)

	// Unit Service
	unitMqRepo := unit_repositories.NewUnitMessageQueueRepository(producer)
	unitSvc := unit_services.NewUnitHttpService(unitRepo, repo, unitMqRepo, masterSyncCacheRepo)

	stockBalanceSvc := product_serrvices.NewProductBarcodeHttpService(repo, repoMaster, unitmaster, unitSvc, *creditorRepo, repoMq, repoCh, productcategorySvc, masterSyncCacheRepo, priceHistorySvc, warehouseRepo)

	svc := services.NewProductImportService(chRepo, taskStatusRepo, repo, stockBalanceSvc, unitRepo, groupProductRepo, groupsuboneProductRepo, groupsubtwoproductRepo, brandProductRepo, designProductRepo, modelProductRepo, patternProductRepo, gradeProductRepo, categoryProductRepo, classProductRepo, branchRepo, businessTypeRepo, utils.RandStringBytesMaskImprSrcUnsafe, utils.NewGUID, safeTimeNow)

	return ProductImportHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h ProductImportHttp) RegisterHttp() {
	h.ms.POST("/productimport/upload", h.UploadExcel)
	h.ms.POST("/productimport", h.Create)
	h.ms.GET("/productimport/:task-id", h.List)
	h.ms.DELETE("/productimport/:task-id", h.DeleteByTask)
	h.ms.POST("/productimport/:task-id", h.SaveTask)
	h.ms.PUT("/productimport/item/:guid", h.Update)
	h.ms.DELETE("/productimport/item/:guid", h.Delete)

	h.ms.POST("/productimport/:task-id/verify", h.Verify)
	h.ms.GET("/productimport/task/:task-id/status", h.GetTaskStatus)

	// เพิ่ม endpoints ใหม่สำหรับ compare และ preview
	h.ms.GET("/productimport/:task-id/compare", h.GetCompareResult)
	h.ms.GET("/productimport/:task-id/preview", h.PreviewChanges)
	h.ms.POST("/productimport/:task-id/apply", h.ApplyChanges)
	h.ms.GET("/productimport/:task-id/summary", h.GetImportSummary)
}

// Create ProductImport godoc
// @Description Create ProductImport
// @Tags		ProductImport
// @Param		file  formData      file  true  "excel file"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /productimport/upload [post]
func (h ProductImportHttp) UploadExcel(ctx microservice.IContext) error {
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

// Create ProductImport godoc
// @Description Create ProductImport
// @Tags		ProductImport
// @Param		ProductImport  body      models.ProductImport  true  "ProductImport"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /productimport [post]
func (h ProductImportHttp) Create(ctx microservice.IContext) error {
	shopID := ctx.UserInfo().ShopID
	authUsername := ctx.UserInfo().Username

	input := ctx.ReadInput()

	docReq := models.ProductImport{}
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

// List ProductImport godoc
// @Description List ProductImport
// @Tags		ProductImport
// @Param		task-id		path		string		true		"task id"
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /productimport/{task-id} [get]
func (h ProductImportHttp) List(ctx microservice.IContext) error {
	shopID := ctx.UserInfo().ShopID

	taskID := ctx.Param("task-id")

	if taskID == "" {
		ctx.Response(http.StatusCreated, common.ApiResponse{
			Success: true,
			Data:    []string{},
		})
		return nil
	}

	// เช็ค task status ก่อน
	status, err := h.svc.GetTaskStatus(shopID, taskID)
	if err == nil {
		// ถ้ามี status แปลว่ายังทำงานอยู่หรือ error
		ctx.Response(http.StatusOK, common.ApiResponse{
			Success: true,
			Data: map[string]interface{}{
				"task_status": status,
				"items":       []string{},
			},
		})
		return nil
	}

	// ถ้าไม่มี status แปลว่าเสร็จแล้ว ให้ list ตามปกติ
	pageable := utils.GetPageable(ctx.QueryParam)

	results, page, err := h.svc.List(shopID, taskID, pageable)

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

// Get ProductImport Part godoc
// @Description Get ProductImport Part
// @Tags		ProductImport
// @Param		guid		path		string		true		"guid"
// @Accept 		json
// @Success		201	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /productimport/item/{guid} [put]
func (h ProductImportHttp) Update(ctx microservice.IContext) error {
	shopID := ctx.UserInfo().ShopID

	input := ctx.ReadInput()

	guid := ctx.Param("guid")

	docReq := models.ProductImportRaw{}
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

// Delete ProductImport By GUID godoc
// @Description Delete ProductImport By GUID
// @Tags		ProductImport
// @Param		guid		path		string		true		"guid"
// @Accept 		json
// @Success		201	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /productimport/item/{guid} [delete]
func (h ProductImportHttp) Delete(ctx microservice.IContext) error {
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

// Delete ProductImport By Task ID godoc
// @Description Delete ProductImport By Task ID
// @Tags		ProductImport
// @Param		task-id		path		string		true		"task id"
// @Accept 		json
// @Success		201	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /productimport/{task-id} [delete]
func (h ProductImportHttp) DeleteByTask(ctx microservice.IContext) error {
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

// Save ProductImport Task godoc
// @Description Save ProductImport Task with Compare and Insert/Update modes
// @Tags		ProductImport
// @Param		task-id		path		string		true		"task id"
// @Accept 		json
// @Success		201	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /productimport/{task-id} [post]
func (h ProductImportHttp) SaveTask(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	taskID := ctx.Param("task-id")

	payload := ctx.ReadInput()

	var docReq models.ProductImportHeader
	err := json.Unmarshal([]byte(payload), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.SaveTask(shopID, authUsername, taskID, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
	})
	return nil
}

// Verify ProductImport By Task ID godoc
// @Description Verify ProductImport By Task ID
// @Tags		ProductImport
// @Param		task-id		path		string		true		"task id"
// @Accept 		json
// @Success		201	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /productimport/{task-id}/verify [post]
func (h ProductImportHttp) Verify(ctx microservice.IContext) error {
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

// Get Task Status godoc
// @Description Get Task Status
// @Tags		ProductImport
// @Param		task-id		path		string		true		"task id"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /productimport/task/{task-id}/status [get]
func (h ProductImportHttp) GetTaskStatus(ctx microservice.IContext) error {
	shopID := ctx.UserInfo().ShopID
	taskID := ctx.Param("task-id")

	status, err := h.svc.GetTaskStatus(shopID, taskID)
	if err != nil {
		ctx.ResponseError(http.StatusNotFound, "Task not found or completed")
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    status,
	})
	return nil
}

// GetCompareResult godoc
// @Description Get comparison result between import data and existing products
// @Tags		ProductImport
// @Param		task-id		path		string		true		"task id"
// @Param		page		query		int			false		"Page number (default: 1)"
// @Param		limit		query		int			false		"Records per page (default: 100, max: 500)"
// @Param		fast		query		bool		false		"Fast mode - return summary only (default: false)"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /productimport/{task-id}/compare [get]
func (h ProductImportHttp) GetCompareResult(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	taskID := ctx.Param("task-id")

	// 🔧 เพิ่ม pagination และ fast mode
	pageable := utils.GetPageable(ctx.QueryParam)
	fastMode := ctx.Request().URL.Query().Get("fast") == "true"

	// จำกัด limit สูงสุดเพื่อป้องกัน memory overflow
	if pageable.Limit > 500 {
		pageable.Limit = 500
	}
	if pageable.Limit == 0 {
		pageable.Limit = 100
	}

	if fastMode {
		// 🚀 Fast mode - return summary only (very fast)
		summary, err := h.svc.GetCompareResultSummary(shopID, taskID)
		if err != nil {
			ctx.ResponseError(http.StatusBadRequest, err.Error())
			return err
		}

		// Return only summary to make it faster
		ctx.Response(http.StatusOK, common.ApiResponse{
			Success: true,
			Data: map[string]interface{}{
				"mode":    "summary",
				"summary": summary,
			},
		})
	} else {
		// 🔧 True pagination mode - process only requested page
		result, pagination, err := h.svc.GetCompareResultPaginated(shopID, taskID, pageable)
		if err != nil {
			ctx.ResponseError(http.StatusBadRequest, err.Error())
			return err
		}

		ctx.Response(http.StatusOK, common.ApiResponse{
			Success:    true,
			Pagination: pagination,
			Data:       result,
		})
	}

	return nil
}

// PreviewChanges godoc
// @Description Preview changes that will be made during import
// @Tags		ProductImport
// @Param		task-id		path		string		true		"task id"
// @Param		import_mode	query		string		false		"Import mode (INSERT_ONLY, UPDATE_ONLY, BOTH, AUTO)"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /productimport/{task-id}/preview [get]
func (h ProductImportHttp) PreviewChanges(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	taskID := ctx.Param("task-id")
	importMode := ctx.Request().URL.Query().Get("import_mode")
	if importMode == "" {
		importMode = "AUTO"
	}

	result, err := h.svc.PreviewChanges(shopID, taskID, importMode)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    result,
	})

	return nil
}

// ApplyChanges godoc
// @Description Apply changes and perform the actual import
// @Tags		ProductImport
// @Param		task-id		path		string		true		"task id"
// @Param		async		query		bool		false		"Async processing (default: false)"
// @Param		batch_size	query		int			false		"Batch size for processing (default: 100)"
// @Param		ApplyChangesRequest  body      ApplyChangesRequest  true  "ApplyChangesRequest"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /productimport/{task-id}/apply [post]
func (h ProductImportHttp) ApplyChanges(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	taskID := ctx.Param("task-id")
	payload := ctx.ReadInput()

	// 🔧 เพิ่ม async และ batch size parameters
	asyncMode := ctx.Request().URL.Query().Get("async") == "true"
	batchSizeStr := ctx.Request().URL.Query().Get("batch_size")
	batchSize := 100 // default
	if batchSizeStr != "" {
		if size, err := strconv.Atoi(batchSizeStr); err == nil && size > 0 && size <= 1000 {
			batchSize = size
		}
	}

	var req ApplyChangesRequest
	err := json.Unmarshal([]byte(payload), &req)
	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	// Default import mode
	if req.ImportMode == "" {
		req.ImportMode = "AUTO"
	}

	if asyncMode {
		// 🚀 สำหรับ async mode ให้ใช้ existing ApplyChanges แต่ใน goroutine
		// สร้าง response ก่อนเพื่อไม่ให้ user รอ
		responseTaskID := taskID + "_apply_" + strconv.FormatInt(time.Now().Unix(), 10)

		ctx.Response(http.StatusAccepted, common.ApiResponse{
			Success: true,
			ID:      responseTaskID,
			Data: map[string]interface{}{
				"mode":            "async",
				"process_task_id": responseTaskID,
				"import_mode":     req.ImportMode,
				"batch_size":      batchSize,
				"status_url":      "/productimport/task/" + taskID + "/status",
				"original_task":   taskID,
			},
			Message: "Import started in background. Check original task status.",
		})

		// 🔧 สร้าง initial task status สำหรับ progress tracking
		go func() {
			// สร้าง initial status
			err := h.svc.CreateTaskStatus(shopID, taskID, "APPLYING", 0, "Starting import process...")
			if err != nil {
				// Log error but continue
			}

			// เรียก ApplyChanges ที่มี progress tracking
			err = h.svc.ApplyChangesWithProgress(shopID, authUsername, taskID, req.ImportMode, req.ForceUpdate, batchSize)

			// อัพเดต final status
			if err != nil {
				// ❌ System error (database, network, etc.)
				_ = h.svc.UpdateTaskStatus(shopID, taskID, "ERROR", 0, err.Error())
			}
			// ✅ ไม่ต้อง update status เป็น "COMPLETED" เพราะ ApplyChangesWithProgress
			// จะตั้ง status เป็น "COMPLETED" หรือ "COMPLETED_WITH_ERRORS" เองแล้ว
		}()
	} else {
		// 🔧 Synchronous processing - สร้าง task status สำหรับ sync mode ด้วย
		err = h.svc.CreateTaskStatus(shopID, taskID, "APPLYING", 0, "Starting synchronous import...")
		if err != nil {
			// Log warning but continue
		}

		err = h.svc.ApplyChangesWithProgress(shopID, authUsername, taskID, req.ImportMode, req.ForceUpdate, batchSize)
		if err != nil {
			_ = h.svc.UpdateTaskStatus(shopID, taskID, "ERROR", 0, err.Error())
			ctx.ResponseError(http.StatusBadRequest, err.Error())
			return err
		}

		// ✅ ไม่ต้อง update status เป็น "COMPLETED" เพราะ ApplyChangesWithProgress
		// จะตั้ง status เป็น "COMPLETED" หรือ "COMPLETED_WITH_ERRORS" เองแล้ว

		ctx.Response(http.StatusOK, common.ApiResponse{
			Success: true,
			Data: map[string]interface{}{
				"mode":        "sync",
				"import_mode": req.ImportMode,
				"batch_size":  batchSize,
				"task_id":     taskID,
			},
			Message: "Changes applied successfully with mode: " + req.ImportMode,
		})
	}

	return nil
}

// GetImportSummary godoc
// @Description Get summary of import operation
// @Tags		ProductImport
// @Param		task-id		path		string		true		"task id"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /productimport/{task-id}/summary [get]
func (h ProductImportHttp) GetImportSummary(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	taskID := ctx.Param("task-id")

	result, err := h.svc.GetImportSummary(shopID, taskID)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    result,
	})

	return nil
}
