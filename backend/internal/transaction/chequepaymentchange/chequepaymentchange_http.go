package chequepaymentchange

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/chequepaymentchange/models"
	"smlcloudplatform/internal/transaction/chequepaymentchange/repositories"
	"smlcloudplatform/internal/transaction/chequepaymentchange/services"
	trancache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

type IChequePaymentChangeHttp interface{}

type ChequePaymentChangeHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IChequePaymentChangeHttpService
}

func NewChequePaymentChangeHttp(ms *microservice.Microservice, cfg config.IConfig) ChequePaymentChangeHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	producer := ms.Producer(cfg.MQConfig())

	repo := repositories.NewChequePaymentChangeRepository(pst)
	repoMq := repositories.NewChequePaymentChangeMessageQueueRepository(producer)

	transRepo := trancache.NewCacheRepository(cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewChequePaymentChangeHttpService(repo, transRepo, repoMq, masterSyncCacheRepo)

	return ChequePaymentChangeHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h ChequePaymentChangeHttp) RegisterHttp() {

	h.ms.POST("/transaction/chequepayment/chequepaymentchange/bulk", h.SaveBulk)

	h.ms.GET("/transaction/chequepayment/chequepaymentchange", h.SearchChequePaymentChangePage)
	h.ms.GET("/transaction/chequepayment/chequepaymentchange/list", h.SearchChequePaymentChangeStep)
	h.ms.POST("/transaction/chequepayment/chequepaymentchange", h.CreateChequePaymentChange)
	h.ms.GET("/transaction/chequepayment/chequepaymentchange/:id", h.InfoChequePaymentChange)
	h.ms.GET("/transaction/chequepayment/chequepaymentchange/code/:code", h.InfoChequePaymentChangeByCode)
	h.ms.PUT("/transaction/chequepayment/chequepaymentchange/:id", h.UpdateChequePaymentChange)
	h.ms.DELETE("/transaction/chequepayment/chequepaymentchange/:id", h.DeleteChequePaymentChange)
	h.ms.DELETE("/transaction/chequepayment/chequepaymentchange", h.DeleteChequePaymentChangeByGUIDs)
}

// Create ChequePaymentChange godoc
// @Description Create ChequePaymentChange
// @Tags		ChequePaymentChange
// @Param		ChequePaymentChange  body      models.ChequePaymentChange  true  "ChequePaymentChange"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequepayment/chequepaymentchange [post]
func (h ChequePaymentChangeHttp) CreateChequePaymentChange(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	docReq := &models.ChequePaymentChange{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, docNo, err := h.svc.CreateChequePaymentChange(holdingCode, authUsername, *docReq)

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

// Update ChequePaymentChange godoc
// @Description Update ChequePaymentChange
// @Tags		ChequePaymentChange
// @Param		id  path      string  true  "ChequePaymentChange ID"
// @Param		ChequePaymentChange  body      models.ChequePaymentChange  true  "ChequePaymentChange"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequepayment/chequepaymentchange/{id} [put]
func (h ChequePaymentChangeHttp) UpdateChequePaymentChange(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.ChequePaymentChange{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateChequePaymentChange(holdingCode, id, authUsername, *docReq)

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

// Delete ChequePaymentChange godoc
// @Description Delete ChequePaymentChange
// @Tags		ChequePaymentChange
// @Param		id  path      string  true  "ChequePaymentChange ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequepayment/chequepaymentchange/{id} [delete]
func (h ChequePaymentChangeHttp) DeleteChequePaymentChange(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteChequePaymentChange(holdingCode, id, authUsername)

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

// Delete ChequePaymentChange godoc
// @Description Delete ChequePaymentChange
// @Tags		ChequePaymentChange
// @Param		ChequePaymentChange  body      []string  true  "ChequePaymentChange GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequepayment/chequepaymentchange [delete]
func (h ChequePaymentChangeHttp) DeleteChequePaymentChangeByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteChequePaymentChangeByGUIDs(holdingCode, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get ChequePaymentChange godoc
// @Description get ChequePaymentChange info by guidfixed
// @Tags		ChequePaymentChange
// @Param		id  path      string  true  "ChequePaymentChange guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequepayment/chequepaymentchange/{id} [get]
func (h ChequePaymentChangeHttp) InfoChequePaymentChange(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get ChequePaymentChange %v", id)
	doc, err := h.svc.InfoChequePaymentChange(holdingCode, id)

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

// Get ChequePaymentChange By Code godoc
// @Description get ChequePaymentChange info by Code
// @Tags		ChequePaymentChange
// @Param		code  path      string  true  "ChequePaymentChange Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequepayment/chequepaymentchange/code/{code} [get]
func (h ChequePaymentChangeHttp) InfoChequePaymentChangeByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	code := ctx.Param("code")

	doc, err := h.svc.InfoChequePaymentChangeByCode(holdingCode, code)

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

// List ChequePaymentChange step godoc
// @Description get list step
// @Tags		ChequePaymentChange
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
// @Router /transaction/chequepayment/chequepaymentchange [get]
func (h ChequePaymentChangeHttp) SearchChequePaymentChangePage(ctx microservice.IContext) error {
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

	docList, pagination, err := h.svc.SearchChequePaymentChange(holdingCode, filters, pageable)

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

// List ChequePaymentChange godoc
// @Description search limit offset
// @Tags		ChequePaymentChange
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
// @Router /transaction/chequepayment/chequepaymentchange/list [get]
func (h ChequePaymentChangeHttp) SearchChequePaymentChangeStep(ctx microservice.IContext) error {
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

	docList, total, err := h.svc.SearchChequePaymentChangeStep(holdingCode, lang, filters, pageableStep)

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

// Create ChequePaymentChange Bulk godoc
// @Description Create ChequePaymentChange
// @Tags		ChequePaymentChange
// @Param		ChequePaymentChange  body      []models.ChequePaymentChange  true  "ChequePaymentChange"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequepayment/chequepaymentchange/bulk [post]
func (h ChequePaymentChangeHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()

	dataReq := []models.ChequePaymentChange{}
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
