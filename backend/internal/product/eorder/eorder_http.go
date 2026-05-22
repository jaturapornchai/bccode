package eorder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	salechannel_repositories "smlcloudplatform/internal/channel/salechannel/repositories"
	"smlcloudplatform/internal/config"
	couponRepositories "smlcloudplatform/internal/coupon/repositories"
	couponServices "smlcloudplatform/internal/coupon/services"
	creditorRepo "smlcloudplatform/internal/debtaccount/creditor/repositories"
	repoCust "smlcloudplatform/internal/debtaccount/debtor/repositories"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	notify_repositories "smlcloudplatform/internal/notify/repositories"
	notify_services "smlcloudplatform/internal/notify/services"
	repo_order_device "smlcloudplatform/internal/order/device/repositories"
	repo_order_setting "smlcloudplatform/internal/order/setting/repositories"
	branch_repositories "smlcloudplatform/internal/organization/branch/repositories"
	repo_media "smlcloudplatform/internal/pos/media/repositories"
	"smlcloudplatform/internal/product/eorder/models"
	"smlcloudplatform/internal/product/eorder/services"
	productmaster "smlcloudplatform/internal/product/product/repositories"
	product_repo "smlcloudplatform/internal/product/productbarcode/repositories"
	product_services "smlcloudplatform/internal/product/productbarcode/services"
	category_repositories "smlcloudplatform/internal/product/productcategory/repositories"
	category_services "smlcloudplatform/internal/product/productcategory/services"
	unit_repositories "smlcloudplatform/internal/product/unit/repositories"
	unitmaster "smlcloudplatform/internal/product/unit/repositories"
	unit_services "smlcloudplatform/internal/product/unit/services"
	"smlcloudplatform/internal/restaurant/kitchen"
	"smlcloudplatform/internal/restaurant/table"
	"smlcloudplatform/internal/restaurant/zone"
	"smlcloudplatform/internal/shop"
	saleinvoice_repositories "smlcloudplatform/internal/transaction/saleinvoice/repositories"
	saleinvoice_services "smlcloudplatform/internal/transaction/saleinvoice/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	warehouse_repositories "smlcloudplatform/internal/warehouse/repositories"
	"smlcloudplatform/pkg/microservice"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type IEOrderHttp interface{}

type EOrderHttp struct {
	ms             *microservice.Microservice
	cfg            config.IConfig
	svcCategory    category_services.IProductCategoryHttpService
	svcProduct     product_services.IProductBarcodeHttpService
	svcEOrder      services.EOrderService
	svcZone        zone.IZoneService
	svcTable       table.TableService
	svcKitchen     kitchen.IKitchenService
	svcSaleInvoice saleinvoice_services.ISaleInvoiceService
	svcNotify      notify_services.INotifyHttpService
}

func NewEOrderHttp(ms *microservice.Microservice, cfg config.IConfig) EOrderHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	pstClickHouse := ms.ClickHousePersister(cfg.ClickHouseConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	prod := ms.Producer(cfg.MQConfig())

	creditorRepo := creditorRepo.NewCreditorRepository(pst)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	repo := product_repo.NewProductBarcodeRepository(pst, cache)
	repoMaster := productmaster.NewProductRepository(pst)
	unitmaster := unitmaster.NewUnitRepository(pst)
	repoCategory := category_repositories.NewProductCategoryRepository(pst)
	svcCategory := category_services.NewProductCategoryHttpService(repoCategory, masterSyncCacheRepo, repo)

	clickHouseRepo := product_repo.NewProductBarcodeClickhouseRepository(pstClickHouse)
	mqRepo := product_repo.NewProductBarcodeMessageQueueRepository(prod)

	// Price History Service
	priceHistoryRepo := product_repo.NewProductPriceHistoryRepository(pst)
	priceHistorySvc := product_services.NewProductPriceHistoryService(priceHistoryRepo, utils.NewGUID, time.Now)

	// Warehouse Repository
	warehouseRepo := warehouse_repositories.NewWarehouseRepository(pst)

	// Unit Service
	unitMqRepo := unit_repositories.NewUnitMessageQueueRepository(prod)
	unitSvc := unit_services.NewUnitHttpService(unitmaster, repo, unitMqRepo, masterSyncCacheRepo)

	svcProduct := product_services.NewProductBarcodeHttpService(repo, repoMaster, unitmaster, unitSvc, *creditorRepo, mqRepo, clickHouseRepo, svcCategory, masterSyncCacheRepo, priceHistorySvc, warehouseRepo)

	repoCust := repoCust.NewDebtorRepository(pst)
	repoShop := shop.NewShopRepository(pst)
	repoTable := table.NewTableRepository(pst)
	repoOrder := repo_order_setting.NewSettingRepository(pst)
	repoDevice := repo_order_device.NewDeviceRepository(pst)
	repoMedia := repo_media.NewMediaRepository(pst)
	repoKitchen := kitchen.NewKitchenRepository(pst)
	repoSaleChannel := salechannel_repositories.NewSaleChannelRepository(pst)
	repoBranch := branch_repositories.NewBranchRepository(pst)

	repoZone := zone.NewZoneRepository(pst)

	svcZone := zone.NewZoneService(repoZone, masterSyncCacheRepo)

	repoSaleInvoice := saleinvoice_repositories.NewSaleInvoiceRepository(pst)

	// Initialize coupon service for SaleInvoice
	couponRepo := couponRepositories.NewCouponRepository(pst)
	couponReservationRepo := couponRepositories.NewCouponReservationRepository(pst)
	couponUsageHistoryRepo := couponRepositories.NewCouponUsageHistoryRepository(pst)
	couponService := couponServices.NewCouponHttpService(couponRepo, couponReservationRepo, couponUsageHistoryRepo, repo, masterSyncCacheRepo)

	svcSaleInvoice := saleinvoice_services.NewSaleInvoiceService(
		repoSaleInvoice,
		repoCust,
		nil,           // trans_cache.ICacheRepository
		nil,           // productbarcode_repositories.IProductBarcodeRepository
		nil,           // repoCust.IPointTransactionRepository
		nil,           // saleinvoice_repositories.ISaleInvoiceMessageQueueRepository
		nil,           // mastersync.IMasterSyncCacheRepository
		couponService, // coupon service
		nil,           // saleinvoice_services.ISaleInvocieParser
		nil,           // saleinvoice_services.ISaleInvoiceExport
	)

	svcTable := table.NewTableService(repoTable, masterSyncCacheRepo)
	svcKitchen := kitchen.NewKitchenService(repoKitchen, masterSyncCacheRepo)

	repoNotify := notify_repositories.NewNotifyRepository(pst)
	svcNotify := notify_services.NewNotifyHttpService(repoNotify, masterSyncCacheRepo, 15*time.Second)

	svcEOrder := services.NewEOrderService(
		repoShop,
		repoTable,
		repoOrder,
		repoMedia,
		repoKitchen,
		repoDevice,
		repoSaleChannel,
		repoBranch,
		repoNotify,
	)
	return EOrderHttp{
		ms:             ms,
		cfg:            cfg,
		svcCategory:    svcCategory,
		svcProduct:     svcProduct,
		svcEOrder:      svcEOrder,
		svcZone:        svcZone,
		svcTable:       *svcTable,
		svcKitchen:     svcKitchen,
		svcSaleInvoice: svcSaleInvoice,
		svcNotify:      svcNotify,
	}
}

func (h EOrderHttp) RegisterHttp() {

	h.ms.GET("/e-order/category", h.SearchProductCategoryPage)
	// h.ms.GET("/e-order/product", h.SearchProductBarcodePage)
	h.ms.GET("/e-order/product-barcode", h.SearchProductBarcodePage)
	h.ms.POST("/e-order/product-barcode", h.SearchProductBarcodeManyBarcodePage)

	h.ms.GET("/e-order/shop-info/v1.1", h.ShopInfo)
	h.ms.GET("/e-order/shop-info", h.ShopInfoOld)
	h.ms.GET("/e-order/restaurant/zone", h.SearchZone)
	h.ms.GET("/e-order/restaurant/kitchen", h.SearchKitchen)
	h.ms.GET("/e-order/restaurant/table", h.SearchTable)
	h.ms.GET("/e-order/sale-invoice/last-pos-docno", h.GetLastPOSDocNo)
	h.ms.GET("/e-order/notify", h.Notify)
	h.ms.POST("/line-notify", h.LineNotify)
}

// List Product Category
// @Description List Product Category
// @Tags		E-Order
// @Param		shopid		query	string		false  "Shop ID"
// @Param		q		query	string		false  "Search Value"
// @Param		group-number		query	int		false  "group number"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Router /e-order/category [get]
func (h EOrderHttp) SearchProductCategoryPage(ctx microservice.IContext) error {
	shopID := ctx.QueryParam("shopid")

	if len(shopID) == 0 {
		ctx.ResponseError(http.StatusBadRequest, "shopid is empty")
		return nil
	}

	pageable := utils.GetPageable(ctx.QueryParam)
	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "group-number",
			Field: "group_number",
			Type:  requestfilter.FieldTypeInt,
		},
	})
	docList, pagination, err := h.svcCategory.SearchProductCategory(shopID, filters, pageable)

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

// List Product
// @Description List Product
// @Tags		E-Order
// @Param		shopid		query	string		false  "Shop ID"
// @Param		barcodes		query	string		false  "barcode json array"
// @Param		isalacarte		query	string		false  "is A La Carte"
// @Param		ordertypes		query	string		false  "order types ex. a01,a02"
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Router /e-order/product-barcode [get]
func (h EOrderHttp) SearchProductBarcodePage(ctx microservice.IContext) error {
	shopID := ctx.QueryParam("shopid")

	if len(shopID) == 0 {
		ctx.ResponseError(http.StatusBadRequest, "shopid is empty")
		return nil
	}

	rawBarcodes := ctx.QueryParam("barcodes")

	barcodes := []string{}
	if len(rawBarcodes) > 0 {
		json.Unmarshal([]byte(rawBarcodes), &barcodes)
	}

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "isalacarte",
			Field: "isalacarte",
			Type:  requestfilter.FieldTypeBoolean,
		},
		{
			Param: "ordertypes",
			Field: "ordertypes.code",
			Type:  requestfilter.FieldTypeString,
		},
	})

	if len(barcodes) > 0 {
		filters["barcode"] = bson.M{"$in": barcodes}
	}

	pageable := utils.GetPageable(ctx.QueryParam)
	docList, pagination, err := h.svcProduct.SearchProductBarcode(shopID, filters, pageable)

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

// List Product
// @Description List Product
// @Tags		E-Order
// @Param		shopid		query	string		false  "Shop ID"
// @Param		barcodes		body	[]string	false  "barcode json array"
// @Param		isalacarte		query	string		false  "is A La Carte"
// @Param		ordertypes		query	string		false  "order types ex. a01,a02"
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Router /e-order/product-barcode [post]
func (h EOrderHttp) SearchProductBarcodeManyBarcodePage(ctx microservice.IContext) error {
	shopID := ctx.QueryParam("shopid")

	if len(shopID) == 0 {
		ctx.ResponseError(http.StatusBadRequest, "shopid is empty")
		return nil
	}

	rawBarcodes := ctx.ReadInput()

	barcodes := []string{}
	if len(rawBarcodes) > 0 {
		json.Unmarshal([]byte(rawBarcodes), &barcodes)
	}

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "isalacarte",
			Field: "isalacarte",
			Type:  requestfilter.FieldTypeBoolean,
		},
		{
			Param: "ordertypes",
			Field: "ordertypes.code",
			Type:  requestfilter.FieldTypeString,
		},
	})

	if len(barcodes) > 0 {
		filters["barcode"] = bson.M{"$in": barcodes}
	}

	pageable := utils.GetPageable(ctx.QueryParam)
	docList, pagination, err := h.svcProduct.SearchProductBarcode(shopID, filters, pageable)

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

// List Product By Barcodes
// @Description List Product By Barcodes
// @Tags		E-Order
// @Param		shopid		query	string		false  "Shop ID"
// @Param		barcodes		query	string		false  "barcode json array"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Router /e-order/product-barcode [get]
func (h EOrderHttp) GetProductBarcodeByBarcodes(ctx microservice.IContext) error {
	shopID := ctx.QueryParam("shopid")

	if len(shopID) == 0 {
		ctx.ResponseError(http.StatusBadRequest, "shopid is empty")
		return nil
	}

	rawBarcodes := ctx.QueryParam("barcodes")

	barcodes := []string{}
	if len(rawBarcodes) > 0 {
		json.Unmarshal([]byte(rawBarcodes), &barcodes)
	}

	docList, err := h.svcProduct.GetProductBarcodeByBarcodes(shopID, barcodes)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    docList,
	})
	return nil
}

// Get Shop Info
// @Description Get Shop Info
// @Tags		E-Order
// @Param		shopid		query	string		false  "Shop ID"
// @Param		order-station		query	string		false  "Order station code"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Router /e-order/shop-info [get]
func (h EOrderHttp) ShopInfoOld(ctx microservice.IContext) error {
	shopID := ctx.QueryParam("shopid")
	orderStationCode := ctx.QueryParam("order-station")

	if len(shopID) == 0 {
		ctx.ResponseError(http.StatusBadRequest, "shopid is empty")
		return nil
	}

	data, err := h.svcEOrder.GetShopInfoOld(shopID, orderStationCode)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    data,
	})
	return nil
}

// Get Shop Info v1.1
// @Description Get Shop Info v1.1
// @Tags		E-Order
// @Param		shopid		query	string		false  "Shop ID"
// @Param		order-station		query	string		false  "Order station code"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Router /e-order/shop-info/v1.1 [get]
func (h EOrderHttp) ShopInfo(ctx microservice.IContext) error {
	shopID := ctx.QueryParam("shopid")
	orderStationCode := ctx.QueryParam("order-station")

	if len(shopID) == 0 {
		ctx.ResponseError(http.StatusBadRequest, "shopid is empty")
		return nil
	}

	data, err := h.svcEOrder.GetShopInfo(shopID, orderStationCode)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    data,
	})
	return nil
}

// List E Order Restaurant Zone godoc
// @Description List Restaurant Zone Category
// @Tags		E-Order
// @Param		group-number	query	integer		false  "Group Number"
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Size"
// @Accept 		json
// @Success		200	{object}	models.ZonePageResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Router /e-order/restaurant/zone [get]
func (h EOrderHttp) SearchZone(ctx microservice.IContext) error {
	shopID := ctx.QueryParam("shopid")

	if len(shopID) == 0 {
		ctx.ResponseError(http.StatusBadRequest, "shopid is empty")
		return nil
	}

	pageable := utils.GetPageable(ctx.QueryParam)
	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "group-number",
			Field: "group_number",
			Type:  requestfilter.FieldTypeInt,
		},
	})

	docList, pagination, err := h.svcZone.SearchZone(shopID, filters, pageable)

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

// List E Order Restaurant Kitchen godoc
// @Description List Restaurant Kitchen Category
// @Tags		E-Order
// @Param		group-number	query	integer		false  "Group Number"
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Size"
// @Accept 		json
// @Success		200	{object}	models.KitchenPageResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Router /e-order/restaurant/kitchen [get]
func (h EOrderHttp) SearchKitchen(ctx microservice.IContext) error {
	shopID := ctx.QueryParam("shopid")

	if len(shopID) == 0 {
		ctx.ResponseError(http.StatusBadRequest, "shopid is empty")
		return nil
	}

	pageable := utils.GetPageable(ctx.QueryParam)

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "group-number",
			Field: "group_number",
			Type:  requestfilter.FieldTypeInt,
		},
	})

	docList, pagination, err := h.svcKitchen.SearchKitchen(shopID, filters, pageable)

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

// List E Order Restaurant Table godoc
// @Description List Restaurant Table
// @Tags		E-Order
// @Param		group-number	query	integer		false  "Group Number"
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Size"
// @Accept 		json
// @Success		200	{object}	models.TablePageResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Router /e-order/restaurant/table [get]
func (h EOrderHttp) SearchTable(ctx microservice.IContext) error {
	shopID := ctx.QueryParam("shopid")

	if len(shopID) == 0 {
		ctx.ResponseError(http.StatusBadRequest, "shopid is empty")
		return nil
	}

	pageable := utils.GetPageable(ctx.QueryParam)

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "group-number",
			Field: "group_number",
			Type:  requestfilter.FieldTypeInt,
		},
	})

	docList, pagination, err := h.svcTable.SearchTable(shopID, filters, pageable)

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

// Get E Order SaleInvoice Last DocNo godoc
// @Description get SaleInvoice Last DocNo
// @Tags		E-Order
// @Param		posid	query	string		false  "POS ID"
// @Param		maxdocno	query	string		false  "Max DocNo"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Router /e-order/sale-invoice/last-pos-docno [get]
func (h EOrderHttp) GetLastPOSDocNo(ctx microservice.IContext) error {
	shopID := ctx.QueryParam("shopid")
	if len(shopID) == 0 {
		ctx.ResponseError(http.StatusBadRequest, "shopid is empty")
		return nil
	}

	posID := ctx.QueryParam("posid")
	maxDocNo := ctx.QueryParam("maxdocno")

	if posID == "" || maxDocNo == "" {
		ctx.ResponseError(http.StatusBadRequest, "posid and maxdocno is required")
		return nil
	}

	doc, err := h.svcSaleInvoice.GetLastPOSDocNo(shopID, posID, maxDocNo)

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

func (h EOrderHttp) Test(ctx microservice.IContext) error {
	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})
	return nil
}

// List E Order Notify godoc
// @Description List Notify
// @Tags		E-Order
// @Param		type	query	string		false  "notify type"
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Size"
// @Accept 		json
// @Success		200	{object}	models.TablePageResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Router /e-order/notify [get]
func (h EOrderHttp) Notify(ctx microservice.IContext) error {
	shopID := ctx.QueryParam("shopid")

	if len(shopID) == 0 {
		ctx.ResponseError(http.StatusBadRequest, "shopid is empty")
		return nil
	}

	pageable := utils.GetPageable(ctx.QueryParam)

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "type",
			Field: "type",
			Type:  requestfilter.FieldTypeString,
		},
	})

	docList, pagination, err := h.svcNotify.SearchNotifyInfo(shopID, filters, pageable)

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

// List E Order Notify godoc
// @Description List Notify
// @Tags		E-Order
// @Param		LinePayload  body     models.LinePayload  true  "Line Payload"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Router /line-notify [post]
func (h EOrderHttp) LineNotify(ctx microservice.IContext) error {

	payload := models.LinePayload{}

	input := ctx.ReadInput()

	err := json.Unmarshal([]byte(input), &payload)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	url := "https://notify-api.line.me/api/notify"
	linePayload := bytes.NewBufferString("message=" + payload.Message)

	req, err := http.NewRequest("POST", url, linePayload)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+payload.Token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("received non-200 response status: %d", resp.StatusCode)
	}

	ctx.Response(http.StatusOK, map[string]interface{}{
		"success": true,
	})
	return nil
}
