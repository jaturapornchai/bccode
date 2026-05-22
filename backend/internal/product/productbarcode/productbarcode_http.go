package productbarcode

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"smlcloudplatform/internal/config"
	creditorRepo "smlcloudplatform/internal/debtaccount/creditor/repositories"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productmaster "smlcloudplatform/internal/product/product/repositories"
	"smlcloudplatform/internal/product/productbarcode/models"
	"smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/internal/product/productbarcode/services"
	productcategory_repositories "smlcloudplatform/internal/product/productcategory/repositories"
	productcategory_services "smlcloudplatform/internal/product/productcategory/services"
	unit_repositories "smlcloudplatform/internal/product/unit/repositories"
	unitmaster "smlcloudplatform/internal/product/unit/repositories"
	unit_services "smlcloudplatform/internal/product/unit/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	warehouse_repositories "smlcloudplatform/internal/warehouse/repositories"
	"smlcloudplatform/pkg/microservice"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

type IProductBarcodeHttp interface{}

type ProductBarcodeHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IProductBarcodeHttpService
}

func NewProductBarcodeHttp(ms *microservice.Microservice, cfg config.IConfig) ProductBarcodeHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	pstClickHouse := ms.ClickHousePersister(cfg.ClickHouseConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	prod := ms.Producer(cfg.MQConfig())
	repoMaster := productmaster.NewProductRepository(pst)
	unitmaster := unitmaster.NewUnitRepository(pst)
	creditorRepo := creditorRepo.NewCreditorRepository(pst)
	repo := repositories.NewProductBarcodeRepository(pst, cache)
	clickHouseRepo := repositories.NewProductBarcodeClickhouseRepository(pstClickHouse)
	mqRepo := repositories.NewProductBarcodeMessageQueueRepository(prod)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)

	productcategoryRepo := productcategory_repositories.NewProductCategoryRepository(pst)
	productcategorySvc := productcategory_services.NewProductCategoryHttpService(productcategoryRepo, masterSyncCacheRepo, repo)

	// Warehouse Repository for shelf search
	warehouseRepo := warehouse_repositories.NewWarehouseRepository(pst)

	// Price History Service
	priceHistoryRepo := repositories.NewProductPriceHistoryRepository(pst)
	priceHistorySvc := services.NewProductPriceHistoryService(priceHistoryRepo, utils.NewGUID, time.Now)

	// Unit Service
	unitMqRepo := unit_repositories.NewUnitMessageQueueRepository(prod)
	unitSvc := unit_services.NewUnitHttpService(unitmaster, repo, unitMqRepo, masterSyncCacheRepo)

	svc := services.NewProductBarcodeHttpService(repo, repoMaster, unitmaster, unitSvc, *creditorRepo, mqRepo, clickHouseRepo, productcategorySvc, masterSyncCacheRepo, priceHistorySvc, warehouseRepo)

	return ProductBarcodeHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h ProductBarcodeHttp) RegisterHttp() {

	h.ms.POST("/product/barcode/bulk", h.SaveBulk)
	h.ms.POST("/product/barcode/import", h.Import)

	h.ms.GET("/product/barcode", h.SearchProductBarcodePage)
	h.ms.GET("/product/barcode2", h.SearchProductBarcodePage2)
	h.ms.GET("/product/barcode/list", h.SearchProductBarcodeLimit)
	h.ms.POST("/product/barcode", h.CreateProductBarcode)
	h.ms.GET("/product/barcode/:id", h.InfoProductBarcode)
	h.ms.GET("/product/barcode/ref/:barcode", h.GetroductBarcodeByRef)
	h.ms.GET("/product/barcode/pk/:barcode", h.InfoProductBarcodeByBarcode)
	h.ms.GET("/product/barcode/by-code", h.InfoArray)
	h.ms.GET("/product/barcode/master", h.InfoArrayMaster)
	h.ms.PUT("/product/barcode/xsort", h.UpdateProductBarcodeXSort)
	h.ms.PUT("/product/barcode/:id", h.UpdateProductBarcode)
	h.ms.PUT("/product/barcode/branch", h.UpdateProductBarcodeBranch)
	h.ms.PUT("/product/barcode/business-type", h.UpdateProductBarcodeBusinessType)

	h.ms.DELETE("/product/barcode/:id", h.DeleteProductBarcode)
	h.ms.DELETE("/product/barcode", h.DeleteProductBarcodeByGUIDs)

	h.ms.GET("/product/barcode/units", h.GetroductBarcodeByAllUnits)
	h.ms.GET("/product/barcode/groups", h.GetroductBarcodeByGroups)

	h.ms.GET("/product/barcode/export", h.Export)

	h.ms.GET("/product/barcode/bom/:barcode", h.InfoBOMView)
	h.ms.GET("/product/barcode/price-history", h.GetPriceHistory)
	h.ms.GET("/product/barcode/price-history/:barcode", h.GetPriceHistoryByBarcode)
	h.ms.POST("/product/barcode/import-refbarcode", h.ImportRefBarcodeUpdate)
}

// Create ProductBarcode godoc
// @Description Create ProductBarcode
// @Tags		ProductBarcode
// @Param		ProductBarcode  body      models.ProductBarcodeRequest  true  "ProductBarcode"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/barcode [post]
func (h ProductBarcodeHttp) CreateProductBarcode(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.ProductBarcodeRequest{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if docReq.XSorts == nil {
		docReq.XSorts = &[]common.XSort{}
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateProductBarcode(shopID, authUsername, *docReq)

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

// Update ProductBarcode godoc
// @Description Update ProductBarcode
// @Tags		ProductBarcode
// @Param		id  path      string  true  "ProductBarcode ID"
// @Param		ProductBarcode  body      models.ProductBarcodeRequest  true  "ProductBarcode"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/barcode/{id} [put]
func (h ProductBarcodeHttp) UpdateProductBarcode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.ProductBarcodeRequest{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if docReq.XSorts == nil {
		docReq.XSorts = &[]common.XSort{}
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateProductBarcode(shopID, id, authUsername, *docReq)

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

// Update ProductBarcode Branch godoc
// @Description Update ProductBarcode Branch
// @Tags		ProductBarcode
// @Param		ProductBarcodeBranchRequest  body      models.ProductBarcodeBranchRequest  true  "Product BarcodeBranch Request"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/barcode/branch [put]
func (h ProductBarcodeHttp) UpdateProductBarcodeBranch(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	docReq := &models.ProductBarcodeBranchRequest{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateProductBarcodeBranch(shopID, authUsername, docReq.Branch, docReq.Products)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Update ProductBarcode Business Type  godoc
// @Description Update ProductBarcode Business Type
// @Tags		ProductBarcode
// @Param		ProductBarcodeBusinessTypeRequest  body      models.ProductBarcodeBusinessTypeRequest  true  "Product Barcode Business Type Request"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/barcode/business-type [put]
func (h ProductBarcodeHttp) UpdateProductBarcodeBusinessType(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	docReq := &models.ProductBarcodeBusinessTypeRequest{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateProductBarcodeBusinessType(shopID, authUsername, docReq.BusinessType, docReq.Products)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Update XSort	 ProductBarcode godoc
// @Description Update XSort ProductBarcode
// @Tags		ProductBarcode
// @Param		XSort  body      []common.XSortModifyReqesut  true  "XSort"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/barcode/xsort [put]
func (h ProductBarcodeHttp) UpdateProductBarcodeXSort(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	req := &[]common.XSortModifyReqesut{}
	err := json.Unmarshal([]byte(input), &req)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(req); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.XSortsSave(shopID, authUsername, *req)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Delete ProductBarcode godoc
// @Description Delete ProductBarcode
// @Tags		ProductBarcode
// @Param		id  path      string  true  "ProductBarcode ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/barcode/{id} [delete]
func (h ProductBarcodeHttp) DeleteProductBarcode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteProductBarcode(shopID, id, authUsername)

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

// Get ProductBarcode godoc
// @Description get struct array by ID
// @Tags		ProductBarcode
// @Param		id  path      string  true  "ProductBarcode ID"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/barcode/{id} [get]
func (h ProductBarcodeHttp) InfoProductBarcode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get ProductBarcode %v", id)
	doc, err := h.svc.InfoProductBarcode(shopID, id)

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

// Get ProductBarcode By Reference Barcode godoc
// @Description get by reference barcode
// @Tags		ProductBarcode
// @Param		barcode  path      string  true  "Reference Barcode"
// @Param		shopsid  query     string  false  "shopsid ex. s001,s002"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/barcode/ref/{barcode} [get]
func (h ProductBarcodeHttp) GetroductBarcodeByRef(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	refBarcode := ctx.Param("barcode")
	shopsidParam := strings.Trim(ctx.QueryParam("shopsid"), " ")

	// If shopsid parameter is provided, search across multiple shops
	if shopsidParam != "" {
		shopsList := strings.Split(shopsidParam, ",")
		for i, shop := range shopsList {
			shopsList[i] = strings.Trim(shop, " ")
		}

		// Use GetProductBarcodeByBarcodeRefMultiShops for multiple shops lookup
		docs, err := h.svc.GetProductBarcodeByBarcodeRefMultiShops(shopsList, refBarcode)

		if err != nil {
			ctx.ResponseError(http.StatusBadRequest, err.Error())
			return err
		}

		ctx.Response(http.StatusOK, common.ApiResponse{
			Success: true,
			Data:    docs,
		})
		return nil
	}

	// Single shop lookup (default behavior)
	docs, err := h.svc.GetProductBarcodeByBarcodeRef(shopID, refBarcode)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    docs,
	})
	return nil
}

// Get ProductBarcode By Barcode godoc
// @Description get data by barcode
// @Tags		ProductBarcode
// @Param		barcode  path      string  true  "Barcode"
// @Param		shopsid  query     string  false  "shopsid ex. s001,s002"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/barcode/pk/{barcode} [get]
func (h ProductBarcodeHttp) InfoProductBarcodeByBarcode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	barcode := ctx.Param("barcode")
	shopsidParam := strings.Trim(ctx.QueryParam("shopsid"), " ")

	// If shopsid parameter is provided, use it instead of user's shopID
	if shopsidParam != "" {
		shopsList := strings.Split(shopsidParam, ",")
		for i, shop := range shopsList {
			shopsList[i] = strings.Trim(shop, " ")
		}
		// Use the first shop from the list for single barcode lookup
		if len(shopsList) > 0 {
			shopID = shopsList[0]
		}
	}

	doc, err := h.svc.InfoProductBarcodeByBarcode(shopID, barcode)

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

// Get ProductBarcode By code array godoc
// @Description get ProductBarcode by code array
// @Tags		ProductBarcode
// @Param		codes	query	string		false  "Code filter, json array encode "
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/barcode/by-code [get]
func (h ProductBarcodeHttp) InfoArray(ctx microservice.IContext) error {
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

// Get Master ProductBarcode By code array godoc
// @Description get master ProductBarcode by code array
// @Tags		ProductBarcode
// @Param		codes	query	string		false  "Code filter, json array encode "
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/barcode/master [get]
func (h ProductBarcodeHttp) InfoArrayMaster(ctx microservice.IContext) error {
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

	// where to filter array master
	doc, err := h.svc.InfoWTFArrayMaster(docReq)

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

// List ProductBarcode godoc
// @Description get struct array by ID
// @Tags		ProductBarcode
// @Param		businesstypecode		query	string		false  "business type code ex. bt1,bt2"
// @Param		branchcode		query	string		false  "branch code ex. b1,b2"
// @Param		isalacarte		query	boolean		false  "is A La Carte"
// @Param		isusesubbarcodes		query	boolean		false  "is use sub barcodes"
// @Param		isbom		query	boolean		false  "is use BOM"
// @Param		ordertypes		query	string		false  "order types ex. a01,a02"
// @Param		itemtype		query	int8		false  "item type"
// @Param		shopsid		query	string		false  "shopsid ex. b1,b2"
// @Param		zeroprice		query	boolean		false  "filter products with standard price = 0"
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/barcode [get]
func (h ProductBarcodeHttp) SearchProductBarcodePage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	filters := h.searchFilter(ctx.QueryParam)

	docList, pagination, err := h.svc.SearchProductBarcode(shopID, filters, pageable)

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

// List ProductBarcode2 godoc
// @Description get struct array by ID
// @Tags		ProductBarcode2
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/barcode2 [get]
func (h ProductBarcodeHttp) SearchProductBarcodePage2(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)
	docList, pagination, err := h.svc.SearchProductBarcode2(shopID, pageable)

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

// List ProductBarcode godoc
// @Description search limit offset
// @Tags		ProductBarcode
// @Param		businesstypecode		query	string		false  "business type code ex. bt1,bt2"
// @Param		branchcode		query	string		false  "branch code ex. b1,b2"
// @Param		isalacarte		query	boolean		false  "is A La Carte"
// @Param		isusesubbarcodes		query	boolean		false  "is use sub barcodes"
// @Param		isbom		query	boolean		false  "is use BOM"
// @Param		ordertypes		query	string		false  "order types ex. a01,a02"
// @Param		itemtype		query	int8		false  "item type"
// @Param		shopsid		query	string		false  "shopsid ex. s001,s002"
// @Param		zeroprice		query	boolean		false  "filter products with standard price = 0"
// @Param		q		query	string		false  "Universal search - searches in all product codes and names (barcode, itemcode, groupcode, brandcode, modelcode, patterncode, etc. and all names)"
// @Param		warehousecode		query	string		false  "Warehouse Code for shelf search"
// @Param		locationcode		query	string		false  "Location Code for shelf search"
// @Param		shelfcode		query	string		false  "Shelf Code for shelf search"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/barcode/list [get]
func (h ProductBarcodeHttp) SearchProductBarcodeLimit(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")
	shopsidParam := strings.Trim(ctx.QueryParam("shopsid"), " ")

	filters := h.searchFilter(ctx.QueryParam)

	var docList []models.ProductBarcodeInfo
	var total int
	var err error

	// Handle multiple shop IDs
	if shopsidParam != "" {
		shopsList := strings.Split(shopsidParam, ",")
		for i, shop := range shopsList {
			shopsList[i] = strings.Trim(shop, " ")
		}

		// Remove any existing shopid filter and replace with multiple shops
		delete(filters, "shopid")
		filters["shopid"] = map[string]interface{}{
			"$in": shopsList,
		}

		// Use the multi-shop service method
		docList, total, err = h.svc.SearchProductBarcodeStepMultiShops(lang, filters, pageableStep)
	} else {
		// Use the single-shop service method
		docList, total, err = h.svc.SearchProductBarcodeStep(shopID, lang, filters, pageableStep)
	}

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

// Create ProductBarcode Bulk godoc
// @Description Create ProductBarcode
// @Tags		ProductBarcode
// @Param		ProductBarcode  body      []models.ProductBarcode  true  "ProductBarcode"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/barcode/bulk [post]
func (h ProductBarcodeHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.ProductBarcode{}
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

// Delete ProductBarcode By GUIDs godoc
// @Description Delete ProductBarcode
// @Tags		ProductBarcode
// @Param		ProductBarcode  body      []string  true  "ProductBarcode GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/barcode [delete]
func (h ProductBarcodeHttp) DeleteProductBarcodeByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteProductBarcodeByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get ProductBarcode By Reference Barcode godoc
// @Description get by reference barcode
// @Tags		ProductBarcode
// @Accept 		json
// @Param		codes	query	string		false  "array of units"
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/barcode/units [get]
func (h ProductBarcodeHttp) GetroductBarcodeByAllUnits(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	// inputBody := ctx.ReadInput()

	// unitCodes := []string{}
	// err := json.Unmarshal([]byte(inputBody), &unitCodes)

	// if err != nil {
	// 	ctx.ResponseError(400, err.Error())
	// 	return err
	// }

	reqUnitCodes := ctx.QueryParam("codes")
	unitCodes := []string{}

	tempUnitCodes := strings.Split(reqUnitCodes, ",")
	unitCodes = append(unitCodes, tempUnitCodes...)

	docs, pagination, err := h.svc.GetProductBarcodeByUnits(shopID, unitCodes, pageable)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success:    true,
		Pagination: pagination,
		Data:       docs,
	})
	return nil
}

// Get ProductBarcode By Groups
// @Description get by group codes
// @Tags		ProductBarcode
// @Accept 		json
// @Param		codes	query	string		false  "array of group"
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/barcode/groups [get]
func (h ProductBarcodeHttp) GetroductBarcodeByGroups(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	reqUnitCodes := ctx.QueryParam("codes")
	groupCodes := []string{}

	tempUnitCodes := strings.Split(reqUnitCodes, ",")
	groupCodes = append(groupCodes, tempUnitCodes...)

	docs, pagination, err := h.svc.GetProductBarcodeByGroups(shopID, groupCodes, pageable)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success:    true,
		Pagination: pagination,
		Data:       docs,
	})
	return nil
}

// Get  Export
// @Description ProductBarcode Export
// @Tags		ProductBarcode
// @Param		lang	query	string		false  "language code"
// @Param		barcode	query	string		false  "Label Barcode"
// @Param		productname	query	string		false  "Label Product Name"
// @Param		unitcode	query	string		false  "Label Unit Code"
// @Param		unitname	query	string		false  "Label Unit Name"
// @Param		price	query	string		false  "Label Price"
// @Param		itemtype	query	string		false  "Label Item Type"
// @Param		groupcode	query	string		false  "Label Group Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/barcode/export [get]
func (h ProductBarcodeHttp) Export(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	languageCode := ctx.QueryParam("lang")

	if languageCode == "" {
		languageCode = "en"
	}

	keyCols := []string{
		"barcode",     //บาร์โค้ด",
		"productname", //"ชื่อสินค้า",
		"unitcode",    //"หน่วยนับ",
		"unit_name",    //"ชื่อหน่วยนับ",
		"price",       //ราคาขาย",
		"item_type",    //ประเภทสินค้า",
		"group_code",   //กลุ่มสินค้า",
	}

	languageHeader := map[string]string{}
	for _, key := range keyCols {
		languageHeader[key] = ctx.QueryParam(key)
	}

	results, err := h.svc.Export(shopID, languageCode, languageHeader)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	fileName := fmt.Sprintf("%s_productbarcode_%s.csv", shopID, time.Now().Format("20060102150405"))

	ctx.EchoContext().Response().Header().Set(echo.HeaderContentType, "application/octet-stream; charset=UTF-8")
	ctx.EchoContext().Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename=\""+fileName+"\"")
	ctx.EchoContext().Response().WriteHeader(http.StatusOK)

	t := transform.NewWriter(ctx.EchoContext().Response(), unicode.UTF8BOM.NewEncoder())

	csvWriter := csv.NewWriter(t)
	defer csvWriter.Flush()

	for _, value := range results {

		err := csvWriter.Write(value)
		if err != nil {
			log.Fatal("Error writing record to CSV:", err)
			return err
		}
	}

	return nil
}

// Get ProductBarcode BOM godoc
// @Description get product barcode bom view information
// @Tags		ProductBarcode
// @Param		barcode  path      string  true  "Barcode"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/barcode/bom/{barcode} [get]
func (h ProductBarcodeHttp) InfoBOMView(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	barcode := ctx.Param("barcode")

	doc, err := h.svc.InfoBomView(shopID, barcode)

	if err != nil {
		h.ms.Logger.Errorf("Error getting document %v: %v", barcode, err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    doc,
	})
	return nil
}

func (h ProductBarcodeHttp) postProcessBooleanFilters(filters map[string]interface{}) map[string]interface{} {
	// สำหรับ boolean field ที่อาจไม่มีใน document เก่า
	// เปลี่ยน {field: false} เป็น {field: {$ne: true}} เพื่อ match ทั้ง false และ document ที่ไม่มี field
	booleanFieldsToFix := []string{"isusesubbarcodes", "isbom"}
	for _, field := range booleanFieldsToFix {
		if val, ok := filters[field]; ok {
			if boolVal, isBool := val.(bool); isBool && !boolVal {
				filters[field] = bson.M{"$ne": true}
			}
		}
	}
	return filters
}

func (h ProductBarcodeHttp) searchFilter(queryParam func(string) string) map[string]interface{} {
	filters := requestfilter.GenerateFilters(queryParam, []requestfilter.FilterRequest{
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
		{
			Param: "materialtype",
			Field: "materialtype",
			Type:  requestfilter.FieldTypeInt,
		},
		{
			Param: "item_type",
			Field: "item_type",
			Type:  requestfilter.FieldTypeInt,
		},
		{
			Param: "businesstypecode",
			Field: "businesstypes.code",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "branchcode",
			Field: "branches.code",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "isusesubbarcodes",
			Field: "isusesubbarcodes",
			Type:  requestfilter.FieldTypeBoolean,
		},
		{
			Param: "shopsid",
			Field: "shopid",
			Type:  requestfilter.FieldTypeString,
		},
		// Group filters
		{
			Param: "groupguid",
			Field: "groupguid",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "group_code",
			Field: "group_code",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "groupsuboneguid",
			Field: "groupsuboneguid",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "groupsubonecode",
			Field: "groupsubonecode",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "groupsubtwoguid",
			Field: "groupsubtwoguid",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "groupsubtwocode",
			Field: "groupsubtwocode",
			Type:  requestfilter.FieldTypeString,
		},
		// Group Names filters
		{
			Param: "groupnames.th",
			Field: "groupnames.th",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "groupnames.en",
			Field: "groupnames.en",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "groupnames.cn",
			Field: "groupnames.cn",
			Type:  requestfilter.FieldTypeString,
		},
		// Group Sub One Names filters
		{
			Param: "groupsubonenames.th",
			Field: "groupsubonenames.th",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "groupsubonenames.en",
			Field: "groupsubonenames.en",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "groupsubonenames.cn",
			Field: "groupsubonenames.cn",
			Type:  requestfilter.FieldTypeString,
		},
		// Group Sub Two Names filters
		{
			Param: "groupsubtwonames.th",
			Field: "groupsubtwonames.th",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "groupsubtwonames.en",
			Field: "groupsubtwonames.en",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "groupsubtwonames.cn",
			Field: "groupsubtwonames.cn",
			Type:  requestfilter.FieldTypeString,
		},
		// Brand filters
		{
			Param: "brandguid",
			Field: "brandguid",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "brand_code",
			Field: "brand_code",
			Type:  requestfilter.FieldTypeString,
		},
		// Brand Names filters
		{
			Param: "brandnames.th",
			Field: "brandnames.th",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "brandnames.en",
			Field: "brandnames.en",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "brandnames.cn",
			Field: "brandnames.cn",
			Type:  requestfilter.FieldTypeString,
		},
		// Design filters
		{
			Param: "designguid",
			Field: "designguid",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "designcode",
			Field: "designcode",
			Type:  requestfilter.FieldTypeString,
		},
		// Design Names filters
		{
			Param: "designnames.th",
			Field: "designnames.th",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "designnames.en",
			Field: "designnames.en",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "designnames.cn",
			Field: "designnames.cn",
			Type:  requestfilter.FieldTypeString,
		},
		// Model filters
		{
			Param: "modelguid",
			Field: "modelguid",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "modelcode",
			Field: "modelcode",
			Type:  requestfilter.FieldTypeString,
		},
		// Model Names filters
		{
			Param: "modelnames.th",
			Field: "modelnames.th",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "modelnames.en",
			Field: "modelnames.en",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "modelnames.cn",
			Field: "modelnames.cn",
			Type:  requestfilter.FieldTypeString,
		},
		// Pattern filters
		{
			Param: "patternguid",
			Field: "patternguid",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "patterncode",
			Field: "patterncode",
			Type:  requestfilter.FieldTypeString,
		},
		// Pattern Names filters
		{
			Param: "patternnames.th",
			Field: "patternnames.th",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "patternnames.en",
			Field: "patternnames.en",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "patternnames.cn",
			Field: "patternnames.cn",
			Type:  requestfilter.FieldTypeString,
		},
		// Grade filters
		{
			Param: "gradeguid",
			Field: "gradeguid",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "gradecode",
			Field: "gradecode",
			Type:  requestfilter.FieldTypeString,
		},
		// Grade Names filters
		{
			Param: "gradenames.th",
			Field: "gradenames.th",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "gradenames.en",
			Field: "gradenames.en",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "gradenames.cn",
			Field: "gradenames.cn",
			Type:  requestfilter.FieldTypeString,
		},
		// Category filters
		{
			Param: "category_guid",
			Field: "category_guid",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "categorycode",
			Field: "categorycode",
			Type:  requestfilter.FieldTypeString,
		},
		// Category Names filters
		{
			Param: "categorynames.th",
			Field: "categorynames.th",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "categorynames.en",
			Field: "categorynames.en",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "categorynames.cn",
			Field: "categorynames.cn",
			Type:  requestfilter.FieldTypeString,
		},
		// Class filters
		{
			Param: "classguid",
			Field: "classguid",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "classcode",
			Field: "classcode",
			Type:  requestfilter.FieldTypeString,
		},
		// Class Names filters
		{
			Param: "classnames.th",
			Field: "classnames.th",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "classnames.en",
			Field: "classnames.en",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "classnames.cn",
			Field: "classnames.cn",
			Type:  requestfilter.FieldTypeString,
		},
		// Product Names filters
		{
			Param: "names.th",
			Field: "names.th",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "names.en",
			Field: "names.en",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "names.cn",
			Field: "names.cn",
			Type:  requestfilter.FieldTypeString,
		},
		// Item Unit Names filters
		{
			Param: "itemunitnames.th",
			Field: "itemunitnames.th",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "itemunitnames.en",
			Field: "itemunitnames.en",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "itemunitnames.cn",
			Field: "itemunitnames.cn",
			Type:  requestfilter.FieldTypeString,
		},
		// Reference Unit Names filters
		{
			Param: "refunitnames.th",
			Field: "refunitnames.th",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "refunitnames.en",
			Field: "refunitnames.en",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "refunitnames.cn",
			Field: "refunitnames.cn",
			Type:  requestfilter.FieldTypeString,
		},
		// Manufacturer Names filters
		{
			Param: "manufacturernames.th",
			Field: "manufacturernames.th",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "manufacturernames.en",
			Field: "manufacturernames.en",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "manufacturernames.cn",
			Field: "manufacturernames.cn",
			Type:  requestfilter.FieldTypeString,
		},
	})

	// Handle general search parameter 'q' for codes and names
	qParam := queryParam("q")
	if qParam != "" {
		// Create comprehensive OR search for codes and names
		filters["$or"] = []bson.M{
			// Product codes
			{"barcode": bson.M{"$regex": qParam, "$options": "i"}},
			{"itemcode": bson.M{"$regex": qParam, "$options": "i"}},

			// Group codes and names
			{"group_code": bson.M{"$regex": qParam, "$options": "i"}},
			{"groupnames.name": bson.M{"$regex": qParam, "$options": "i"}},
			{"groupsubonecode": bson.M{"$regex": qParam, "$options": "i"}},
			{"groupsubonenames.name": bson.M{"$regex": qParam, "$options": "i"}},
			{"groupsubtwocode": bson.M{"$regex": qParam, "$options": "i"}},
			{"groupsubtwonames.name": bson.M{"$regex": qParam, "$options": "i"}},

			// Brand codes and names
			{"brand_code": bson.M{"$regex": qParam, "$options": "i"}},
			{"brandnames.name": bson.M{"$regex": qParam, "$options": "i"}},

			// Design codes and names
			{"designcode": bson.M{"$regex": qParam, "$options": "i"}},
			{"designnames.name": bson.M{"$regex": qParam, "$options": "i"}},

			// Model codes and names
			{"modelcode": bson.M{"$regex": qParam, "$options": "i"}},
			{"modelnames.name": bson.M{"$regex": qParam, "$options": "i"}},

			// Pattern codes and names
			{"patterncode": bson.M{"$regex": qParam, "$options": "i"}},
			{"patternnames.name": bson.M{"$regex": qParam, "$options": "i"}},

			// Grade codes and names
			{"gradecode": bson.M{"$regex": qParam, "$options": "i"}},
			{"gradenames.name": bson.M{"$regex": qParam, "$options": "i"}},

			// Category codes and names
			{"categorycode": bson.M{"$regex": qParam, "$options": "i"}},
			{"categorynames.name": bson.M{"$regex": qParam, "$options": "i"}},

			// Class codes and names
			{"classcode": bson.M{"$regex": qParam, "$options": "i"}},
			{"classnames.name": bson.M{"$regex": qParam, "$options": "i"}},

			// Product names
			{"names.name": bson.M{"$regex": qParam, "$options": "i"}},

			// Unit codes and names
			{"item_unit_code": bson.M{"$regex": qParam, "$options": "i"}},
			{"itemunitnames.name": bson.M{"$regex": qParam, "$options": "i"}},

			// Manufacturer codes and names
			{"manufacturercode": bson.M{"$regex": qParam, "$options": "i"}},
			{"manufacturernames.name": bson.M{"$regex": qParam, "$options": "i"}},
		}
	}

	if temp, ok := filters["branches.code"]; ok {
		if tempBson, ok := temp.(bson.M); ok {
			if tempIn, ok := tempBson["$in"]; ok {
				filters["ignorebranches.code"] = bson.M{
					"$nin": tempIn,
				}
			}
		} else {
			filters["ignorebranches.code"] = bson.M{
				"$ne": temp,
			}
		}
		delete(filters, "branches.code")
	}

	// Shelf filters - search products that are stored in specific shelf locations
	warehouseCodeParam := queryParam("warehousecode")
	locationCodeParam := queryParam("locationcode")
	shelfCodeParam := queryParam("shelfcode")

	if warehouseCodeParam != "" || locationCodeParam != "" || shelfCodeParam != "" {
		// If any shelf-related parameter is provided, we need to find products in those shelves
		// This will require a separate aggregation pipeline or lookup to warehouse collection
		shelfFilters := make(map[string]interface{})

		if warehouseCodeParam != "" {
			shelfFilters["warehouse.code"] = warehouseCodeParam
		}
		if locationCodeParam != "" {
			shelfFilters["location.code"] = locationCodeParam
		}
		if shelfCodeParam != "" {
			shelfFilters["shelf.code"] = shelfCodeParam
		}

		// Add shelf search criteria - this will be handled by service layer
		filters["_shelf_search"] = shelfFilters
	}

	if queryParam("isbom") != "" {
		if queryParam("isbom") == "true" {
			filters["bom"] = bson.M{
				"$exists": true,
				"$ne":     []string{},
			}
		} else {
			filters["bom"] = bson.M{
				"$eq": []string{},
			}
		}
	}

	// Filter for zero price (standard price keynumber=1 and price=0)
	zeroPriceParam := queryParam("zeroprice")
	if zeroPriceParam != "" {
		if zeroPriceParam == "true" {
			filters["prices"] = bson.M{
				"$elemMatch": bson.M{
					"key_number": 1,
					"price":     0,
				},
			}
		} else if zeroPriceParam == "false" {
			// Filter for products that do NOT have zero standard price
			filters["$or"] = []bson.M{
				{
					"prices": bson.M{
						"$not": bson.M{
							"$elemMatch": bson.M{
								"key_number": 1,
								"price":     0,
							},
						},
					},
				},
				{
					"prices": bson.M{
						"$exists": false,
					},
				},
			}
		}

	}

	return h.postProcessBooleanFilters(filters)
}

// Create ProductBarcode import godoc
// @Description Create ProductBarcode
// @Tags		ProductBarcode
// @Param		ProductBarcode  body      []models.ProductBarcode  true  "ProductBarcode"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/barcode/import [post]
func (h ProductBarcodeHttp) Import(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.ProductBarcode{}
	err := json.Unmarshal([]byte(input), &dataReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	bulkResponse, err := h.svc.Import(shopID, authUsername, dataReq)

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

// Import RefBarcode Update godoc
// @Description Import JSON to update refbarcode with standvalue and dividevalue
// @Tags		ProductBarcode
// @Param		RefBarcodeImport  body      []models.RefBarcodeImportRequest  true  "RefBarcode Import Data"
// @Accept 		json
// @Success		200	{object}	models.RefBarcodeImportResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/barcode/import-refbarcode [post]
func (h ProductBarcodeHttp) ImportRefBarcodeUpdate(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.RefBarcodeImportRequest{}
	err := json.Unmarshal([]byte(input), &dataReq)
	if err != nil {
		ctx.ResponseError(400, "Invalid JSON format: "+err.Error())
		return err
	}

	// Validate input
	if len(dataReq) == 0 {
		ctx.ResponseError(400, "No data provided")
		return fmt.Errorf("no data provided")
	}

	if len(dataReq) > 5000 {
		ctx.ResponseError(400, "Too many records. Maximum 5000 records allowed")
		return fmt.Errorf("too many records")
	}

	// Validate each record
	for i, req := range dataReq {
		if err := ctx.Validate(&req); err != nil {
			ctx.ResponseError(400, fmt.Sprintf("Validation error at row %d: %s", i+1, err.Error()))
			return err
		}

		if req.Barcode == req.BarcodeRef {
			ctx.ResponseError(400, fmt.Sprintf("Row %d: Barcode and BarcodeRef cannot be the same", i+1))
			return fmt.Errorf("invalid data at row %d", i+1)
		}

		if req.DivideValue <= 0 {
			ctx.ResponseError(400, fmt.Sprintf("Row %d: DivideValue must be greater than 0", i+1))
			return fmt.Errorf("invalid dividevalue at row %d", i+1)
		}
	}

	fmt.Printf("Starting refbarcode import for %d records\n", len(dataReq))

	response, err := h.svc.ImportRefBarcodeUpdate(shopID, authUsername, dataReq)
	if err != nil {
		ctx.ResponseError(http.StatusInternalServerError, "Import failed: "+err.Error())
		return err
	}

	statusCode := http.StatusOK
	if !response.Success {
		statusCode = http.StatusPartialContent
	}

	ctx.Response(statusCode, response)
	return nil
}
