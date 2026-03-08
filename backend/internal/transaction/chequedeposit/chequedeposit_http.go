package chequedeposit

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/chequedeposit/models"
	"smlcloudplatform/internal/transaction/chequedeposit/repositories"
	"smlcloudplatform/internal/transaction/chequedeposit/services"
	trancache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

type IChequeDepositHttp interface{}

type ChequeDepositHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IChequeDepositHttpService
}

func NewChequeDepositHttp(ms *microservice.Microservice, cfg config.IConfig) ChequeDepositHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	producer := ms.Producer(cfg.MQConfig())

	repo := repositories.NewChequeDepositRepository(pst)
	repoMq := repositories.NewChequeDepositMessageQueueRepository(producer)

	transRepo := trancache.NewCacheRepository(cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewChequeDepositHttpService(repo, transRepo, repoMq, masterSyncCacheRepo)

	return ChequeDepositHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h ChequeDepositHttp) RegisterHttp() {

	h.ms.POST("/transaction/chequereceive/chequedeposit/bulk", h.SaveBulk)

	h.ms.GET("/transaction/chequereceive/chequedeposit", h.SearchChequeDepositPage)
	h.ms.GET("/transaction/chequereceive/chequedeposit/list", h.SearchChequeDepositStep)
	h.ms.POST("/transaction/chequereceive/chequedeposit", h.CreateChequeDeposit)
	h.ms.GET("/transaction/chequereceive/chequedeposit/:id", h.InfoChequeDeposit)
	h.ms.GET("/transaction/chequereceive/chequedeposit/code/:code", h.InfoChequeDepositByCode)
	h.ms.PUT("/transaction/chequereceive/chequedeposit/:id", h.UpdateChequeDeposit)
	h.ms.DELETE("/transaction/chequereceive/chequedeposit/:id", h.DeleteChequeDeposit)
	h.ms.DELETE("/transaction/chequereceive/chequedeposit", h.DeleteChequeDepositByGUIDs)
}

// Create ChequeDeposit godoc
// @Description Create ChequeDeposit
// @Tags		ChequeDeposit
// @Param		ChequeDeposit  body      models.ChequeDeposit  true  "ChequeDeposit"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequereceive/chequedeposit [post]
func (h ChequeDepositHttp) CreateChequeDeposit(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.ChequeDeposit{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, docNo, err := h.svc.CreateChequeDeposit(shopID, authUsername, *docReq)

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

// Update ChequeDeposit godoc
// @Description Update ChequeDeposit
// @Tags		ChequeDeposit
// @Param		id  path      string  true  "ChequeDeposit ID"
// @Param		ChequeDeposit  body      models.ChequeDeposit  true  "ChequeDeposit"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequereceive/chequedeposit/{id} [put]
func (h ChequeDepositHttp) UpdateChequeDeposit(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.ChequeDeposit{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateChequeDeposit(shopID, id, authUsername, *docReq)

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

// Delete ChequeDeposit godoc
// @Description Delete ChequeDeposit
// @Tags		ChequeDeposit
// @Param		id  path      string  true  "ChequeDeposit ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequereceive/chequedeposit/{id} [delete]
func (h ChequeDepositHttp) DeleteChequeDeposit(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteChequeDeposit(shopID, id, authUsername)

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

// Delete ChequeDeposit godoc
// @Description Delete ChequeDeposit
// @Tags		ChequeDeposit
// @Param		ChequeDeposit  body      []string  true  "ChequeDeposit GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequereceive/chequedeposit [delete]
func (h ChequeDepositHttp) DeleteChequeDepositByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteChequeDepositByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get ChequeDeposit godoc
// @Description get ChequeDeposit info by guidfixed
// @Tags		ChequeDeposit
// @Param		id  path      string  true  "ChequeDeposit guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequereceive/chequedeposit/{id} [get]
func (h ChequeDepositHttp) InfoChequeDeposit(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get ChequeDeposit %v", id)
	doc, err := h.svc.InfoChequeDeposit(shopID, id)

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

// Get ChequeDeposit By Code godoc
// @Description get ChequeDeposit info by Code
// @Tags		ChequeDeposit
// @Param		code  path      string  true  "ChequeDeposit Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequereceive/chequedeposit/code/{code} [get]
func (h ChequeDepositHttp) InfoChequeDepositByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")

	doc, err := h.svc.InfoChequeDepositByCode(shopID, code)

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

// List ChequeDeposit step godoc
// @Description get list step
// @Tags		ChequeDeposit
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
// @Router /transaction/chequereceive/chequedeposit [get]
func (h ChequeDepositHttp) SearchChequeDepositPage(ctx microservice.IContext) error {
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

	docList, pagination, err := h.svc.SearchChequeDeposit(shopID, filters, pageable)

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

// List ChequeDeposit godoc
// @Description search limit offset
// @Tags		ChequeDeposit
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
// @Router /transaction/chequereceive/chequedeposit/list [get]
func (h ChequeDepositHttp) SearchChequeDepositStep(ctx microservice.IContext) error {
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

	docList, total, err := h.svc.SearchChequeDepositStep(shopID, lang, filters, pageableStep)

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

// Create ChequeDeposit Bulk godoc
// @Description Create ChequeDeposit
// @Tags		ChequeDeposit
// @Param		ChequeDeposit  body      []models.ChequeDeposit  true  "ChequeDeposit"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequereceive/chequedeposit/bulk [post]
func (h ChequeDepositHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.ChequeDeposit{}
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
