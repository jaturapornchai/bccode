package saleinvoicereturn

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	custRepositories "smlcloudplatform/internal/debtaccount/debtor/repositories"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	trancache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/transaction/saleinvoicereturn/models"
	"smlcloudplatform/internal/transaction/saleinvoicereturn/repositories"
	"smlcloudplatform/internal/transaction/saleinvoicereturn/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

type ISaleInvoiceReturnHttp interface{}

type SaleInvoiceReturnHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.ISaleInvoiceReturnService
}

func NewSaleInvoiceReturnHttp(ms *microservice.Microservice, cfg config.IConfig) SaleInvoiceReturnHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	producer := ms.Producer(cfg.MQConfig())

	repo := repositories.NewSaleInvoiceReturnRepository(pst)
	repoMq := repositories.NewSaleInvoiceReturnMessageQueueRepository(producer)

	productBarcodeRepo := productbarcode_repositories.NewProductBarcodeRepository(pst, cache)

	transCacheRepo := trancache.NewCacheRepository(cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)

	// Add point transaction repository and customer repository
	pointTransactionRepo := custRepositories.NewPointTransactionRepository(pst)
	custRepo := custRepositories.NewDebtorRepository(pst)

	svc := services.NewSaleInvoiceReturnService(repo, transCacheRepo, productBarcodeRepo, repoMq, masterSyncCacheRepo, services.SaleInvocieReturnParser{}, pointTransactionRepo, custRepo)

	return SaleInvoiceReturnHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h SaleInvoiceReturnHttp) RegisterHttp() {

	h.ms.POST("/transaction/sale-invoice-return/bulk", h.SaveBulk)

	h.ms.GET("/transaction/sale-invoice-return", h.SearchSaleInvoiceReturnPage)
	h.ms.GET("/transaction/sale-invoice-return/list", h.SearchSaleInvoiceReturnStep)
	h.ms.POST("/transaction/sale-invoice-return", h.CreateSaleInvoiceReturn)
	h.ms.GET("/transaction/sale-invoice-return/:id", h.InfoSaleInvoiceReturn)
	h.ms.GET("/transaction/sale-invoice-return/code/:code", h.InfoSaleInvoiceReturnByCode)
	h.ms.GET("/transaction/sale-invoice-return/last-pos-docno", h.GetLastPOSDocNo)
	h.ms.PUT("/transaction/sale-invoice-return/:id", h.UpdateSaleInvoiceReturn)
	h.ms.DELETE("/transaction/sale-invoice-return/:id", h.DeleteSaleInvoiceReturn)
	h.ms.DELETE("/transaction/sale-invoice-return", h.DeleteSaleInvoiceReturnByGUIDs)
}

// Create SaleInvoiceReturn godoc
// @Description Create SaleInvoiceReturn
// @Tags		SaleInvoiceReturn
// @Param		SaleInvoiceReturn  body      models.SaleInvoiceReturn  true  "SaleInvoiceReturn"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/sale-invoice-return [post]
func (h SaleInvoiceReturnHttp) CreateSaleInvoiceReturn(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	docReq := &models.SaleInvoiceReturn{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, docNo, err := h.svc.CreateSaleInvoiceReturn(holdingCode, authUsername, *docReq)

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

// Update SaleInvoiceReturn godoc
// @Description Update SaleInvoiceReturn
// @Tags		SaleInvoiceReturn
// @Param		id  path      string  true  "SaleInvoiceReturn ID"
// @Param		SaleInvoiceReturn  body      models.SaleInvoiceReturn  true  "SaleInvoiceReturn"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/sale-invoice-return/{id} [put]
func (h SaleInvoiceReturnHttp) UpdateSaleInvoiceReturn(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.SaleInvoiceReturn{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateSaleInvoiceReturn(holdingCode, id, authUsername, *docReq)

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

// Delete SaleInvoiceReturn godoc
// @Description Delete SaleInvoiceReturn
// @Tags		SaleInvoiceReturn
// @Param		id  path      string  true  "SaleInvoiceReturn ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/sale-invoice-return/{id} [delete]
func (h SaleInvoiceReturnHttp) DeleteSaleInvoiceReturn(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteSaleInvoiceReturn(holdingCode, id, authUsername)

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

// Delete SaleInvoiceReturn godoc
// @Description Delete SaleInvoiceReturn
// @Tags		SaleInvoiceReturn
// @Param		SaleInvoiceReturn  body      []string  true  "SaleInvoiceReturn GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/sale-invoice-return [delete]
func (h SaleInvoiceReturnHttp) DeleteSaleInvoiceReturnByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteSaleInvoiceReturnByGUIDs(holdingCode, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get SaleInvoiceReturn godoc
// @Description get SaleInvoiceReturn info by guidfixed
// @Tags		SaleInvoiceReturn
// @Param		id  path      string  true  "SaleInvoiceReturn guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/sale-invoice-return/{id} [get]
func (h SaleInvoiceReturnHttp) InfoSaleInvoiceReturn(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get SaleInvoiceReturn %v", id)
	doc, err := h.svc.InfoSaleInvoiceReturn(holdingCode, id)

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

// Get SaleInvoiceReturn By Code godoc
// @Description get SaleInvoiceReturn info by Code
// @Tags		SaleInvoiceReturn
// @Param		code  path      string  true  "SaleInvoiceReturn Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/sale-invoice-return/code/{code} [get]
func (h SaleInvoiceReturnHttp) InfoSaleInvoiceReturnByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	code := ctx.Param("code")

	doc, err := h.svc.InfoSaleInvoiceReturnByCode(holdingCode, code)

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

// Get SaleInvoiceReturn By Code godoc
// @Description get SaleInvoiceReturn info by Code
// @Tags		SaleInvoiceReturn
// @Param		posid	query	string		false  "POS ID"
// @Param		maxdocno	query	string		false  "Max DocNo"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/sale-invoice-return/last-pos-docno [get]
func (h SaleInvoiceReturnHttp) GetLastPOSDocNo(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	posID := ctx.QueryParam("posid")
	maxDocNo := ctx.QueryParam("maxdocno")

	if posID == "" || maxDocNo == "" {
		ctx.ResponseError(http.StatusBadRequest, "posid and maxdocno is required")
		return nil
	}

	doc, err := h.svc.GetLastPOSDocNo(holdingCode, posID, maxDocNo)

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

// List SaleInvoiceReturn step godoc
// @Description get list step
// @Tags		SaleInvoiceReturn
// @Param		q		query	string		false  "Search Value"
// @Param		custcode	query	string		false  "cust code"
// @Param		branchcode	query	string		false  "branch code"
// @Param		fromdate	query	string		false  "from date"
// @Param		todate	query	string		false  "to date"
// @Param		ispos	query	boolean		false  "is POS"
// @Param		posid 	query	string		false  "POS id"
// @Param		machinecode query	string		false  "machine code"
// @Parm		zonegroupnumber query	string		false  "zone group number"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/sale-invoice-return [get]
func (h SaleInvoiceReturnHttp) SearchSaleInvoiceReturnPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageable := utils.GetPageable(ctx.QueryParam)

	filters := h.searchFilter(ctx.QueryParam)

	docList, pagination, err := h.svc.SearchSaleInvoiceReturn(holdingCode, filters, pageable)

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

// List SaleInvoiceReturn godoc
// @Description search limit offset
// @Tags		SaleInvoiceReturn
// @Param		q		query	string		false  "Search Value"
// @Param		custcode	query	string		false  "cust code"
// @Param		branchcode	query	string		false  "branch code"
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
// @Router /transaction/sale-invoice-return/list [get]
func (h SaleInvoiceReturnHttp) SearchSaleInvoiceReturnStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	filters := h.searchFilter(ctx.QueryParam)

	docList, total, err := h.svc.SearchSaleInvoiceReturnStep(holdingCode, lang, filters, pageableStep)

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

// Create SaleInvoiceReturn Bulk godoc
// @Description Create SaleInvoiceReturn
// @Tags		SaleInvoiceReturn
// @Param		SaleInvoiceReturn  body      []models.SaleInvoiceReturn  true  "SaleInvoiceReturn"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/sale-invoice-return/bulk [post]
func (h SaleInvoiceReturnHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()

	dataReq := []models.SaleInvoiceReturn{}
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

func (h SaleInvoiceReturnHttp) searchFilter(queryParam func(string) string) map[string]interface{} {
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
	})

	return filters
}
