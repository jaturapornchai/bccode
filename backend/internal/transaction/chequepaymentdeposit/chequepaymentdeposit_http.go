package chequepaymentdeposit

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/chequepaymentdeposit/models"
	"smlcloudplatform/internal/transaction/chequepaymentdeposit/repositories"
	"smlcloudplatform/internal/transaction/chequepaymentdeposit/services"
	trancache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

type IChequePaymentDepositHttp interface{}

type ChequePaymentDepositHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IChequePaymentDepositHttpService
}

func NewChequePaymentDepositHttp(ms *microservice.Microservice, cfg config.IConfig) ChequePaymentDepositHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	producer := ms.Producer(cfg.MQConfig())

	repo := repositories.NewChequePaymentDepositRepository(pst)
	repoMq := repositories.NewChequePaymentDepositMessageQueueRepository(producer)

	transRepo := trancache.NewCacheRepository(cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewChequePaymentDepositHttpService(repo, transRepo, repoMq, masterSyncCacheRepo)

	return ChequePaymentDepositHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h ChequePaymentDepositHttp) RegisterHttp() {

	h.ms.POST("/transaction/chequepayment/chequepaymentdeposit/bulk", h.SaveBulk)

	h.ms.GET("/transaction/chequepayment/chequepaymentdeposit", h.SearchChequePaymentDepositPage)
	h.ms.GET("/transaction/chequepayment/chequepaymentdeposit/list", h.SearchChequePaymentDepositStep)
	h.ms.POST("/transaction/chequepayment/chequepaymentdeposit", h.CreateChequePaymentDeposit)
	h.ms.GET("/transaction/chequepayment/chequepaymentdeposit/:id", h.InfoChequePaymentDeposit)
	h.ms.GET("/transaction/chequepayment/chequepaymentdeposit/code/:code", h.InfoChequePaymentDepositByCode)
	h.ms.PUT("/transaction/chequepayment/chequepaymentdeposit/:id", h.UpdateChequePaymentDeposit)
	h.ms.DELETE("/transaction/chequepayment/chequepaymentdeposit/:id", h.DeleteChequePaymentDeposit)
	h.ms.DELETE("/transaction/chequepayment/chequepaymentdeposit", h.DeleteChequePaymentDepositByGUIDs)
}

// Create ChequePaymentDeposit godoc
// @Description Create ChequePaymentDeposit
// @Tags		ChequePaymentDeposit
// @Param		ChequePaymentDeposit  body      models.ChequePaymentDeposit  true  "ChequePaymentDeposit"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequepayment/chequepaymentdeposit [post]
func (h ChequePaymentDepositHttp) CreateChequePaymentDeposit(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	docReq := &models.ChequePaymentDeposit{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, docNo, err := h.svc.CreateChequePaymentDeposit(holdingCode, authUsername, *docReq)

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

// Update ChequePaymentDeposit godoc
// @Description Update ChequePaymentDeposit
// @Tags		ChequePaymentDeposit
// @Param		id  path      string  true  "ChequePaymentDeposit ID"
// @Param		ChequePaymentDeposit  body      models.ChequePaymentDeposit  true  "ChequePaymentDeposit"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequepayment/chequepaymentdeposit/{id} [put]
func (h ChequePaymentDepositHttp) UpdateChequePaymentDeposit(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.ChequePaymentDeposit{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateChequePaymentDeposit(holdingCode, id, authUsername, *docReq)

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

// Delete ChequePaymentDeposit godoc
// @Description Delete ChequePaymentDeposit
// @Tags		ChequePaymentDeposit
// @Param		id  path      string  true  "ChequePaymentDeposit ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequepayment/chequepaymentdeposit/{id} [delete]
func (h ChequePaymentDepositHttp) DeleteChequePaymentDeposit(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteChequePaymentDeposit(holdingCode, id, authUsername)

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

// Delete ChequePaymentDeposit godoc
// @Description Delete ChequePaymentDeposit
// @Tags		ChequePaymentDeposit
// @Param		ChequePaymentDeposit  body      []string  true  "ChequePaymentDeposit GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequepayment/chequepaymentdeposit [delete]
func (h ChequePaymentDepositHttp) DeleteChequePaymentDepositByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteChequePaymentDepositByGUIDs(holdingCode, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get ChequePaymentDeposit godoc
// @Description get ChequePaymentDeposit info by guidfixed
// @Tags		ChequePaymentDeposit
// @Param		id  path      string  true  "ChequePaymentDeposit guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequepayment/chequepaymentdeposit/{id} [get]
func (h ChequePaymentDepositHttp) InfoChequePaymentDeposit(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get ChequePaymentDeposit %v", id)
	doc, err := h.svc.InfoChequePaymentDeposit(holdingCode, id)

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

// Get ChequePaymentDeposit By Code godoc
// @Description get ChequePaymentDeposit info by Code
// @Tags		ChequePaymentDeposit
// @Param		code  path      string  true  "ChequePaymentDeposit Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequepayment/chequepaymentdeposit/code/{code} [get]
func (h ChequePaymentDepositHttp) InfoChequePaymentDepositByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	code := ctx.Param("code")

	doc, err := h.svc.InfoChequePaymentDepositByCode(holdingCode, code)

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

// List ChequePaymentDeposit step godoc
// @Description get list step
// @Tags		ChequePaymentDeposit
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
// @Router /transaction/chequepayment/chequepaymentdeposit [get]
func (h ChequePaymentDepositHttp) SearchChequePaymentDepositPage(ctx microservice.IContext) error {
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

	docList, pagination, err := h.svc.SearchChequePaymentDeposit(holdingCode, filters, pageable)

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

// List ChequePaymentDeposit godoc
// @Description search limit offset
// @Tags		ChequePaymentDeposit
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
// @Router /transaction/chequepayment/chequepaymentdeposit/list [get]
func (h ChequePaymentDepositHttp) SearchChequePaymentDepositStep(ctx microservice.IContext) error {
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

	docList, total, err := h.svc.SearchChequePaymentDepositStep(holdingCode, lang, filters, pageableStep)

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

// Create ChequePaymentDeposit Bulk godoc
// @Description Create ChequePaymentDeposit
// @Tags		ChequePaymentDeposit
// @Param		ChequePaymentDeposit  body      []models.ChequePaymentDeposit  true  "ChequePaymentDeposit"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequepayment/chequepaymentdeposit/bulk [post]
func (h ChequePaymentDepositHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()

	dataReq := []models.ChequePaymentDeposit{}
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
