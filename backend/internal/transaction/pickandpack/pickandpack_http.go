package pickandpack

import (
	"encoding/json"
	"fmt"
	"net/http"
	"smlcloudplatform/internal/config"
	couponRepositories "smlcloudplatform/internal/coupon/repositories"
	couponServices "smlcloudplatform/internal/coupon/services"
	repoCust "smlcloudplatform/internal/debtaccount/debtor/repositories"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productBarcodeRepositories "smlcloudplatform/internal/product/productbarcode/repositories"
	transmodels "smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/pickandpack/models"
	"smlcloudplatform/internal/transaction/pickandpack/repositories"
	"smlcloudplatform/internal/transaction/pickandpack/services"
	trancache "smlcloudplatform/internal/transaction/repositories"
	saleinvoiceRepositories "smlcloudplatform/internal/transaction/saleinvoice/repositories"
	saleinvoiceServices "smlcloudplatform/internal/transaction/saleinvoice/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type IPickandpackHttp interface{}

type PickandpackHttp struct {
	ms             *microservice.Microservice
	cfg            config.IConfig
	svc            services.IPickandpackHttpService
	saleInvoiceSvc saleinvoiceServices.ISaleInvoiceService
}

func NewPickandpackHttp(ms *microservice.Microservice, cfg config.IConfig) PickandpackHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	producer := ms.Producer(cfg.MQConfig())

	repo := repositories.NewPickandpackRepository(pst)
	repoMq := repositories.NewPickandpackMessageQueueRepository(producer)

	transRepo := trancache.NewCacheRepository(cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewPickandpackHttpService(repo, transRepo, repoMq, masterSyncCacheRepo)

	// Create SaleInvoice service dependencies
	saleInvoiceRepo := saleinvoiceRepositories.NewSaleInvoiceRepository(pst)
	custRepo := repoCust.NewDebtorRepository(pst)
	productBarcodeRepo := productBarcodeRepositories.NewProductBarcodeRepository(pst, cache)
	saleInvoiceRepoMq := saleinvoiceRepositories.NewSaleInvoiceMessageQueueRepository(producer)

	// Initialize coupon service for SaleInvoice
	couponRepo := couponRepositories.NewCouponRepository(pst)
	couponReservationRepo := couponRepositories.NewCouponReservationRepository(pst)
	couponUsageHistoryRepo := couponRepositories.NewCouponUsageHistoryRepository(pst)
	couponService := couponServices.NewCouponHttpService(couponRepo, couponReservationRepo, couponUsageHistoryRepo, productBarcodeRepo, masterSyncCacheRepo)

	saleInvoiceSvc := saleinvoiceServices.NewSaleInvoiceService(
		saleInvoiceRepo,
		custRepo,
		transRepo,
		productBarcodeRepo,
		nil, // pointTransactionRepo - not needed for InfoSaleInvoice
		saleInvoiceRepoMq,
		masterSyncCacheRepo,
		couponService,
		saleinvoiceServices.SaleInvocieParser{},
		saleinvoiceServices.SaleInvocieExport{},
	)

	return PickandpackHttp{
		ms:             ms,
		cfg:            cfg,
		svc:            svc,
		saleInvoiceSvc: saleInvoiceSvc,
	}
}

func (h PickandpackHttp) RegisterHttp() {

	h.ms.POST("/transaction/pickandpack/bulk", h.SaveBulk)

	h.ms.GET("/transaction/pickandpack", h.SearchPickandpackPage)
	h.ms.GET("/transaction/pickandpack/list", h.SearchPickandpackStep)
	h.ms.GET("/transaction/pickandpack/available-saleinvoice", h.SearchAvailableSaleInvoice)
	h.ms.GET("/transaction/pickandpack/by-warehouse", h.SearchPickandpackByWarehouse)
	h.ms.GET("/transaction/pickandpack/saleinvoice-status", h.GetSaleInvoicePackingStatus)
	h.ms.GET("/transaction/pickandpack/history", h.GetPickandpackHistory)
	h.ms.GET("/transaction/pickandpack/dashboard/warehouse", h.GetWarehouseDashboard)
	h.ms.GET("/transaction/pickandpack/dashboard", h.GetPickandpackDashboard)
	h.ms.POST("/transaction/pickandpack", h.CreatePickandpack)
	h.ms.POST("/transaction/pickandpack/approve/:id", h.ApprovePickandpack)
	h.ms.PUT("/transaction/pickandpack/closejob/:id", h.CloseSaleInvoiceJob)
	h.ms.DELETE("/transaction/pickandpack/cancel-by-saleinvoice/:id", h.CancelPickandpackBySaleInvoice)
	h.ms.PUT("/transaction/pickandpack/updateprint/:docno", h.UpdatePrint)
	h.ms.PUT("/transaction/pickandpack/confirmpickandpack/:docno", h.ConfirmPickandpack)
	h.ms.PUT("/transaction/pickandpack/cancel/:docno", h.CancelPickandpack)
	h.ms.PUT("/transaction/pickandpack/status/:id/:status", h.UpdatePickandpackStatus)
	h.ms.PUT("/transaction/pickandpack/details/:id", h.UpdatePickandpackDetails)
	h.ms.GET("/transaction/pickandpack/:id", h.InfoPickandpack)
	h.ms.GET("/transaction/pickandpack/code/:code", h.InfoPickandpackByCode)
	h.ms.PUT("/transaction/pickandpack/:id", h.UpdatePickandpack)
	h.ms.DELETE("/transaction/pickandpack/:id", h.DeletePickandpack)
	h.ms.DELETE("/transaction/pickandpack", h.DeletePickandpackByGUIDs)
}

// Create Pickandpack godoc
// @Description Create Pickandpack
// @Tags		Pickandpack
// @Param		Pickandpack  body      models.Pickandpack  true  "Pickandpack"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/pickandpack [post]
func (h PickandpackHttp) CreatePickandpack(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.Pickandpack{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, docNo, err := h.svc.CreatePickandpack(shopID, authUsername, *docReq)

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

// Search Available SaleInvoice godoc
// @Description Search SaleInvoice that are available for Pickandpack (packingstatus = 0 or null)
// @Tags		Pickandpack
// @Param		q		query	string		false  "Search Keyword (custcode, docno)"
// @Param		fromdate	query	string		false  "from date"
// @Param		todate	query	string		false  "to date"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/pickandpack/available-saleinvoice [get]
func (h PickandpackHttp) SearchAvailableSaleInvoice(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)
	searchKeyword := ctx.QueryParam("q")

	// Generate filters for SaleInvoice search
	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "-",
			Field: "docdatetime",
			Type:  requestfilter.FieldTypeRangeDate,
		},
	})

	// Add IsCancel = false to the filter query parameters manually
	filterMap := make(map[string]interface{})
	for k, v := range filters {
		filterMap[k] = v
	}
	filterMap["iscancel"] = false

	// บังคับเงื่อนไข packingstatus (available only)
	filterMap["$or"] = []map[string]interface{}{
		{"packingstatus": 0},
		{"packingstatus": nil},
		{"packingstatus": map[string]interface{}{"$exists": false}},
	}

	// เพิ่ม search keyword filter ถ้ามี param q
	if searchKeyword != "" {
		keyword := strings.ToLower(searchKeyword)
		filterMap["$and"] = []map[string]interface{}{
			{
				"$or": []map[string]interface{}{
					{"custcode": map[string]interface{}{"$regex": keyword, "$options": "i"}},
					{"docno": map[string]interface{}{"$regex": keyword, "$options": "i"}},
				},
			},
		}
	}

	// Search for SaleInvoice documents
	saleInvoiceList, pagination, err := h.saleInvoiceSvc.SearchSaleInvoice(shopID, filterMap, pageable)
	if err != nil {
		h.ms.Logger.Errorf("Error searching SaleInvoice: %v", err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	// ไม่ต้องกรองข้อมูลอีกแล้ว เพราะ database ทำให้แล้ว
	ctx.Response(http.StatusOK, common.ApiResponse{
		Success:    true,
		Data:       saleInvoiceList,
		Pagination: pagination,
	})
	return nil
}

// Search Pickandpack By Warehouse godoc
// @Description Search Pickandpack by WhCode and LocationCode with multiple values
// @Tags		Pickandpack
// @Param		whcodes		query	string		false  "Warehouse codes (comma separated, e.g., WH001,WH002,WH003)"
// @Param		locationcodes	query	string		false  "Location codes (comma separated, e.g., L001,L002,L003)"
// @Param		q		query	string		false  "Search keyword for docno or custcode"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/pickandpack/by-warehouse [get]
func (h PickandpackHttp) SearchPickandpackByWarehouse(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	// Get comma-separated values for whcodes and locationcodes
	whCodesParam := ctx.QueryParam("whcodes")
	locationCodesParam := ctx.QueryParam("locationcodes")
	searchKeyword := strings.TrimSpace(ctx.QueryParam("q"))

	// Create base filters
	filters := make(map[string]interface{})

	// Parse and add WhCode filter (WHERE IN)
	if whCodesParam != "" {
		whCodes := strings.Split(whCodesParam, ",")
		// Trim spaces from each code
		for i, code := range whCodes {
			whCodes[i] = strings.TrimSpace(code)
		}
		// Filter empty codes
		var validWhCodes []string
		for _, code := range whCodes {
			if code != "" {
				validWhCodes = append(validWhCodes, code)
			}
		}
		if len(validWhCodes) > 0 {
			if len(validWhCodes) == 1 {
				filters["whcode"] = validWhCodes[0]
			} else {
				filters["whcode"] = map[string]interface{}{
					"$in": validWhCodes,
				}
			}
		}
	}

	// Parse and add LocationCode filter (WHERE IN)
	if locationCodesParam != "" {
		locationCodes := strings.Split(locationCodesParam, ",")
		// Trim spaces from each code
		for i, code := range locationCodes {
			locationCodes[i] = strings.TrimSpace(code)
		}
		// Filter empty codes
		var validLocationCodes []string
		for _, code := range locationCodes {
			if code != "" {
				validLocationCodes = append(validLocationCodes, code)
			}
		}
		if len(validLocationCodes) > 0 {
			if len(validLocationCodes) == 1 {
				filters["locationcode"] = validLocationCodes[0]
			} else {
				filters["locationcode"] = map[string]interface{}{
					"$in": validLocationCodes,
				}
			}
		}
	}

	// Add search filter for docno and custcode
	if searchKeyword != "" {
		filters["$or"] = []map[string]interface{}{
			{
				"docno": map[string]interface{}{
					"$regex":   searchKeyword,
					"$options": "i", // case insensitive
				},
			},
			{
				"custcode": map[string]interface{}{
					"$regex":   searchKeyword,
					"$options": "i", // case insensitive
				},
			},
		}
	}

	// Search Pickandpack documents
	docList, pagination, err := h.svc.SearchPickandpack(shopID, filters, pageable)
	if err != nil {
		h.ms.Logger.Errorf("Error searching Pickandpack by warehouse: %v", err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	h.ms.Logger.Infof("Found %d Pickandpack documents for warehouse filter - WhCodes: %s, LocationCodes: %s, SearchKeyword: %s",
		len(docList), whCodesParam, locationCodesParam, searchKeyword)

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success:    true,
		Data:       docList,
		Pagination: pagination,
	})
	return nil
}

// Get SaleInvoice Packing Status godoc
// @Description Get SaleInvoice packing status with progress information for display
// @Tags		Pickandpack
// @Param		q		query	string		false  "Search Keyword (custcode, docno)"
// @Param		fromdate	query	string		false  "from date"
// @Param		todate	query	string		false  "to date"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/pickandpack/saleinvoice-status [get]
func (h PickandpackHttp) GetSaleInvoicePackingStatus(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)
	searchKeyword := ctx.QueryParam("q")

	// Handle date range filtering manually for better control
	filterMap := make(map[string]interface{})

	fromDateStr := strings.TrimSpace(ctx.QueryParam("fromdate"))
	toDateStr := strings.TrimSpace(ctx.QueryParam("todate"))

	if len(fromDateStr) > 0 && len(toDateStr) > 0 {
		fromDate, err1 := time.Parse("2006-01-02", fromDateStr)
		toDate, err2 := time.Parse("2006-01-02", toDateStr)

		if err1 == nil && err2 == nil {
			// Set time to beginning of fromDate and end of toDate to include full days
			fromDateTime := time.Date(fromDate.Year(), fromDate.Month(), fromDate.Day(), 0, 0, 0, 0, time.UTC)
			toDateTime := time.Date(toDate.Year(), toDate.Month(), toDate.Day(), 23, 59, 59, 999999999, time.UTC)

			filterMap["docdatetime"] = bson.M{
				"$gte": fromDateTime,
				"$lte": toDateTime,
			}
		}
	}

	// Add other filters
	filterMap["iscancel"] = false
	filterMap["packingstatus"] = 1
	filterMap["isclose"] = false

	// Search for SaleInvoice documents
	saleInvoiceList, pagination, err := h.saleInvoiceSvc.SearchSaleInvoice(shopID, filterMap, pageable)
	if err != nil {
		h.ms.Logger.Errorf("Error searching SaleInvoice: %v", err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	// Get all Pickandpack documents to check their status
	pickandpackMap, err := h.getAllPickandpackByRefSaleInvoice(shopID)
	if err != nil {
		h.ms.Logger.Errorf("Error getting Pickandpack documents: %v", err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	// Process each SaleInvoice and calculate packing status
	var statusList []interface{}
	for _, invoice := range saleInvoiceList {
		// Apply search keyword filter if provided
		if searchKeyword != "" {
			keyword := strings.ToLower(searchKeyword)
			if !strings.Contains(strings.ToLower(invoice.CustCode), keyword) &&
				!strings.Contains(strings.ToLower(invoice.DocNo), keyword) {
				continue
			}
		}

		// Get Pickandpack documents for this SaleInvoice
		pickandpacks, exists := pickandpackMap[invoice.DocNo]
		if !exists || len(pickandpacks) == 0 {
			// No Pickandpack documents found
			statusList = append(statusList, map[string]interface{}{
				"saleInvoiceDocNo": invoice.DocNo,
				"saleInvoice":      invoice,
				"status":           "ยังไม่จัด",
				"statusCode":       "NOT_STARTED",
				"progress":         0,
				"totalPickandpack": 0,
				"pickandpackList":  []interface{}{},
				"completedCount":   0,
				"processingCount":  0,
				"pendingCount":     0,
				"cancelledCount":   0,
				"sendtype":         invoice.Sendtype,
				"email":            invoice.Email,
				"address":          invoice.Address,
				"phone":            invoice.Phone,
			})
			continue
		}

		// Calculate status based on PackStatus of all Pickandpack documents
		totalCount := len(pickandpacks)
		completedCount := 0
		processingCount := 0
		pendingCount := 0
		cancelledCount := 0

		pickandpackDetails := []interface{}{}
		for _, pp := range pickandpacks {
			pickandpackDetails = append(pickandpackDetails, map[string]interface{}{
				"docNo":         pp.DocNo,
				"id":            pp.GuidFixed,
				"packStatus":    pp.PackStatus,
				"statusText":    getStatusText(pp.PackStatus),
				"whCode":        pp.WhCode,
				"whNames":       pp.WhNames,
				"locationCode":  pp.LocationCode,
				"locationNames": pp.LocationNames,
				"totalQty":      pp.TotalQty,
				"totalAmount":   pp.TotalAmount,
			})

			switch pp.PackStatus {
			case 0:
				pendingCount++
			case 1:
				processingCount++
			case 2:
				completedCount++
			case 3:
				cancelledCount++
			}
		}

		// Calculate progress percentage (based only on completed items, excluding cancelled)
		activeCount := totalCount - cancelledCount
		progress := 0
		statusText := ""
		statusCode := ""

		if activeCount == 0 {
			// All cancelled
			progress = 0
			statusText = "ยกเลิกทั้งหมด"
			statusCode = "ALL_CANCELLED"
		} else if completedCount == activeCount {
			// All completed
			progress = 100
			statusText = "จัดเสร็จแล้ว"
			statusCode = "COMPLETED"
		} else if completedCount > 0 {
			// Partially completed - progress based only on completed items
			progress = int((float64(completedCount) / float64(activeCount)) * 100)
			statusText = "จัดเสร็จบางส่วน"
			statusCode = "PARTIALLY_COMPLETED"
		} else if processingCount > 0 {
			// Some in progress but none completed yet
			progress = 0
			statusText = "กำลังจัด"
			statusCode = "IN_PROGRESS"
		} else {
			// All pending
			progress = 0
			statusText = "รอดำเนินการ"
			statusCode = "PENDING"
		}

		statusList = append(statusList, map[string]interface{}{
			"saleInvoiceDocNo": invoice.DocNo,
			"saleInvoice":      invoice,
			"status":           statusText,
			"statusCode":       statusCode,
			"progress":         progress,
			"totalPickandpack": totalCount,
			"pickandpackList":  pickandpackDetails,
			"completedCount":   completedCount,
			"processingCount":  processingCount,
			"pendingCount":     pendingCount,
			"cancelledCount":   cancelledCount,
			"sendtype":         invoice.Sendtype,
			"email":            invoice.Email,
			"address":          invoice.Address,
			"phone":            invoice.Phone,
		})
	}

	// Update pagination for filtered results
	pagination.Total = int64(len(statusList))

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success:    true,
		Data:       statusList,
		Pagination: pagination,
	})
	return nil
}

// Helper function to get all Pickandpack documents grouped by RefSaleInvoice
func (h PickandpackHttp) getAllPickandpackByRefSaleInvoice(shopID string) (map[string][]models.PickandpackInfo, error) {
	// Create empty filters to get all Pickandpack documents
	emptyFilters := make(map[string]interface{})

	// Create pageable with large limit to get all documents
	getPageParam := func(key string) string {
		if key == "page" {
			return "1"
		}
		if key == "limit" {
			return "10000"
		}
		return ""
	}
	pageable := utils.GetPageable(getPageParam)

	// Get all Pickandpack documents
	pickandpackList, _, err := h.svc.SearchPickandpack(shopID, emptyFilters, pageable)
	if err != nil {
		return nil, err
	}

	// Group by RefSaleInvoice
	pickandpackMap := make(map[string][]models.PickandpackInfo)
	for _, pp := range pickandpackList {
		if pp.RefSaleInvoice != "" {
			pickandpackMap[pp.RefSaleInvoice] = append(pickandpackMap[pp.RefSaleInvoice], pp)
		}
	}

	return pickandpackMap, nil
}

// Helper function to update SaleInvoice packing status and progress
func (h PickandpackHttp) updateSaleInvoicePackingStatus(shopID, saleInvoiceDocNo string, saleInvoiceguid string, packingStatus, packingProgress int, updatedBy string) error {
	// If status and progress are -1, auto calculate from Pickandpack documents
	if packingStatus == -1 || packingProgress == -1 {
		// Get all Pickandpack documents for this SaleInvoice
		filters := map[string]interface{}{
			"refsaleinvoice": saleInvoiceDocNo,
		}

		pickandpackList, _, err := h.svc.SearchPickandpack(shopID, filters, utils.GetPageable(func(key string) string {
			if key == "page" {
				return "1"
			}
			if key == "limit" {
				return "100"
			}
			return ""
		}))
		if err != nil {
			return err
		}

		if len(pickandpackList) == 0 {
			// No Pickandpack documents, set status to 0 (not started)
			packingStatus = 0
			packingProgress = 0
		} else {
			// Calculate progress based on PackStatus
			totalCount := len(pickandpackList)
			completedCount := 0
			cancelledCount := 0

			for _, pp := range pickandpackList {
				switch pp.PackStatus {
				case 2: // COMPLETED
					completedCount++
				case 3: // CANCELLED
					cancelledCount++
				}
			}

			activeCount := totalCount - cancelledCount
			if activeCount == 0 {
				// All cancelled
				packingStatus = 3 // CANCELLED
				packingProgress = 0
			} else if completedCount == activeCount {
				// All completed
				packingStatus = 2 // COMPLETED
				packingProgress = 100
			} else if completedCount > 0 {
				// Partially completed
				packingStatus = 1 // IN_PROGRESS
				packingProgress = int((float64(completedCount) / float64(activeCount)) * 100)
			} else {
				// None completed yet
				packingStatus = 1 // IN_PROGRESS
				packingProgress = 0
			}
		}
	}

	// For now, just log the values since we don't have UpdateSaleInvoiceStatus method
	h.ms.Logger.Infof("Would update SaleInvoice %s: PackingStatus=%d, PackingProgress=%d",
		saleInvoiceDocNo, packingStatus, packingProgress)

	// TODO: When SaleInvoice model has PackingStatus and PackingProgress fields, uncomment this:

	// Get the SaleInvoice document first
	saleInvoiceInfo, err := h.saleInvoiceSvc.InfoSaleInvoice(shopID, saleInvoiceguid)
	if err != nil {
		return fmt.Errorf("failed to get SaleInvoice %s: %v", saleInvoiceDocNo, err)
	}

	// Update the fields
	saleInvoiceInfo.PackingStatus = int8(packingStatus)
	saleInvoiceInfo.PackingProgress = packingProgress
	saleInvoiceInfo.IsClose = false
	saleInvoiceInfo.PackingBy = updatedBy
	PackingAt := time.Now().UTC()
	saleInvoiceInfo.PackingAt = &PackingAt

	// Save back using UpdateSaleInvoice method (pass the embedded SaleInvoice struct)
	err = h.saleInvoiceSvc.UpdateSaleInvoice(shopID, saleInvoiceInfo.GuidFixed, "system", saleInvoiceInfo.SaleInvoice)
	if err != nil {
		return fmt.Errorf("failed to update SaleInvoice %s: %v", saleInvoiceInfo.GuidFixed, err)
	}

	return nil
}

// Helper function to get existing RefSaleInvoice values
func (h PickandpackHttp) getExistingRefSaleInvoices(shopID string) (map[string]bool, error) {
	// Create empty filters to get all Pickandpack documents
	emptyFilters := make(map[string]interface{})

	// Create pageable with large limit to get all documents
	getPageParam := func(key string) string {
		if key == "page" {
			return "1"
		}
		if key == "limit" {
			return "10000"
		}
		return ""
	}
	pageable := utils.GetPageable(getPageParam)

	// Get all Pickandpack documents
	pickandpackList, _, err := h.svc.SearchPickandpack(shopID, emptyFilters, pageable)
	if err != nil {
		return nil, err
	}

	// Create a map of existing RefSaleInvoice values
	existingRefs := make(map[string]bool)
	for _, pp := range pickandpackList {
		if pp.RefSaleInvoice != "" {
			existingRefs[pp.RefSaleInvoice] = true
		}
	}

	return existingRefs, nil
}

// Approve Pickandpack godoc
// @Description Approve Pickandpack
// @Tags		Pickandpack
// @Param		id  path      string  true  "SaleInvoice ID"
// @Param		Pickandpack  body      models.Pickandpack  true  "Pickandpack"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/pickandpack/approve/{id} [post]
func (h PickandpackHttp) ApprovePickandpack(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID

	id := ctx.Param("id")

	doc, err := h.saleInvoiceSvc.InfoSaleInvoice(shopID, id)

	if err != nil {
		h.ms.Logger.Errorf("Error getting document %v: %v", id, err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	input := ctx.ReadInput()

	docRequest := &models.Pickandpack{}
	errx := json.Unmarshal([]byte(input), &docRequest)

	if errx != nil {
		ctx.ResponseError(400, errx.Error())
		return errx
	}

	// Log the retrieved document for debugging
	h.ms.Logger.Debugf("Retrieved SaleInvoice document: %+v", doc)

	// Group details by WhCode and LocationCode
	warehouseGroups := make(map[string][]transmodels.Detail)
	warehouseNames := make(map[string]*[]common.NameX) // Store WhNames for each group
	locationNames := make(map[string]*[]common.NameX)  // Store LocationNames for each group

	if doc.Details != nil {
		for _, detail := range *doc.Details {
			// Create a unique key for warehouse + location combination
			warehouseKey := fmt.Sprintf("%s-%s", detail.WhCode, detail.LocationCode)

			// Add detail to the corresponding warehouse group
			warehouseGroups[warehouseKey] = append(warehouseGroups[warehouseKey], detail)

			// Store names from the first detail in each group
			if _, exists := warehouseNames[warehouseKey]; !exists {
				warehouseNames[warehouseKey] = detail.WhNames
				locationNames[warehouseKey] = detail.LocationNames
			}
		}
	}

	h.ms.Logger.Infof("Found %d warehouse groups for SaleInvoice %s", len(warehouseGroups), doc.DocNo)

	// Create separate Pickandpack documents for each warehouse group
	var createdDocs []map[string]interface{}
	groupIndex := 1

	for warehouseKey, details := range warehouseGroups {
		// Extract WhCode and LocationCode from key
		parts := strings.Split(warehouseKey, "-")
		whCode := parts[0]
		locationCode := ""
		if len(parts) > 1 {
			locationCode = parts[1]
		}

		// Process details to set eventqty = qty
		processedDetails := make([]transmodels.Detail, len(details))
		for i, detail := range details {
			processedDetails[i] = detail
			// Set eventqty equal to qty for each detail
			processedDetails[i].EventQty = detail.Qty
		}

		h.ms.Logger.Debugf("Processed %d details for warehouse %s, set eventqty = qty for each item", len(processedDetails), warehouseKey)

		// Create Pickandpack document for this warehouse group
		docReq := &models.Pickandpack{

			Transaction: transmodels.Transaction{
				TransactionHeader: doc.TransactionHeader,
				Details:           &processedDetails,
			},

			WhCode:         whCode,
			WhNames:        warehouseNames[warehouseKey],
			LocationCode:   locationCode,
			LocationNames:  locationNames[warehouseKey],
			PackStatus:     0,
			RefSaleInvoice: doc.DocNo,
			Sendtype:       docRequest.Sendtype,
			Email:          docRequest.Email,
			Phone:          docRequest.Phone,
			Address:        docRequest.Address,
		}

		docReq.DocDatetime = time.Now().UTC()
		docReq.DocNo = fmt.Sprintf("%s-%d", doc.DocNo, groupIndex)
		docReq.Description = fmt.Sprintf("Pick and Pack for %s - Warehouse: %s, Location: %s", doc.DocNo, whCode, locationCode)
		// Copy other relevant header fields
		docReq.CustCode = doc.CustCode
		docReq.CustNames = doc.CustNames
		docReq.TotalValue = calculateTotalValue(processedDetails)
		docReq.TotalAmount = calculateTotalAmount(processedDetails)
		docReq.TotalQty = calculateTotalQty(processedDetails)

		docReq.Branch = doc.Branch
		docReq.InquiryType = doc.InquiryType
		docReq.VatRate = doc.VatRate
		docReq.VatType = doc.VatType
		docReq.IsClose = false

		// Create the Pickandpack document
		idx, docNo, err := h.svc.CreatePickandpack(shopID, authUsername, *docReq)

		if err != nil {
			h.ms.Logger.Errorf("Error creating Pickandpack document for warehouse %s: %v", warehouseKey, err)
			ctx.ResponseError(http.StatusBadRequest, fmt.Sprintf("Error creating Pickandpack for warehouse %s: %v", warehouseKey, err))
			return err
		}

		h.ms.Logger.Infof("Created Pickandpack document: ID=%s, DocNo=%s for warehouse=%s, location=%s", idx, docNo, whCode, locationCode)

		createdDocs = append(createdDocs, map[string]interface{}{
			"id":          idx,
			"docNo":       docNo,
			"warehouse":   whCode,
			"location":    locationCode,
			"itemCount":   len(processedDetails),
			"totalQty":    calculateTotalQty(processedDetails),
			"totalAmount": calculateTotalAmount(processedDetails),
		})

		groupIndex++
	}

	// อัพเดท SaleInvoice fields จาก docRequest
	doc.Phone = docRequest.Phone
	doc.Sendtype = docRequest.Sendtype
	doc.Email = docRequest.Email
	doc.Address = docRequest.Address

	// Save SaleInvoice with updated fields
	err = h.saleInvoiceSvc.UpdateSaleInvoice(shopID, doc.GuidFixed, authUsername, doc.SaleInvoice)
	if err != nil {
		h.ms.Logger.Errorf("Error updating SaleInvoice fields for %s: %v", doc.DocNo, err)
		ctx.ResponseError(http.StatusBadRequest, fmt.Sprintf("Error updating SaleInvoice fields: %v", err))
		return err
	}

	h.ms.Logger.Infof("Updated SaleInvoice %s with Phone=%s, Sendtype=%s, Email=%s, Address=%s",
		doc.DocNo, docRequest.Phone, docRequest.Sendtype, docRequest.Email, docRequest.Address)

	// หลังจากสร้าง Pickandpack เสร็จ - Update SaleInvoice PackingStatus
	err = h.updateSaleInvoicePackingStatus(shopID, doc.DocNo, doc.GuidFixed, 1, 0, authUsername)
	if err != nil {
		h.ms.Logger.Warnf("Warning: Could not update SaleInvoice packing status for %s: %v", doc.DocNo, err)
		// Don't return error, just log warning as Pickandpack creation was successful
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		Data: map[string]interface{}{
			"message":          fmt.Sprintf("Successfully created %d Pickandpack documents from SaleInvoice %s", len(createdDocs), doc.DocNo),
			"saleInvoiceDocNo": doc.DocNo,
			"totalGroups":      len(createdDocs),
			"createdDocuments": createdDocs,
		},
	})
	return nil
}

// Close SaleInvoice Job godoc
// @Description Close SaleInvoice job by setting IsClose = true
// @Tags		Pickandpack
// @Param		id  path      string  true  "SaleInvoice GUID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/pickandpack/closejob/{id} [put]
func (h PickandpackHttp) CloseSaleInvoiceJob(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	// Get the SaleInvoice document first
	saleInvoiceInfo, err := h.saleInvoiceSvc.InfoSaleInvoice(shopID, id)
	if err != nil {
		h.ms.Logger.Errorf("Error getting SaleInvoice document %v: %v", id, err)
		ctx.ResponseError(http.StatusBadRequest, fmt.Sprintf("SaleInvoice not found: %v", err))
		return err
	}

	// Update IsClose to true and add close job information
	closeJobAt := time.Now().UTC()
	saleInvoiceInfo.IsClose = true
	saleInvoiceInfo.CloseJobBy = userInfo.Username
	saleInvoiceInfo.CloseJobAt = &closeJobAt

	// Save back using UpdateSaleInvoice method (pass the embedded SaleInvoice struct)
	err = h.saleInvoiceSvc.UpdateSaleInvoice(shopID, saleInvoiceInfo.GuidFixed, userInfo.Username, saleInvoiceInfo.SaleInvoice)
	if err != nil {
		h.ms.Logger.Errorf("Error updating SaleInvoice IsClose %v: %v", id, err)
		ctx.ResponseError(http.StatusBadRequest, fmt.Sprintf("Failed to close SaleInvoice job: %v", err))
		return err
	}

	h.ms.Logger.Infof("Successfully closed SaleInvoice job %s (DocNo: %s) by user %s", id, saleInvoiceInfo.DocNo, userInfo.Username)

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      id,
		Data: map[string]interface{}{
			"message":    "SaleInvoice job closed successfully",
			"docNo":      saleInvoiceInfo.DocNo,
			"isClose":    true,
			"closeJobBy": userInfo.Username,
			"closeJobAt": closeJobAt,
		},
	})

	return nil
}

// Cancel Pickandpack By SaleInvoice godoc
// @Description Cancel all Pickandpack documents referenced to a SaleInvoice and reset SaleInvoice status to available
// @Tags		Pickandpack
// @Param		id  path      string  true  "SaleInvoice GUID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/pickandpack/cancel-by-saleinvoice/{id} [delete]
func (h PickandpackHttp) CancelPickandpackBySaleInvoice(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	saleInvoiceId := ctx.Param("id")

	// Get the SaleInvoice document first to validate and get DocNo
	saleInvoiceInfo, err := h.saleInvoiceSvc.InfoSaleInvoice(shopID, saleInvoiceId)
	if err != nil {
		h.ms.Logger.Errorf("Error getting SaleInvoice document %v: %v", saleInvoiceId, err)
		ctx.ResponseError(http.StatusBadRequest, fmt.Sprintf("SaleInvoice not found: %v", err))
		return err
	}

	// Find all Pickandpack documents that reference this SaleInvoice
	filters := map[string]interface{}{
		"refsaleinvoice": saleInvoiceInfo.DocNo,
	}

	pickandpackList, _, err := h.svc.SearchPickandpack(shopID, filters, utils.GetPageable(func(key string) string {
		if key == "page" {
			return "1"
		}
		if key == "limit" {
			return "1000"
		}
		return ""
	}))
	if err != nil {
		h.ms.Logger.Errorf("Error searching Pickandpack documents for SaleInvoice %s: %v", saleInvoiceInfo.DocNo, err)
		ctx.ResponseError(http.StatusBadRequest, fmt.Sprintf("Error finding Pickandpack documents: %v", err))
		return err
	}

	if len(pickandpackList) == 0 {
		h.ms.Logger.Infof("No Pickandpack documents found for SaleInvoice %s", saleInvoiceInfo.DocNo)
		ctx.ResponseError(http.StatusBadRequest, "No Pickandpack documents found for this SaleInvoice")
		return fmt.Errorf("no pickandpack documents found")
	}

	// Collect GUIDs of all Pickandpack documents to delete
	var pickandpackGUIDs []string
	var deletedDocs []map[string]interface{}

	for _, pp := range pickandpackList {
		pickandpackGUIDs = append(pickandpackGUIDs, pp.GuidFixed)
		deletedDocs = append(deletedDocs, map[string]interface{}{
			"id":            pp.GuidFixed,
			"docNo":         pp.DocNo,
			"packStatus":    pp.PackStatus,
			"statusText":    getStatusText(pp.PackStatus),
			"whCode":        pp.WhCode,
			"whNames":       pp.WhNames,
			"locationCode":  pp.LocationCode,
			"locationNames": pp.LocationNames,
			"totalQty":      pp.TotalQty,
			"totalAmount":   pp.TotalAmount,
		})
	}

	h.ms.Logger.Infof("Found %d Pickandpack documents to delete for SaleInvoice %s", len(pickandpackGUIDs), saleInvoiceInfo.DocNo)

	// Delete all Pickandpack documents
	err = h.svc.DeletePickandpackByGUIDs(shopID, authUsername, pickandpackGUIDs)
	if err != nil {
		h.ms.Logger.Errorf("Error deleting Pickandpack documents: %v", err)
		ctx.ResponseError(http.StatusBadRequest, fmt.Sprintf("Error deleting Pickandpack documents: %v", err))
		return err
	}

	h.ms.Logger.Infof("Successfully deleted %d Pickandpack documents for SaleInvoice %s", len(pickandpackGUIDs), saleInvoiceInfo.DocNo)

	// Reset SaleInvoice status to available (PackingStatus = 0, PackingProgress = 0, IsClose = false)
	// และรีเซ็ต close job information
	saleInvoiceInfo.PackingStatus = 0
	saleInvoiceInfo.PackingProgress = 0
	saleInvoiceInfo.IsClose = false
	saleInvoiceInfo.CloseJobBy = ""
	saleInvoiceInfo.CloseJobAt = nil

	// Save back using UpdateSaleInvoice method
	err = h.saleInvoiceSvc.UpdateSaleInvoice(shopID, saleInvoiceInfo.GuidFixed, authUsername, saleInvoiceInfo.SaleInvoice)
	if err != nil {
		h.ms.Logger.Errorf("Error updating SaleInvoice status %v: %v", saleInvoiceId, err)
		// Don't return error here as deletion was successful, just log warning
		h.ms.Logger.Warnf("Warning: Could not reset SaleInvoice status for %s: %v", saleInvoiceInfo.DocNo, err)
	} else {
		h.ms.Logger.Infof("Successfully reset SaleInvoice %s status to available", saleInvoiceInfo.DocNo)
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      saleInvoiceId,
		Data: map[string]interface{}{
			"message":                 "Successfully cancelled all Pickandpack documents and reset SaleInvoice status",
			"saleInvoiceDocNo":        saleInvoiceInfo.DocNo,
			"saleInvoiceId":           saleInvoiceId,
			"deletedPickandpackCount": len(pickandpackGUIDs),
			"deletedDocuments":        deletedDocs,
			"statusReset": map[string]interface{}{
				"packingStatus":   0,
				"packingProgress": 0,
				"isClose":         false,
				"closeJobBy":      "",
				"closeJobAt":      nil,
				"resetBy":         authUsername,
				"resetTime":       time.Now().UTC(),
			},
		},
	})

	return nil
}

// Helper functions to calculate totals for each warehouse group
func calculateTotalValue(details []transmodels.Detail) float64 {
	total := 0.0
	for _, detail := range details {
		total += detail.SumAmount
	}
	return total
}

func calculateTotalAmount(details []transmodels.Detail) float64 {
	total := 0.0
	for _, detail := range details {
		total += detail.SumAmount
	}
	return total
}

func calculateTotalQty(details []transmodels.Detail) float64 {
	total := 0.0
	for _, detail := range details {
		total += detail.Qty
	}
	return total
}

// Update Pickandpack Status godoc
// @Description Update Pickandpack Status
// @Tags		Pickandpack
// @Param		id  path      string  true  "Pickandpack ID (guidfixed)"
// @Param		status  path      int  true  "PackStatus (0=PENDING, 1=PROCESSING, 2=COMPLETED, 3=CANCELLED)"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/pickandpack/status/{id}/{status} [put]
func (h PickandpackHttp) UpdatePickandpackStatus(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	statusParam := ctx.Param("status")

	// Parse status to int8
	status := int8(0)
	if statusParam == "1" {
		status = 1 // PROCESSING
	} else if statusParam == "2" {
		status = 2 // COMPLETED
	} else if statusParam == "3" {
		status = 3 // CANCELLED
	}

	// Get existing document
	doc, err := h.svc.InfoPickandpack(shopID, id)
	if err != nil {
		h.ms.Logger.Errorf("Error getting Pickandpack document %v: %v", id, err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	// Update only the PackStatus
	doc.PackStatus = status

	// Convert PickandpackInfo to Pickandpack for update
	updateDoc := doc.Pickandpack

	// Save the updated document
	err = h.svc.UpdatePickandpack(shopID, id, authUsername, updateDoc)
	if err != nil {
		h.ms.Logger.Errorf("Error updating Pickandpack status %v: %v", id, err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	h.ms.Logger.Infof("Updated Pickandpack %s status to %d by user %s", id, status, authUsername)

	// หลังจากอัพเดท Pickandpack status - คำนวณ progress ของ SaleInvoice ใหม่
	err = h.updateSaleInvoicePackingStatus(shopID, doc.RefSaleInvoice, doc.GuidFixed, -1, -1, authUsername) // -1 = auto calculate
	if err != nil {
		h.ms.Logger.Warnf("Warning: Could not update SaleInvoice packing progress for %s: %v", doc.RefSaleInvoice, err)
		// Don't return error, just log warning as Pickandpack update was successful
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      id,
		Data: map[string]interface{}{
			"status":      status,
			"statusText":  getStatusText(status),
			"updatedBy":   authUsername,
			"updatedTime": time.Now().UTC(),
		},
	})

	return nil
}

// Update Pickandpack Details godoc
// @Description Update Pickandpack Details
// @Tags		Pickandpack
// @Param		id  path      string  true  "Pickandpack ID (guidfixed)"
// @Param		details  body      []transmodels.Detail  true  "Updated Details"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/pickandpack/details/{id} [put]
func (h PickandpackHttp) UpdatePickandpackDetails(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	// Parse the new details from request body
	var newDetails []transmodels.Detail
	err := json.Unmarshal([]byte(input), &newDetails)
	if err != nil {
		ctx.ResponseError(400, fmt.Sprintf("Invalid details format: %v", err))
		return err
	}

	// Get existing document
	doc, err := h.svc.InfoPickandpack(shopID, id)
	if err != nil {
		h.ms.Logger.Errorf("Error getting Pickandpack document %v: %v", id, err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	// Update details and recalculate totals
	doc.Details = &newDetails
	doc.TotalValue = calculateTotalValue(newDetails)
	doc.TotalAmount = calculateTotalAmount(newDetails)
	doc.TotalQty = calculateTotalQty(newDetails)

	// Convert PickandpackInfo to Pickandpack for update
	updateDoc := doc.Pickandpack

	// Save the updated document
	err = h.svc.UpdatePickandpack(shopID, id, authUsername, updateDoc)
	if err != nil {
		h.ms.Logger.Errorf("Error updating Pickandpack details %v: %v", id, err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	h.ms.Logger.Infof("Updated Pickandpack %s details (%d items) by user %s", id, len(newDetails), authUsername)

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      id,
		Data: map[string]interface{}{
			"itemCount":   len(newDetails),
			"totalValue":  doc.TotalValue,
			"totalAmount": doc.TotalAmount,
			"totalQty":    doc.TotalQty,
			"updatedBy":   authUsername,
			"updatedTime": time.Now().UTC(),
		},
	})

	return nil
}

// Helper function to get status text
func getStatusText(status int8) string {
	switch status {
	case 0:
		return "PENDING"
	case 1:
		return "PROCESSING"
	case 2:
		return "COMPLETED"
	case 3:
		return "CANCELLED"
	default:
		return "UNKNOWN"
	}
}

// Update Pickandpack godoc
// @Description Update Pickandpack
// @Tags		Pickandpack
// @Param		id  path      string  true  "Pickandpack ID"
// @Param		Pickandpack  body      models.Pickandpack  true  "Pickandpack"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/pickandpack/{id} [put]
func (h PickandpackHttp) UpdatePickandpack(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.Pickandpack{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdatePickandpack(shopID, id, authUsername, *docReq)

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

// Delete Pickandpack godoc
// @Description Delete Pickandpack
// @Tags		Pickandpack
// @Param		id  path      string  true  "Pickandpack ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/pickandpack/{id} [delete]
func (h PickandpackHttp) DeletePickandpack(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeletePickandpack(shopID, id, authUsername)

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

// Delete Pickandpack godoc
// @Description Delete Pickandpack
// @Tags		Pickandpack
// @Param		Pickandpack  body      []string  true  "Pickandpack GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/pickandpack [delete]
func (h PickandpackHttp) DeletePickandpackByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeletePickandpackByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get Pickandpack godoc
// @Description get Pickandpack info by guidfixed
// @Tags		Pickandpack
// @Param		id  path      string  true  "Pickandpack guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/pickandpack/{id} [get]
func (h PickandpackHttp) InfoPickandpack(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get Pickandpack %v", id)
	doc, err := h.svc.InfoPickandpack(shopID, id)

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

// Get Pickandpack By Code godoc
// @Description get Pickandpack info by Code
// @Tags		Pickandpack
// @Param		code  path      string  true  "Pickandpack Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/pickandpack/code/{code} [get]
func (h PickandpackHttp) InfoPickandpackByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")

	doc, err := h.svc.InfoPickandpackByCode(shopID, code)

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

// List Pickandpack step godoc
// @Description get list step
// @Tags		Pickandpack
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
// @Router /transaction/pickandpack [get]
func (h PickandpackHttp) SearchPickandpackPage(ctx microservice.IContext) error {
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

	filterMap := make(map[string]interface{})
	for k, v := range filters {
		filterMap[k] = v
	}
	filterMap["iscancel"] = false
	filterMap["$or"] = []map[string]interface{}{
		{"isclose": false},
		{"isclose": nil},
		{"isclose": map[string]interface{}{"$exists": false}},
	}
	filterMap["isclose"] = false

	docList, pagination, err := h.svc.SearchPickandpack(shopID, filterMap, pageable)

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

// List Pickandpack godoc
// @Description search limit offset
// @Tags		Pickandpack
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
// @Router /transaction/pickandpack/list [get]
func (h PickandpackHttp) SearchPickandpackStep(ctx microservice.IContext) error {
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

	docList, total, err := h.svc.SearchPickandpackStep(shopID, lang, filters, pageableStep)

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

// Create Pickandpack Bulk godoc
// @Description Create Pickandpack
// @Tags		Pickandpack
// @Param		Pickandpack  body      []models.Pickandpack  true  "Pickandpack"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/pickandpack/bulk [post]
func (h PickandpackHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.Pickandpack{}
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

// Get Pickandpack History godoc
// @Description Get SaleInvoice history that is completed and closed, including Pickandpack details
// @Tags       Pickandpack
// @Param      q         query   string  false  "Search Keyword (custcode, docno)"
// @Param      fromdate  query   string  false  "from date"
// @Param      todate    query   string  false  "to date"
// @Param      page      query   integer false  "Page"
// @Param      limit     query   integer false  "Limit"
// @Accept     json
// @Success    200 {array} common.ApiResponse
// @Failure    401 {object} common.AuthResponseFailed
// @Security   AccessToken
// @Router /transaction/pickandpack/history [get]
func (h PickandpackHttp) GetPickandpackHistory(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	// Filter SaleInvoice: packingstatus = 1, isclose = true
	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "-",
			Field: "docdatetime",
			Type:  requestfilter.FieldTypeRangeDate,
		},
	})
	filters["packingstatus"] = 1
	filters["isclose"] = true
	filters["iscancel"] = false

	saleInvoiceList, pagination, err := h.saleInvoiceSvc.SearchSaleInvoice(shopID, filters, pageable)
	if err != nil {
		h.ms.Logger.Errorf("Error searching SaleInvoice: %v", err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	// Get all Pickandpack documents for these SaleInvoices
	pickandpackMap := make(map[string][]models.PickandpackInfo)
	for _, invoice := range saleInvoiceList {
		ppList, _, err := h.svc.SearchPickandpack(shopID, map[string]interface{}{"refsaleinvoice": invoice.DocNo}, utils.GetPageable(func(key string) string {
			if key == "page" {
				return "1"
			}
			if key == "limit" {
				return "1000"
			}
			return ""
		}))
		if err == nil {
			pickandpackMap[invoice.DocNo] = ppList
		}
	}

	// Compose result with details
	var result []map[string]interface{}
	for _, invoice := range saleInvoiceList {
		ppList := pickandpackMap[invoice.DocNo]
		var pickandpackDetails []map[string]interface{}
		for _, pp := range ppList {
			pickandpackDetails = append(pickandpackDetails, map[string]interface{}{
				"docNo":         pp.DocNo,
				"id":            pp.GuidFixed,
				"packStatus":    pp.PackStatus,
				"whCode":        pp.WhCode,
				"whNames":       pp.WhNames,
				"locationCode":  pp.LocationCode,
				"locationNames": pp.LocationNames,
				"totalQty":      pp.TotalQty,
				"totalAmount":   pp.TotalAmount,
				"details":       pp.Details, // include details here
			})
		}
		result = append(result, map[string]interface{}{
			"saleInvoiceDocNo": invoice.DocNo,
			"saleInvoice":      invoice,
			"pickandpackList":  pickandpackDetails,
		})
	}

	pagination.Total = int64(len(result))
	ctx.Response(http.StatusOK, common.ApiResponse{
		Success:    true,
		Data:       result,
		Pagination: pagination,
	})
	return nil
}

// Get Pickandpack Dashboard godoc
// @Description Get dashboard statistics for Pickandpack: available SaleInvoice count and completed/closed job count in date range
// @Tags       Pickandpack
// @Param      fromdate  query   string  false  "from date"
// @Param      todate    query   string  false  "to date"
// @Accept     json
// @Success    200 {object} common.ApiResponse
// @Failure    401 {object} common.AuthResponseFailed
// @Security   AccessToken
// @Router /transaction/pickandpack/dashboard [get]
func (h PickandpackHttp) GetPickandpackDashboard(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	// Generate filters for date range
	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "-",
			Field: "docdatetime",
			Type:  requestfilter.FieldTypeRangeDate,
		},
	})

	// Count Available SaleInvoice (packingstatus = 0 or null)
	availableFilters := make(map[string]interface{})
	for k, v := range filters {
		availableFilters[k] = v
	}
	availableFilters["iscancel"] = false
	// เงื่อนไข packingstatus สำหรับ available
	availableFilters["$or"] = []map[string]interface{}{
		{"packingstatus": 0},
		{"packingstatus": nil},
		{"packingstatus": map[string]interface{}{"$exists": false}},
	}

	// Search for Available SaleInvoice documents
	availableSaleInvoiceList, _, err := h.saleInvoiceSvc.SearchSaleInvoice(shopID, availableFilters, utils.GetPageable(func(key string) string {
		if key == "page" {
			return "1"
		}
		if key == "limit" {
			return "10000"
		}
		return ""
	}))
	if err != nil {
		h.ms.Logger.Errorf("Error searching available SaleInvoice: %v", err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	availableCount := len(availableSaleInvoiceList)

	// Count Completed/Closed jobs (packingstatus = 1, isclose = true)
	completedFilters := make(map[string]interface{})
	for k, v := range filters {
		completedFilters[k] = v
	}
	completedFilters["packingstatus"] = 1
	completedFilters["isclose"] = true
	completedFilters["iscancel"] = false

	completedSaleInvoiceList, _, err := h.saleInvoiceSvc.SearchSaleInvoice(shopID, completedFilters, utils.GetPageable(func(key string) string {
		if key == "page" {
			return "1"
		}
		if key == "limit" {
			return "10000"
		}
		return ""
	}))
	if err != nil {
		h.ms.Logger.Errorf("Error searching completed SaleInvoice: %v", err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	completedCount := len(completedSaleInvoiceList)

	// Count In-Progress jobs (packingstatus = 1, isclose = false)
	inProgressFilters := make(map[string]interface{})
	for k, v := range filters {
		inProgressFilters[k] = v
	}
	inProgressFilters["packingstatus"] = 1
	inProgressFilters["isclose"] = false
	inProgressFilters["iscancel"] = false

	inProgressSaleInvoiceList, _, err := h.saleInvoiceSvc.SearchSaleInvoice(shopID, inProgressFilters, utils.GetPageable(func(key string) string {
		if key == "page" {
			return "1"
		}
		if key == "limit" {
			return "10000"
		}
		return ""
	}))
	if err != nil {
		h.ms.Logger.Errorf("Error searching in-progress SaleInvoice: %v", err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	inProgressCount := len(inProgressSaleInvoiceList)

	// Count Total SaleInvoice (ไม่รวมที่ยกเลิก) สำหรับคำนวณ percentage
	totalFilters := make(map[string]interface{})
	for k, v := range filters {
		totalFilters[k] = v
	}
	totalFilters["iscancel"] = false

	totalSaleInvoiceList, _, err := h.saleInvoiceSvc.SearchSaleInvoice(shopID, totalFilters, utils.GetPageable(func(key string) string {
		if key == "page" {
			return "1"
		}
		if key == "limit" {
			return "10000"
		}
		return ""
	}))
	if err != nil {
		h.ms.Logger.Errorf("Error searching total SaleInvoice: %v", err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	totalCount := len(totalSaleInvoiceList)

	// คำนวณ percentage (ป้องกัน division by zero)
	var availablePercentage, completedPercentage, inProgressPercentage float64
	if totalCount > 0 {
		availablePercentage = float64(availableCount) / float64(totalCount) * 100
		completedPercentage = float64(completedCount) / float64(totalCount) * 100
		inProgressPercentage = float64(inProgressCount) / float64(totalCount) * 100
	}

	// Prepare dashboard data
	dashboardData := map[string]interface{}{
		"availableSaleInvoice": map[string]interface{}{
			"count":       availableCount,
			"description": "จำนวนเอกสาร SaleInvoice ที่พร้อมสำหรับการจัด (PackingStatus = 0 หรือ null)",
		},
		"completedJobs": map[string]interface{}{
			"count":       completedCount,
			"description": "จำนวนเอกสารที่จัดเสร็จและปิดงานแล้ว (PackingStatus = 2, IsClose = true)",
		},
		"inProgressJobs": map[string]interface{}{
			"count":       inProgressCount,
			"description": "จำนวนเอกสารที่กำลังจัด (PackingStatus = 1, IsClose = false)",
		},
		"totalSaleInvoice": map[string]interface{}{
			"count":       totalCount,
			"description": "จำนวนเอกสาร SaleInvoice ทั้งหมดในช่วงวันที่เลือก (ไม่รวมที่ยกเลิก)",
		},
		"summary": map[string]interface{}{
			"availablePercentage":  availablePercentage,
			"completedPercentage":  completedPercentage,
			"inProgressPercentage": inProgressPercentage,
		},
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    dashboardData,
	})
	return nil
}

// GetWarehouseDashboard godoc
// @Description Get warehouse dashboard statistics for pickandpack documents
// @Tags		Pickandpack
// @Param		whcodes			query	string	false	"Warehouse codes (comma separated)"
// @Param		locationcodes	query	string	false	"Location codes (comma separated)"
// @Param		fromdate		query	string	false	"From date (YYYY-MM-DD)"
// @Param		todate			query	string	false	"To date (YYYY-MM-DD)"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/pickandpack/dashboard/warehouse [get]
func (h PickandpackHttp) GetWarehouseDashboard(ctx microservice.IContext) error {
	h.ms.Logger.Debugf("GetWarehouseDashboard called")
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	h.ms.Logger.Debugf("GetWarehouseDashboard shopID: %s", shopID)

	// Parse parameters
	whcodesParam := ctx.QueryParam("whcodes")
	locationcodesParam := ctx.QueryParam("locationcodes")
	fromDate := ctx.QueryParam("fromdate")
	toDate := ctx.QueryParam("todate")
	h.ms.Logger.Debugf("GetWarehouseDashboard params - whcodes: %s, locationcodes: %s, fromdate: %s, todate: %s", whcodesParam, locationcodesParam, fromDate, toDate)

	// Convert comma-separated strings to arrays
	var whcodes []string
	if whcodesParam != "" {
		whcodes = strings.Split(whcodesParam, ",")
		// Trim spaces from each element
		for i, code := range whcodes {
			whcodes[i] = strings.TrimSpace(code)
		}
	}

	var locationcodes []string
	if locationcodesParam != "" {
		locationcodes = strings.Split(locationcodesParam, ",")
		// Trim spaces from each element
		for i, code := range locationcodes {
			locationcodes[i] = strings.TrimSpace(code)
		}
	}

	// Call service method
	dashboardData, err := h.svc.GetWarehouseDashboard(shopID, whcodes, locationcodes, fromDate, toDate)
	if err != nil {
		h.ms.Logger.Errorf("Error getting warehouse dashboard: %v", err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    dashboardData,
	})
	return nil
}

// UpdatePrint godoc
// @Description Update pickandpack print status
// @Tags		Pickandpack
// @Param		docno	path	string	true	"Document Number"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/pickandpack/updateprint/{docno} [put]
func (h PickandpackHttp) UpdatePrint(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	docNo := ctx.Param("docno")
	if docNo == "" {
		ctx.ResponseError(http.StatusBadRequest, "docno is required")
		return fmt.Errorf("docno is required")
	}

	err := h.svc.UpdatePrint(shopID, docNo, authUsername)
	if err != nil {
		h.ms.Logger.Errorf("Error updating print status: %v", err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Message: "Print status updated successfully",
	})
	return nil
}

// ConfirmPickandpack godoc
// @Description Confirm pickandpack document
// @Tags		Pickandpack
// @Param		docno	path	string	true	"Document Number"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/pickandpack/confirmpickandpack/{docno} [put]
func (h PickandpackHttp) ConfirmPickandpack(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	docNo := ctx.Param("docno")
	if docNo == "" {
		ctx.ResponseError(http.StatusBadRequest, "docno is required")
		return fmt.Errorf("docno is required")
	}

	err := h.svc.ConfirmPickandpack(shopID, docNo, authUsername)
	if err != nil {
		h.ms.Logger.Errorf("Error confirming pickandpack: %v", err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Message: "Pickandpack confirmed successfully",
	})
	return nil
}

// CancelPickandpack godoc
// @Description Cancel pickandpack document
// @Tags		Pickandpack
// @Param		docno	path	string	true	"Document Number"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/pickandpack/cancel/{docno} [put]
func (h PickandpackHttp) CancelPickandpack(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	docNo := ctx.Param("docno")
	if docNo == "" {
		ctx.ResponseError(http.StatusBadRequest, "docno is required")
		return fmt.Errorf("docno is required")
	}

	err := h.svc.CancelPickandpack(shopID, docNo, authUsername)
	if err != nil {
		h.ms.Logger.Errorf("Error cancelling pickandpack: %v", err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Message: "Pickandpack cancelled successfully",
	})
	return nil
}
