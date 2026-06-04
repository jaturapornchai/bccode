package creditcardwithdrawal

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/creditcardwithdrawal/models"
	"smlcloudplatform/internal/transaction/creditcardwithdrawal/repositories"
	"smlcloudplatform/internal/transaction/creditcardwithdrawal/services"
	trancache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

type ICreditCardWithdrawalHttp interface{}

type CreditCardWithdrawalHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.ICreditCardWithdrawalHttpService
}

func NewCreditCardWithdrawalHttp(ms *microservice.Microservice, cfg config.IConfig) CreditCardWithdrawalHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	producer := ms.Producer(cfg.MQConfig())

	repo := repositories.NewCreditCardWithdrawalRepository(pst)
	repoMq := repositories.NewCreditCardWithdrawalMessageQueueRepository(producer)

	transRepo := trancache.NewCacheRepository(cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewCreditCardWithdrawalHttpService(repo, transRepo, repoMq, masterSyncCacheRepo)

	return CreditCardWithdrawalHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h CreditCardWithdrawalHttp) RegisterHttp() {

	h.ms.POST("/transaction/bank/creditcardwithdrawal/bulk", h.SaveBulk)

	h.ms.GET("/transaction/bank/creditcardwithdrawal", h.SearchCreditCardWithdrawalPage)
	h.ms.GET("/transaction/bank/creditcardwithdrawal/list", h.SearchCreditCardWithdrawalStep)
	h.ms.POST("/transaction/bank/creditcardwithdrawal", h.CreateCreditCardWithdrawal)
	h.ms.GET("/transaction/bank/creditcardwithdrawal/:id", h.InfoCreditCardWithdrawal)
	h.ms.GET("/transaction/bank/creditcardwithdrawal/code/:code", h.InfoCreditCardWithdrawalByCode)
	h.ms.PUT("/transaction/bank/creditcardwithdrawal/:id", h.UpdateCreditCardWithdrawal)
	h.ms.DELETE("/transaction/bank/creditcardwithdrawal/:id", h.DeleteCreditCardWithdrawal)
	h.ms.DELETE("/transaction/bank/creditcardwithdrawal", h.DeleteCreditCardWithdrawalByGUIDs)
}

// Create CreditCardWithdrawal godoc
// @Description Create CreditCardWithdrawal
// @Tags		CreditCardWithdrawal
// @Param		CreditCardWithdrawal  body      models.CreditCardWithdrawal  true  "CreditCardWithdrawal"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/creditcardwithdrawal [post]
func (h CreditCardWithdrawalHttp) CreateCreditCardWithdrawal(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	docReq := &models.CreditCardWithdrawal{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, docNo, err := h.svc.CreateCreditCardWithdrawal(holdingCode, authUsername, *docReq)

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

// Update CreditCardWithdrawal godoc
// @Description Update CreditCardWithdrawal
// @Tags		CreditCardWithdrawal
// @Param		id  path      string  true  "CreditCardWithdrawal ID"
// @Param		CreditCardWithdrawal  body      models.CreditCardWithdrawal  true  "CreditCardWithdrawal"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/creditcardwithdrawal/{id} [put]
func (h CreditCardWithdrawalHttp) UpdateCreditCardWithdrawal(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.CreditCardWithdrawal{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateCreditCardWithdrawal(holdingCode, id, authUsername, *docReq)

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

// Delete CreditCardWithdrawal godoc
// @Description Delete CreditCardWithdrawal
// @Tags		CreditCardWithdrawal
// @Param		id  path      string  true  "CreditCardWithdrawal ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/creditcardwithdrawal/{id} [delete]
func (h CreditCardWithdrawalHttp) DeleteCreditCardWithdrawal(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteCreditCardWithdrawal(holdingCode, id, authUsername)

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

// Delete CreditCardWithdrawal godoc
// @Description Delete CreditCardWithdrawal
// @Tags		CreditCardWithdrawal
// @Param		CreditCardWithdrawal  body      []string  true  "CreditCardWithdrawal GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/creditcardwithdrawal [delete]
func (h CreditCardWithdrawalHttp) DeleteCreditCardWithdrawalByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteCreditCardWithdrawalByGUIDs(holdingCode, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get CreditCardWithdrawal godoc
// @Description get CreditCardWithdrawal info by guidfixed
// @Tags		CreditCardWithdrawal
// @Param		id  path      string  true  "CreditCardWithdrawal guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/creditcardwithdrawal/{id} [get]
func (h CreditCardWithdrawalHttp) InfoCreditCardWithdrawal(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get CreditCardWithdrawal %v", id)
	doc, err := h.svc.InfoCreditCardWithdrawal(holdingCode, id)

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

// Get CreditCardWithdrawal By Code godoc
// @Description get CreditCardWithdrawal info by Code
// @Tags		CreditCardWithdrawal
// @Param		code  path      string  true  "CreditCardWithdrawal Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/creditcardwithdrawal/code/{code} [get]
func (h CreditCardWithdrawalHttp) InfoCreditCardWithdrawalByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	code := ctx.Param("code")

	doc, err := h.svc.InfoCreditCardWithdrawalByCode(holdingCode, code)

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

// List CreditCardWithdrawal step godoc
// @Description get list step
// @Tags		CreditCardWithdrawal
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
// @Router /transaction/bank/creditcardwithdrawal [get]
func (h CreditCardWithdrawalHttp) SearchCreditCardWithdrawalPage(ctx microservice.IContext) error {
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

	docList, pagination, err := h.svc.SearchCreditCardWithdrawal(holdingCode, filters, pageable)

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

// List CreditCardWithdrawal godoc
// @Description search limit offset
// @Tags		CreditCardWithdrawal
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
// @Router /transaction/bank/creditcardwithdrawal/list [get]
func (h CreditCardWithdrawalHttp) SearchCreditCardWithdrawalStep(ctx microservice.IContext) error {
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

	docList, total, err := h.svc.SearchCreditCardWithdrawalStep(holdingCode, lang, filters, pageableStep)

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

// Create CreditCardWithdrawal Bulk godoc
// @Description Create CreditCardWithdrawal
// @Tags		CreditCardWithdrawal
// @Param		CreditCardWithdrawal  body      []models.CreditCardWithdrawal  true  "CreditCardWithdrawal"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/creditcardwithdrawal/bulk [post]
func (h CreditCardWithdrawalHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()

	dataReq := []models.CreditCardWithdrawal{}
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
