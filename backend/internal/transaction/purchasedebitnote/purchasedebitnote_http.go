package purchasedebitnote

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/purchasedebitnote/models"
	"smlcloudplatform/internal/transaction/purchasedebitnote/repositories"
	"smlcloudplatform/internal/transaction/purchasedebitnote/services"
	trancache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

type IPurchaseDebitNoteHttp interface{}

type PurchaseDebitNoteHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IPurchaseDebitNoteHttpService
}

func NewPurchaseDebitNoteHttp(ms *microservice.Microservice, cfg config.IConfig) PurchaseDebitNoteHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	producer := ms.Producer(cfg.MQConfig())

	repo := repositories.NewPurchaseDebitNoteRepository(pst)
	repoMq := repositories.NewPurchaseDebitNoteMessageQueueRepository(producer)

	transRepo := trancache.NewCacheRepository(cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewPurchaseDebitNoteHttpService(repo, transRepo, repoMq, masterSyncCacheRepo)

	return PurchaseDebitNoteHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h PurchaseDebitNoteHttp) RegisterHttp() {

	h.ms.POST("/transaction/bank/purchasedebitnote/bulk", h.SaveBulk)

	h.ms.GET("/transaction/bank/purchasedebitnote", h.SearchPurchaseDebitNotePage)
	h.ms.GET("/transaction/bank/purchasedebitnote/list", h.SearchPurchaseDebitNoteStep)
	h.ms.POST("/transaction/bank/purchasedebitnote", h.CreatePurchaseDebitNote)
	h.ms.GET("/transaction/bank/purchasedebitnote/:id", h.InfoPurchaseDebitNote)
	h.ms.GET("/transaction/bank/purchasedebitnote/code/:code", h.InfoPurchaseDebitNoteByCode)
	h.ms.PUT("/transaction/bank/purchasedebitnote/:id", h.UpdatePurchaseDebitNote)
	h.ms.DELETE("/transaction/bank/purchasedebitnote/:id", h.DeletePurchaseDebitNote)
	h.ms.DELETE("/transaction/bank/purchasedebitnote", h.DeletePurchaseDebitNoteByGUIDs)
}

// Create PurchaseDebitNote godoc
// @Description Create PurchaseDebitNote
// @Tags		PurchaseDebitNote
// @Param		PurchaseDebitNote  body      models.PurchaseDebitNote  true  "PurchaseDebitNote"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/purchasedebitnote [post]
func (h PurchaseDebitNoteHttp) CreatePurchaseDebitNote(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	docReq := &models.PurchaseDebitNote{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, docNo, err := h.svc.CreatePurchaseDebitNote(holdingCode, authUsername, *docReq)

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

// Update PurchaseDebitNote godoc
// @Description Update PurchaseDebitNote
// @Tags		PurchaseDebitNote
// @Param		id  path      string  true  "PurchaseDebitNote ID"
// @Param		PurchaseDebitNote  body      models.PurchaseDebitNote  true  "PurchaseDebitNote"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/purchasedebitnote/{id} [put]
func (h PurchaseDebitNoteHttp) UpdatePurchaseDebitNote(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.PurchaseDebitNote{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdatePurchaseDebitNote(holdingCode, id, authUsername, *docReq)

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

// Delete PurchaseDebitNote godoc
// @Description Delete PurchaseDebitNote
// @Tags		PurchaseDebitNote
// @Param		id  path      string  true  "PurchaseDebitNote ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/purchasedebitnote/{id} [delete]
func (h PurchaseDebitNoteHttp) DeletePurchaseDebitNote(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeletePurchaseDebitNote(holdingCode, id, authUsername)

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

// Delete PurchaseDebitNote godoc
// @Description Delete PurchaseDebitNote
// @Tags		PurchaseDebitNote
// @Param		PurchaseDebitNote  body      []string  true  "PurchaseDebitNote GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/purchasedebitnote [delete]
func (h PurchaseDebitNoteHttp) DeletePurchaseDebitNoteByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeletePurchaseDebitNoteByGUIDs(holdingCode, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get PurchaseDebitNote godoc
// @Description get PurchaseDebitNote info by guidfixed
// @Tags		PurchaseDebitNote
// @Param		id  path      string  true  "PurchaseDebitNote guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/purchasedebitnote/{id} [get]
func (h PurchaseDebitNoteHttp) InfoPurchaseDebitNote(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get PurchaseDebitNote %v", id)
	doc, err := h.svc.InfoPurchaseDebitNote(holdingCode, id)

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

// Get PurchaseDebitNote By Code godoc
// @Description get PurchaseDebitNote info by Code
// @Tags		PurchaseDebitNote
// @Param		code  path      string  true  "PurchaseDebitNote Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/purchasedebitnote/code/{code} [get]
func (h PurchaseDebitNoteHttp) InfoPurchaseDebitNoteByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	code := ctx.Param("code")

	doc, err := h.svc.InfoPurchaseDebitNoteByCode(holdingCode, code)

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

// List PurchaseDebitNote step godoc
// @Description get list step
// @Tags		PurchaseDebitNote
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
// @Router /transaction/bank/purchasedebitnote [get]
func (h PurchaseDebitNoteHttp) SearchPurchaseDebitNotePage(ctx microservice.IContext) error {
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

	docList, pagination, err := h.svc.SearchPurchaseDebitNote(holdingCode, filters, pageable)

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

// List PurchaseDebitNote godoc
// @Description search limit offset
// @Tags		PurchaseDebitNote
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
// @Router /transaction/bank/purchasedebitnote/list [get]
func (h PurchaseDebitNoteHttp) SearchPurchaseDebitNoteStep(ctx microservice.IContext) error {
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

	docList, total, err := h.svc.SearchPurchaseDebitNoteStep(holdingCode, lang, filters, pageableStep)

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

// Create PurchaseDebitNote Bulk godoc
// @Description Create PurchaseDebitNote
// @Tags		PurchaseDebitNote
// @Param		PurchaseDebitNote  body      []models.PurchaseDebitNote  true  "PurchaseDebitNote"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/purchasedebitnote/bulk [post]
func (h PurchaseDebitNoteHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()

	dataReq := []models.PurchaseDebitNote{}
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
