package chequepaymentreturn

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/chequepaymentreturn/models"
	"smlcloudplatform/internal/transaction/chequepaymentreturn/repositories"
	"smlcloudplatform/internal/transaction/chequepaymentreturn/services"
	trancache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

type IChequePaymentReturnHttp interface{}

type ChequePaymentReturnHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IChequePaymentReturnHttpService
}

func NewChequePaymentReturnHttp(ms *microservice.Microservice, cfg config.IConfig) ChequePaymentReturnHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	producer := ms.Producer(cfg.MQConfig())

	repo := repositories.NewChequePaymentReturnRepository(pst)
	repoMq := repositories.NewChequePaymentReturnMessageQueueRepository(producer)

	transRepo := trancache.NewCacheRepository(cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewChequePaymentReturnHttpService(repo, transRepo, repoMq, masterSyncCacheRepo)

	return ChequePaymentReturnHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h ChequePaymentReturnHttp) RegisterHttp() {

	h.ms.POST("/transaction/chequepayment/chequepaymentreturn/bulk", h.SaveBulk)

	h.ms.GET("/transaction/chequepayment/chequepaymentreturn", h.SearchChequePaymentReturnPage)
	h.ms.GET("/transaction/chequepayment/chequepaymentreturn/list", h.SearchChequePaymentReturnStep)
	h.ms.POST("/transaction/chequepayment/chequepaymentreturn", h.CreateChequePaymentReturn)
	h.ms.GET("/transaction/chequepayment/chequepaymentreturn/:id", h.InfoChequePaymentReturn)
	h.ms.GET("/transaction/chequepayment/chequepaymentreturn/code/:code", h.InfoChequePaymentReturnByCode)
	h.ms.PUT("/transaction/chequepayment/chequepaymentreturn/:id", h.UpdateChequePaymentReturn)
	h.ms.DELETE("/transaction/chequepayment/chequepaymentreturn/:id", h.DeleteChequePaymentReturn)
	h.ms.DELETE("/transaction/chequepayment/chequepaymentreturn", h.DeleteChequePaymentReturnByGUIDs)
}

// Create ChequePaymentReturn godoc
// @Description Create ChequePaymentReturn
// @Tags		ChequePaymentReturn
// @Param		ChequePaymentReturn  body      models.ChequePaymentReturn  true  "ChequePaymentReturn"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequepayment/chequepaymentreturn [post]
func (h ChequePaymentReturnHttp) CreateChequePaymentReturn(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.ChequePaymentReturn{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, docNo, err := h.svc.CreateChequePaymentReturn(shopID, authUsername, *docReq)

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

// Update ChequePaymentReturn godoc
// @Description Update ChequePaymentReturn
// @Tags		ChequePaymentReturn
// @Param		id  path      string  true  "ChequePaymentReturn ID"
// @Param		ChequePaymentReturn  body      models.ChequePaymentReturn  true  "ChequePaymentReturn"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequepayment/chequepaymentreturn/{id} [put]
func (h ChequePaymentReturnHttp) UpdateChequePaymentReturn(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.ChequePaymentReturn{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateChequePaymentReturn(shopID, id, authUsername, *docReq)

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

// Delete ChequePaymentReturn godoc
// @Description Delete ChequePaymentReturn
// @Tags		ChequePaymentReturn
// @Param		id  path      string  true  "ChequePaymentReturn ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequepayment/chequepaymentreturn/{id} [delete]
func (h ChequePaymentReturnHttp) DeleteChequePaymentReturn(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteChequePaymentReturn(shopID, id, authUsername)

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

// Delete ChequePaymentReturn godoc
// @Description Delete ChequePaymentReturn
// @Tags		ChequePaymentReturn
// @Param		ChequePaymentReturn  body      []string  true  "ChequePaymentReturn GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequepayment/chequepaymentreturn [delete]
func (h ChequePaymentReturnHttp) DeleteChequePaymentReturnByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteChequePaymentReturnByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get ChequePaymentReturn godoc
// @Description get ChequePaymentReturn info by guidfixed
// @Tags		ChequePaymentReturn
// @Param		id  path      string  true  "ChequePaymentReturn guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequepayment/chequepaymentreturn/{id} [get]
func (h ChequePaymentReturnHttp) InfoChequePaymentReturn(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get ChequePaymentReturn %v", id)
	doc, err := h.svc.InfoChequePaymentReturn(shopID, id)

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

// Get ChequePaymentReturn By Code godoc
// @Description get ChequePaymentReturn info by Code
// @Tags		ChequePaymentReturn
// @Param		code  path      string  true  "ChequePaymentReturn Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequepayment/chequepaymentreturn/code/{code} [get]
func (h ChequePaymentReturnHttp) InfoChequePaymentReturnByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")

	doc, err := h.svc.InfoChequePaymentReturnByCode(shopID, code)

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

// List ChequePaymentReturn step godoc
// @Description get list step
// @Tags		ChequePaymentReturn
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
// @Router /transaction/chequepayment/chequepaymentreturn [get]
func (h ChequePaymentReturnHttp) SearchChequePaymentReturnPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

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

	docList, pagination, err := h.svc.SearchChequePaymentReturn(shopID, filters, pageable)

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

// List ChequePaymentReturn godoc
// @Description search limit offset
// @Tags		ChequePaymentReturn
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
// @Router /transaction/chequepayment/chequepaymentreturn/list [get]
func (h ChequePaymentReturnHttp) SearchChequePaymentReturnStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

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

	docList, total, err := h.svc.SearchChequePaymentReturnStep(shopID, lang, filters, pageableStep)

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

// Create ChequePaymentReturn Bulk godoc
// @Description Create ChequePaymentReturn
// @Tags		ChequePaymentReturn
// @Param		ChequePaymentReturn  body      []models.ChequePaymentReturn  true  "ChequePaymentReturn"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequepayment/chequepaymentreturn/bulk [post]
func (h ChequePaymentReturnHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.ChequePaymentReturn{}
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
