package saleinvoice

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"smlcloudplatform/internal/config"
	couponRepositories "smlcloudplatform/internal/coupon/repositories"
	couponServices "smlcloudplatform/internal/coupon/services"
	repoCust "smlcloudplatform/internal/debtaccount/debtor/repositories"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	trans_cache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/transaction/saleinvoice/models"
	"smlcloudplatform/internal/transaction/saleinvoice/repositories"
	"smlcloudplatform/internal/transaction/saleinvoice/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

type ISaleInvoiceHttp interface{}

type SaleInvoiceHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.ISaleInvoiceService
}

func NewSaleInvoiceHttp(ms *microservice.Microservice, cfg config.IConfig) SaleInvoiceHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	producer := ms.Producer(cfg.MQConfig())

	repo := repositories.NewSaleInvoiceRepository(pst)
	repoMq := repositories.NewSaleInvoiceMessageQueueRepository(producer)
	repoCustx := repoCust.NewDebtorRepository(pst)
	pointTransactionRepo := repoCust.NewPointTransactionRepository(pst)

	productBarcodeRepo := productbarcode_repositories.NewProductBarcodeRepository(pst, cache)

	transCacheRepo := trans_cache.NewCacheRepository(cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)

	// Initialize coupon service
	couponRepo := couponRepositories.NewCouponRepository(pst)
	couponReservationRepo := couponRepositories.NewCouponReservationRepository(pst)
	couponUsageHistoryRepo := couponRepositories.NewCouponUsageHistoryRepository(pst)
	couponService := couponServices.NewCouponHttpService(couponRepo, couponReservationRepo, couponUsageHistoryRepo, productBarcodeRepo, masterSyncCacheRepo)

	svc := services.NewSaleInvoiceService(
		repo,
		repoCustx,
		transCacheRepo,
		productBarcodeRepo,
		pointTransactionRepo,
		repoMq,
		masterSyncCacheRepo,
		couponService, // Add coupon service
		services.SaleInvocieParser{},
		services.SaleInvocieExport{},
	)

	return SaleInvoiceHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h SaleInvoiceHttp) RegisterHttp() {

	h.ms.POST("/transaction/sale-invoice/bulk", h.SaveBulk)

	h.ms.GET("/transaction/sale-invoice", h.SearchSaleInvoicePage)
	h.ms.GET("/transaction/sale-invoice/list", h.SearchSaleInvoiceStep)
	h.ms.POST("/transaction/sale-invoice", h.CreateSaleInvoice)
	h.ms.POST("/transaction/sale-invoice/auto-coupon", h.CreateSaleInvoiceWithAutoCoupon) // New endpoint
	h.ms.GET("/transaction/sale-invoice/:id", h.InfoSaleInvoice)
	h.ms.GET("/transaction/sale-invoice/last-pos-docno", h.GetLastPOSDocNo)
	h.ms.GET("/transaction/sale-invoice/code/:code", h.InfoSaleInvoiceByCode)
	h.ms.GET("/transaction/sale-invoice/guidpos/:code", h.InfoSaleInvoiceByGuidPos)
	h.ms.GET("/transaction/sale-invoice/recalpoint/:code", h.RecalPoint)
	h.ms.PUT("/transaction/sale-invoice/:id", h.UpdateSaleInvoice)
	h.ms.DELETE("/transaction/sale-invoice/:id", h.DeleteSaleInvoice)
	h.ms.DELETE("/transaction/sale-invoice", h.DeleteSaleInvoiceByGUIDs)
	h.ms.GET("/transaction/sale-invoice/export", h.Export)
}

// Create SaleInvoice godoc
// @Description Create SaleInvoice
// @Tags		SaleInvoice
// @Param		SaleInvoice  body      models.SaleInvoice  true  "SaleInvoice"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/sale-invoice [post]
func (h SaleInvoiceHttp) CreateSaleInvoice(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.SaleInvoice{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, docNo, err := h.svc.CreateSaleInvoice(shopID, authUsername, *docReq)

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

// Create SaleInvoice with Auto Coupon godoc
// @Description Create SaleInvoice with automatic coupon processing
// @Tags		SaleInvoice
// @Param		SaleInvoice  body      models.SaleInvoice  true  "SaleInvoice with coupons"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/sale-invoice/auto-coupon [post]
func (h SaleInvoiceHttp) CreateSaleInvoiceWithAutoCoupon(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.SaleInvoice{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, docNo, err := h.svc.CreateSaleInvoiceWithAutoCoupon(shopID, authUsername, *docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      idx,
		Data:    docNo,
		Message: "Sale invoice created successfully. Coupons are being processed automatically in the background.",
	})
	return nil
}

// RecalPoint godoc
// @Description RecalPoint
// @Tags		SaleInvoice
// @Accept 		json
// @Param		code  path      string  true  "Cust Code"
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/sale-invoice/recalpoint/:code [get]
func (h SaleInvoiceHttp) RecalPoint(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	code := ctx.Param("code")

	if code == "" {
		ctx.ResponseError(http.StatusBadRequest, "code is required")
		return nil
	}

	err := h.svc.RecalPoint(shopID, authUsername, code)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Update SaleInvoice godoc
// @Description Update SaleInvoice
// @Tags		SaleInvoice
// @Param		id  path      string  true  "SaleInvoice ID"
// @Param		SaleInvoice  body      models.SaleInvoice  true  "SaleInvoice"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/sale-invoice/{id} [put]
func (h SaleInvoiceHttp) UpdateSaleInvoice(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.SaleInvoice{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateSaleInvoice(shopID, id, authUsername, *docReq)

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

// Delete SaleInvoice godoc
// @Description Delete SaleInvoice
// @Tags		SaleInvoice
// @Param		id  path      string  true  "SaleInvoice ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/sale-invoice/{id} [delete]
func (h SaleInvoiceHttp) DeleteSaleInvoice(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteSaleInvoice(shopID, id, authUsername)

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

// Delete SaleInvoice godoc
// @Description Delete SaleInvoice
// @Tags		SaleInvoice
// @Param		SaleInvoice  body      []string  true  "SaleInvoice GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/sale-invoice [delete]
func (h SaleInvoiceHttp) DeleteSaleInvoiceByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteSaleInvoiceByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get SaleInvoice godoc
// @Description get SaleInvoice info by guidfixed
// @Tags		SaleInvoice
// @Param		id  path      string  true  "SaleInvoice guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/sale-invoice/{id} [get]
func (h SaleInvoiceHttp) InfoSaleInvoice(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get SaleInvoice %v", id)
	doc, err := h.svc.InfoSaleInvoice(shopID, id)

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

// Get SaleInvoice By Code godoc
// @Description get SaleInvoice info by Code
// @Tags		SaleInvoice
// @Param		code  path      string  true  "SaleInvoice Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/sale-invoice/code/{code} [get]
func (h SaleInvoiceHttp) InfoSaleInvoiceByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")

	doc, err := h.svc.InfoSaleInvoiceByCode(shopID, code)

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

// Get SaleInvoice By GuidPos godoc
// @Description get SaleInvoice info by GuidPos
// @Tags		SaleInvoice
// @Param		code  path      string  true  "SaleInvoice GuidPos"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/sale-invoice/guidpos/{code} [get]
func (h SaleInvoiceHttp) InfoSaleInvoiceByGuidPos(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")

	doc, err := h.svc.InfoSaleInvoiceByGuidPos(shopID, code)

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

// Get SaleInvoice By Code godoc
// @Description get SaleInvoice info by Code
// @Tags		SaleInvoice
// @Param		posid	query	string		false  "POS ID"
// @Param		maxdocno	query	string		false  "Max DocNo"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/sale-invoice/last-pos-docno [get]
func (h SaleInvoiceHttp) GetLastPOSDocNo(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	posID := ctx.QueryParam("posid")
	maxDocNo := ctx.QueryParam("maxdocno")

	if posID == "" || maxDocNo == "" {
		ctx.ResponseError(http.StatusBadRequest, "posid and maxdocno is required")
		return nil
	}

	doc, err := h.svc.GetLastPOSDocNo(shopID, posID, maxDocNo)

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

// List SaleInvoice step godoc
// @Description get list step
// @Tags		SaleInvoice
// @Param		q		query	string		false  "Search Value"
// @Param		custcode	query	string		false  "cust code"
// @Param		branchcode	query	string		false  "branch code"
// @Param		fromdate	query	string		false  "from date"
// @Param		todate	query	string		false  "to date"
// @Param		ispos	query	boolean		false  "is POS"
// @Param		posid 	query	string		false  "POS id"
// @Param		machinecode query	string		false  "machine code"
// @Param		createdatetimeafter query	string		false  "Create DateTime After" Format(YYYY-MM-DD HH:mm:ss)
// @Param		updatedatetimeafter query	string		false  "Update DateTime After" Format(YYYY-MM-DD HH:mm:ss)
// @Param		createupdateafterdatetime	query	string		false  "Create/Update DateTime After" Format(YYYY-MM-DD HH:mm:ss)
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/sale-invoice [get]
func (h SaleInvoiceHttp) SearchSaleInvoicePage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	filters := h.searchFilter(ctx.QueryParam)

	docList, pagination, err := h.svc.SearchSaleInvoice(shopID, filters, pageable)

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

// List SaleInvoice godoc
// @Description search limit offset
// @Tags		SaleInvoice
// @Param		q		query	string		false  "Search Value"
// @Param		fromdate	query	string		false  "from date"
// @Param		todate	query	string		false  "to date"
// @Param		ispos	query	boolean		false  "is POS"
// @Param		posid 	query	string		false  "POS id"
// @Param		machinecode query	string		false  "machine code"
// @Parm		zonegroupnumber query	string		false  "zone group number"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/sale-invoice/list [get]
func (h SaleInvoiceHttp) SearchSaleInvoiceStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	filters := h.searchFilter(ctx.QueryParam)

	docList, total, err := h.svc.SearchSaleInvoiceStep(shopID, lang, filters, pageableStep)

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

// Create SaleInvoice Bulk godoc
// @Description Create SaleInvoice
// @Tags		SaleInvoice
// @Param		SaleInvoice  body      []models.SaleInvoice  true  "SaleInvoice"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/sale-invoice/bulk [post]
func (h SaleInvoiceHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.SaleInvoice{}
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

// Get  Export
// @Description SaleInvoice Export
// @Tags		SaleInvoice
// @Param		lang	query	string		false  "language code"
// @Param		docdate	query	string		false  "Label Doc Date"
// @Param		docno	query	string		false  "Label Doc No"
// @Param		barcode	query	string		false  "Label Barcode"
// @Param		productname	query	string		false  "Label Product Name"
// @Param		unitcode	query	string		false  "Label Unit Code"
// @Param		unitname	query	string		false  "Label Unit Name"
// @Param		qty	query	string		false  "Label Qty"
// @Param		price	query	string		false  "Label Price"
// @Param		discountamount	query	string		false  "Label Discount Amount"
// @Param		sumamount	query	string		false  "Label Sum Amount"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/sale-invoice/export [get]
func (h SaleInvoiceHttp) Export(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	languageCode := ctx.QueryParam("lang")

	if languageCode == "" {
		languageCode = "en"
	}

	keyCols := []string{
		"docdate",        //"วันที่",
		"docno",          //เลขที่เอกสาร",
		"barcode",        //บาร์โค้ด",
		"productname",    //"ชื่อสินค้า",
		"unitcode",       //"หน่วยนับ",
		"unit_name",       //"ชื่อหน่วยนับ",
		"qty",            //"จำนวน",
		"price",          //ราคา",
		"discountamount", // "มูลค่าส่วนลด",
		"sum_amount",      //"มูลค่าสินค้า",
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

	fileName := fmt.Sprintf("%s_sale_invoice_%s.csv", shopID, time.Now().Format("20060102150405"))

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

func (h SaleInvoiceHttp) searchFilter(queryParam func(string) string) map[string]interface{} {
	filters := requestfilter.GenerateFilters(queryParam, []requestfilter.FilterRequest{
		{
			Param: "custcode",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "ispos",
			Field: "ispos",
			Type:  requestfilter.FieldTypeBoolean,
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
		{
			Param: "posid",
			Field: "posid",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "machinecode",
			Field: "machinecode",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "zonegroupnumber",
			Field: "zonegroupnumber",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "createdatetimeafter",
			Field: "createdatetimeafter",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "updateafterdatetime",
			Field: "updateafterdatetime",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "createupdateafterdatetime",
			Field: "createupdateafterdatetime",
			Type:  requestfilter.FieldTypeString,
		},
	})

	return filters
}
