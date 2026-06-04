package chequepaymentdisqualified

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/chequepaymentdisqualified/models"
	"smlcloudplatform/internal/transaction/chequepaymentdisqualified/repositories"
	"smlcloudplatform/internal/transaction/chequepaymentdisqualified/services"
	trancache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

type IChequePaymentDisqualifiedHttp interface{}

type ChequePaymentDisqualifiedHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IChequePaymentDisqualifiedHttpService
}

func NewChequePaymentDisqualifiedHttp(ms *microservice.Microservice, cfg config.IConfig) ChequePaymentDisqualifiedHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	producer := ms.Producer(cfg.MQConfig())

	repo := repositories.NewChequePaymentDisqualifiedRepository(pst)
	repoMq := repositories.NewChequePaymentDisqualifiedMessageQueueRepository(producer)

	transRepo := trancache.NewCacheRepository(cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewChequePaymentDisqualifiedHttpService(repo, transRepo, repoMq, masterSyncCacheRepo)

	return ChequePaymentDisqualifiedHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h ChequePaymentDisqualifiedHttp) RegisterHttp() {

	h.ms.POST("/transaction/chequepayment/chequepaymentdisqualified/bulk", h.SaveBulk)

	h.ms.GET("/transaction/chequepayment/chequepaymentdisqualified", h.SearchChequePaymentDisqualifiedPage)
	h.ms.GET("/transaction/chequepayment/chequepaymentdisqualified/list", h.SearchChequePaymentDisqualifiedStep)
	h.ms.POST("/transaction/chequepayment/chequepaymentdisqualified", h.CreateChequePaymentDisqualified)
	h.ms.GET("/transaction/chequepayment/chequepaymentdisqualified/:id", h.InfoChequePaymentDisqualified)
	h.ms.GET("/transaction/chequepayment/chequepaymentdisqualified/code/:code", h.InfoChequePaymentDisqualifiedByCode)
	h.ms.PUT("/transaction/chequepayment/chequepaymentdisqualified/:id", h.UpdateChequePaymentDisqualified)
	h.ms.DELETE("/transaction/chequepayment/chequepaymentdisqualified/:id", h.DeleteChequePaymentDisqualified)
	h.ms.DELETE("/transaction/chequepayment/chequepaymentdisqualified", h.DeleteChequePaymentDisqualifiedByGUIDs)
}

// Create ChequePaymentDisqualified godoc
// @Description Create ChequePaymentDisqualified
// @Tags		ChequePaymentDisqualified
// @Param		ChequePaymentDisqualified  body      models.ChequePaymentDisqualified  true  "ChequePaymentDisqualified"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequepayment/chequepaymentdisqualified [post]
func (h ChequePaymentDisqualifiedHttp) CreateChequePaymentDisqualified(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	docReq := &models.ChequePaymentDisqualified{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, docNo, err := h.svc.CreateChequePaymentDisqualified(holdingCode, authUsername, *docReq)

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

// Update ChequePaymentDisqualified godoc
// @Description Update ChequePaymentDisqualified
// @Tags		ChequePaymentDisqualified
// @Param		id  path      string  true  "ChequePaymentDisqualified ID"
// @Param		ChequePaymentDisqualified  body      models.ChequePaymentDisqualified  true  "ChequePaymentDisqualified"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequepayment/chequepaymentdisqualified/{id} [put]
func (h ChequePaymentDisqualifiedHttp) UpdateChequePaymentDisqualified(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.ChequePaymentDisqualified{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateChequePaymentDisqualified(holdingCode, id, authUsername, *docReq)

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

// Delete ChequePaymentDisqualified godoc
// @Description Delete ChequePaymentDisqualified
// @Tags		ChequePaymentDisqualified
// @Param		id  path      string  true  "ChequePaymentDisqualified ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequepayment/chequepaymentdisqualified/{id} [delete]
func (h ChequePaymentDisqualifiedHttp) DeleteChequePaymentDisqualified(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteChequePaymentDisqualified(holdingCode, id, authUsername)

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

// Delete ChequePaymentDisqualified godoc
// @Description Delete ChequePaymentDisqualified
// @Tags		ChequePaymentDisqualified
// @Param		ChequePaymentDisqualified  body      []string  true  "ChequePaymentDisqualified GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequepayment/chequepaymentdisqualified [delete]
func (h ChequePaymentDisqualifiedHttp) DeleteChequePaymentDisqualifiedByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteChequePaymentDisqualifiedByGUIDs(holdingCode, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get ChequePaymentDisqualified godoc
// @Description get ChequePaymentDisqualified info by guidfixed
// @Tags		ChequePaymentDisqualified
// @Param		id  path      string  true  "ChequePaymentDisqualified guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequepayment/chequepaymentdisqualified/{id} [get]
func (h ChequePaymentDisqualifiedHttp) InfoChequePaymentDisqualified(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get ChequePaymentDisqualified %v", id)
	doc, err := h.svc.InfoChequePaymentDisqualified(holdingCode, id)

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

// Get ChequePaymentDisqualified By Code godoc
// @Description get ChequePaymentDisqualified info by Code
// @Tags		ChequePaymentDisqualified
// @Param		code  path      string  true  "ChequePaymentDisqualified Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequepayment/chequepaymentdisqualified/code/{code} [get]
func (h ChequePaymentDisqualifiedHttp) InfoChequePaymentDisqualifiedByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	code := ctx.Param("code")

	doc, err := h.svc.InfoChequePaymentDisqualifiedByCode(holdingCode, code)

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

// List ChequePaymentDisqualified step godoc
// @Description get list step
// @Tags		ChequePaymentDisqualified
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
// @Router /transaction/chequepayment/chequepaymentdisqualified [get]
func (h ChequePaymentDisqualifiedHttp) SearchChequePaymentDisqualifiedPage(ctx microservice.IContext) error {
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

	docList, pagination, err := h.svc.SearchChequePaymentDisqualified(holdingCode, filters, pageable)

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

// List ChequePaymentDisqualified godoc
// @Description search limit offset
// @Tags		ChequePaymentDisqualified
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
// @Router /transaction/chequepayment/chequepaymentdisqualified/list [get]
func (h ChequePaymentDisqualifiedHttp) SearchChequePaymentDisqualifiedStep(ctx microservice.IContext) error {
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

	docList, total, err := h.svc.SearchChequePaymentDisqualifiedStep(holdingCode, lang, filters, pageableStep)

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

// Create ChequePaymentDisqualified Bulk godoc
// @Description Create ChequePaymentDisqualified
// @Tags		ChequePaymentDisqualified
// @Param		ChequePaymentDisqualified  body      []models.ChequePaymentDisqualified  true  "ChequePaymentDisqualified"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequepayment/chequepaymentdisqualified/bulk [post]
func (h ChequePaymentDisqualifiedHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()

	dataReq := []models.ChequePaymentDisqualified{}
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
