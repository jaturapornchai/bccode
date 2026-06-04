package saledebitnote

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	trancache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/transaction/saledebitnote/models"
	"smlcloudplatform/internal/transaction/saledebitnote/repositories"
	"smlcloudplatform/internal/transaction/saledebitnote/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

type ISaleDebitNoteHttp interface{}

type SaleDebitNoteHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.ISaleDebitNoteHttpService
}

func NewSaleDebitNoteHttp(ms *microservice.Microservice, cfg config.IConfig) SaleDebitNoteHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	producer := ms.Producer(cfg.MQConfig())

	repo := repositories.NewSaleDebitNoteRepository(pst)
	repoMq := repositories.NewSaleDebitNoteMessageQueueRepository(producer)

	transRepo := trancache.NewCacheRepository(cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewSaleDebitNoteHttpService(repo, transRepo, repoMq, masterSyncCacheRepo)

	return SaleDebitNoteHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h SaleDebitNoteHttp) RegisterHttp() {

	h.ms.POST("/transaction/bank/saledebitnote/bulk", h.SaveBulk)

	h.ms.GET("/transaction/bank/saledebitnote", h.SearchSaleDebitNotePage)
	h.ms.GET("/transaction/bank/saledebitnote/list", h.SearchSaleDebitNoteStep)
	h.ms.POST("/transaction/bank/saledebitnote", h.CreateSaleDebitNote)
	h.ms.GET("/transaction/bank/saledebitnote/:id", h.InfoSaleDebitNote)
	h.ms.GET("/transaction/bank/saledebitnote/code/:code", h.InfoSaleDebitNoteByCode)
	h.ms.PUT("/transaction/bank/saledebitnote/:id", h.UpdateSaleDebitNote)
	h.ms.DELETE("/transaction/bank/saledebitnote/:id", h.DeleteSaleDebitNote)
	h.ms.DELETE("/transaction/bank/saledebitnote", h.DeleteSaleDebitNoteByGUIDs)
}

// Create SaleDebitNote godoc
// @Description Create SaleDebitNote
// @Tags		SaleDebitNote
// @Param		SaleDebitNote  body      models.SaleDebitNote  true  "SaleDebitNote"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/saledebitnote [post]
func (h SaleDebitNoteHttp) CreateSaleDebitNote(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	docReq := &models.SaleDebitNote{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, docNo, err := h.svc.CreateSaleDebitNote(holdingCode, authUsername, *docReq)

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

// Update SaleDebitNote godoc
// @Description Update SaleDebitNote
// @Tags		SaleDebitNote
// @Param		id  path      string  true  "SaleDebitNote ID"
// @Param		SaleDebitNote  body      models.SaleDebitNote  true  "SaleDebitNote"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/saledebitnote/{id} [put]
func (h SaleDebitNoteHttp) UpdateSaleDebitNote(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.SaleDebitNote{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateSaleDebitNote(holdingCode, id, authUsername, *docReq)

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

// Delete SaleDebitNote godoc
// @Description Delete SaleDebitNote
// @Tags		SaleDebitNote
// @Param		id  path      string  true  "SaleDebitNote ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/saledebitnote/{id} [delete]
func (h SaleDebitNoteHttp) DeleteSaleDebitNote(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteSaleDebitNote(holdingCode, id, authUsername)

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

// Delete SaleDebitNote godoc
// @Description Delete SaleDebitNote
// @Tags		SaleDebitNote
// @Param		SaleDebitNote  body      []string  true  "SaleDebitNote GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/saledebitnote [delete]
func (h SaleDebitNoteHttp) DeleteSaleDebitNoteByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteSaleDebitNoteByGUIDs(holdingCode, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get SaleDebitNote godoc
// @Description get SaleDebitNote info by guidfixed
// @Tags		SaleDebitNote
// @Param		id  path      string  true  "SaleDebitNote guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/saledebitnote/{id} [get]
func (h SaleDebitNoteHttp) InfoSaleDebitNote(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get SaleDebitNote %v", id)
	doc, err := h.svc.InfoSaleDebitNote(holdingCode, id)

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

// Get SaleDebitNote By Code godoc
// @Description get SaleDebitNote info by Code
// @Tags		SaleDebitNote
// @Param		code  path      string  true  "SaleDebitNote Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/saledebitnote/code/{code} [get]
func (h SaleDebitNoteHttp) InfoSaleDebitNoteByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	code := ctx.Param("code")

	doc, err := h.svc.InfoSaleDebitNoteByCode(holdingCode, code)

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

// List SaleDebitNote step godoc
// @Description get list step
// @Tags		SaleDebitNote
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
// @Router /transaction/bank/saledebitnote [get]
func (h SaleDebitNoteHttp) SearchSaleDebitNotePage(ctx microservice.IContext) error {
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

	docList, pagination, err := h.svc.SearchSaleDebitNote(holdingCode, filters, pageable)

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

// List SaleDebitNote godoc
// @Description search limit offset
// @Tags		SaleDebitNote
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
// @Router /transaction/bank/saledebitnote/list [get]
func (h SaleDebitNoteHttp) SearchSaleDebitNoteStep(ctx microservice.IContext) error {
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

	docList, total, err := h.svc.SearchSaleDebitNoteStep(holdingCode, lang, filters, pageableStep)

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

// Create SaleDebitNote Bulk godoc
// @Description Create SaleDebitNote
// @Tags		SaleDebitNote
// @Param		SaleDebitNote  body      []models.SaleDebitNote  true  "SaleDebitNote"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/saledebitnote/bulk [post]
func (h SaleDebitNoteHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()

	dataReq := []models.SaleDebitNote{}
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
