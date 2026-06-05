package purchaseorder

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	currencyrepo "smlcloudplatform/internal/currency/repositories"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/purchaseorder/models"
	"smlcloudplatform/internal/transaction/purchaseorder/repositories"
	"smlcloudplatform/internal/transaction/purchaseorder/services"
	"smlcloudplatform/internal/transaction/purchaseorder/validators"
	trancache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
	"strings"
)

type IPurchaseOrderHttp interface{}

type PurchaseOrderHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IPurchaseOrderHttpService
}

func NewPurchaseOrderHttp(ms *microservice.Microservice, cfg config.IConfig) PurchaseOrderHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	producer := ms.Producer(cfg.MQConfig())

	repo := repositories.NewPurchaseOrderRepository(pst)
	repoMq := repositories.NewPurchaseOrderMessageQueueRepository(producer)
	currencyRepo := currencyrepo.NewCurrencyRepository(pst)

	transRepo := trancache.NewCacheRepository(cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewPurchaseOrderHttpService(repo, transRepo, repoMq, currencyRepo, masterSyncCacheRepo)

	return PurchaseOrderHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

// getRequestLanguage — อ่านภาษาจาก request (Accept-Language header → query param → default "en")
// Flutter ส่ง Accept-Language header มาอัตโนมัติ ตามภาษาที่ user เลือกใช้บนหน้าจอ
// AI Agent/External client สามารถส่งผ่าน query param ?lang=th ได้เช่นกัน
func getRequestLanguage(ctx microservice.IContext) string {
	// ลำดับความสำคัญ: query param > Accept-Language header > default "en"
	if lang := ctx.QueryParam("lang"); lang != "" {
		return lang
	}
	if lang := ctx.Header("Accept-Language"); lang != "" {
		return lang
	}
	return "en"
}

// sanitizeEmptyTimeFields — แก้ empty string ใน time fields ให้เป็น Go zero time
// ป้องกัน json.Unmarshal error เมื่อ Flutter/external client ส่ง "" สำหรับ time.Time fields
// เช่น createdat: "" → createdat: "0001-01-01T00:00:00Z"
func sanitizeEmptyTimeFields(input string) string {
	// time fields ที่เป็น time.Time ใน Go struct แต่ Flutter อาจส่ง ""
	timeFields := []string{
		"docdatetime", "docrefdate", "taxdocdate",
		"createdat", "modified_at",
		"docrefdatetime",
	}
	result := input
	for _, field := range timeFields {
		// แทนที่ "fieldname":"" → "fieldname":"0001-01-01T00:00:00Z"
		result = strings.ReplaceAll(result,
			`"`+field+`":""`,
			`"`+field+`":"0001-01-01T00:00:00Z"`)
	}
	return result
}

func (h PurchaseOrderHttp) RegisterHttp() {

	h.ms.POST("/transaction/purchase-order/bulk", h.SaveBulk)

	h.ms.GET("/transaction/purchase-order", h.SearchPurchaseOrderPage)
	h.ms.GET("/transaction/purchase-order/list", h.SearchPurchaseOrderStep)
	h.ms.POST("/transaction/purchase-order", h.CreatePurchaseOrder)
	h.ms.GET("/transaction/purchase-order/:id", h.InfoPurchaseOrder)
	h.ms.GET("/transaction/purchase-order/code/:code", h.InfoPurchaseOrderByCode)
	h.ms.PUT("/transaction/purchase-order/:id", h.UpdatePurchaseOrder)
	h.ms.DELETE("/transaction/purchase-order/:id", h.DeletePurchaseOrder)
	h.ms.DELETE("/transaction/purchase-order", h.DeletePurchaseOrderByGUIDs)
}

// Create PurchaseOrder godoc
// @Description Create PurchaseOrder
// @Tags		PurchaseOrder
// @Param		PurchaseOrder  body      models.PurchaseOrder  true  "PurchaseOrder"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/purchase-order [post]
func (h PurchaseOrderHttp) CreatePurchaseOrder(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	// ภาษาสำหรับ response message — อ่านจาก Accept-Language header (ที่ frontend ส่งมา)
	// fallback เป็น query param ?lang=xx, default = English
	lang := getRequestLanguage(ctx)

	docReq := &models.PurchaseOrder{}
	// sanitize empty time fields ก่อน unmarshal — ป้องกัน error จาก Flutter ที่ส่ง createdat: "" หรือ modified_at: ""
	err := json.Unmarshal([]byte(sanitizeEmptyTimeFields(input)), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, docNo, validationResult, err := h.svc.CreatePurchaseOrder(holdingCode, authUsername, *docReq)

	// ตรวจสอบ validation errors — return รายละเอียดทุก field ที่ผิดพลาด
	if validationResult != nil && !validationResult.IsValid() {
		ctx.Response(http.StatusBadRequest, validationResult.ToErrorResponse(lang))
		return nil
	}

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, validators.NewCreateSuccessResponse(idx, docNo, lang))
	return nil
}

// Update PurchaseOrder godoc
// @Description Update PurchaseOrder
// @Tags		PurchaseOrder
// @Param		id  path      string  true  "PurchaseOrder ID"
// @Param		PurchaseOrder  body      models.PurchaseOrder  true  "PurchaseOrder"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/purchase-order/{id} [put]
func (h PurchaseOrderHttp) UpdatePurchaseOrder(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")
	input := ctx.ReadInput()

	// ภาษาสำหรับ response message — อ่านจาก Accept-Language header หรือ query param
	lang := getRequestLanguage(ctx)

	docReq := &models.PurchaseOrder{}
	// sanitize empty time fields ก่อน unmarshal — ป้องกัน error จาก Flutter ที่ส่ง createdat: "" หรือ modified_at: ""
	err := json.Unmarshal([]byte(sanitizeEmptyTimeFields(input)), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	validationResult, err := h.svc.UpdatePurchaseOrder(holdingCode, id, authUsername, *docReq)

	// ตรวจสอบ validation errors — return รายละเอียดทุก field ที่ผิดพลาด
	if validationResult != nil && !validationResult.IsValid() {
		ctx.Response(http.StatusBadRequest, validationResult.ToUpdateErrorResponse(lang))
		return nil
	}

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, validators.NewUpdateSuccessResponse(id, lang))

	return nil
}

// Delete PurchaseOrder godoc
// @Description Delete PurchaseOrder
// @Tags		PurchaseOrder
// @Param		id  path      string  true  "PurchaseOrder ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/purchase-order/{id} [delete]
func (h PurchaseOrderHttp) DeletePurchaseOrder(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeletePurchaseOrder(holdingCode, id, authUsername)

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

// Delete PurchaseOrder godoc
// @Description Delete PurchaseOrder
// @Tags		PurchaseOrder
// @Param		PurchaseOrder  body      []string  true  "PurchaseOrder GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/purchase-order [delete]
func (h PurchaseOrderHttp) DeletePurchaseOrderByGUIDs(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	input := ctx.ReadInput()

	docReq := []string{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.DeletePurchaseOrderByGUIDs(holdingCode, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get PurchaseOrder godoc
// @Description get PurchaseOrder info by guidfixed
// @Tags		PurchaseOrder
// @Param		id  path      string  true  "PurchaseOrder guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/purchase-order/{id} [get]
func (h PurchaseOrderHttp) InfoPurchaseOrder(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get PurchaseOrder %v", id)
	doc, err := h.svc.InfoPurchaseOrder(holdingCode, id)

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

// Get PurchaseOrder By Code godoc
// @Description get PurchaseOrder info by Code
// @Tags		PurchaseOrder
// @Param		code  path      string  true  "PurchaseOrder Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/purchase-order/code/{code} [get]
func (h PurchaseOrderHttp) InfoPurchaseOrderByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	code := ctx.Param("code")

	doc, err := h.svc.InfoPurchaseOrderByCode(holdingCode, code)

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

// List PurchaseOrder step godoc
// @Description get list step
// @Tags		PurchaseOrder
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
// @Router /transaction/purchase-order [get]
func (h PurchaseOrderHttp) SearchPurchaseOrderPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

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

	docList, pagination, err := h.svc.SearchPurchaseOrder(holdingCode, filters, pageable)

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

// List PurchaseOrder godoc
// @Description search limit offset
// @Tags		PurchaseOrder
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
// @Router /transaction/purchase-order/list [get]
func (h PurchaseOrderHttp) SearchPurchaseOrderStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

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

	docList, total, err := h.svc.SearchPurchaseOrderStep(holdingCode, lang, filters, pageableStep)

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

// Create PurchaseOrder Bulk godoc
// @Description Create PurchaseOrder
// @Tags		PurchaseOrder
// @Param		PurchaseOrder  body      []models.PurchaseOrder  true  "PurchaseOrder"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/purchase-order/bulk [post]
func (h PurchaseOrderHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()

	dataReq := []models.PurchaseOrder{}
	err := json.Unmarshal([]byte(input), &dataReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	bulkResponse, err := h.svc.SaveInBatch(holdingCode, authUsername, dataReq)

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
