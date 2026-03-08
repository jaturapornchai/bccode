package paidadvancerefund

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/paidadvancerefund/models"
	"smlcloudplatform/internal/transaction/paidadvancerefund/repositories"
	"smlcloudplatform/internal/transaction/paidadvancerefund/services"
	trancache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

type IPaidAdvanceRefundHttp interface{}

type PaidAdvanceRefundHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IPaidAdvanceRefundHttpService
}

func NewPaidAdvanceRefundHttp(ms *microservice.Microservice, cfg config.IConfig) PaidAdvanceRefundHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	producer := ms.Producer(cfg.MQConfig())

	repo := repositories.NewPaidAdvanceRefundRepository(pst)
	repoMq := repositories.NewPaidAdvanceRefundMessageQueueRepository(producer)

	transRepo := trancache.NewCacheRepository(cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewPaidAdvanceRefundHttpService(repo, transRepo, repoMq, masterSyncCacheRepo)

	return PaidAdvanceRefundHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h PaidAdvanceRefundHttp) RegisterHttp() {

	h.ms.POST("/transaction/paidadvancerefund/bulk", h.SaveBulk)

	h.ms.GET("/transaction/paidadvancerefund", h.SearchPaidAdvanceRefundPage)
	h.ms.GET("/transaction/paidadvancerefund/list", h.SearchPaidAdvanceRefundStep)
	h.ms.POST("/transaction/paidadvancerefund", h.CreatePaidAdvanceRefund)
	h.ms.GET("/transaction/paidadvancerefund/:id", h.InfoPaidAdvanceRefund)
	h.ms.GET("/transaction/paidadvancerefund/code/:code", h.InfoPaidAdvanceRefundByCode)
	h.ms.PUT("/transaction/paidadvancerefund/:id", h.UpdatePaidAdvanceRefund)
	h.ms.DELETE("/transaction/paidadvancerefund/:id", h.DeletePaidAdvanceRefund)
	h.ms.DELETE("/transaction/paidadvancerefund", h.DeletePaidAdvanceRefundByGUIDs)
}

// Create PaidAdvanceRefund godoc
// @Description Create PaidAdvanceRefund
// @Tags		PaidAdvanceRefund
// @Param		PaidAdvanceRefund  body      models.PaidAdvanceRefund  true  "PaidAdvanceRefund"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/paidadvancerefund [post]
func (h PaidAdvanceRefundHttp) CreatePaidAdvanceRefund(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.PaidAdvanceRefund{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, docNo, err := h.svc.CreatePaidAdvanceRefund(shopID, authUsername, *docReq)

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

// Update PaidAdvanceRefund godoc
// @Description Update PaidAdvanceRefund
// @Tags		PaidAdvanceRefund
// @Param		id  path      string  true  "PaidAdvanceRefund ID"
// @Param		PaidAdvanceRefund  body      models.PaidAdvanceRefund  true  "PaidAdvanceRefund"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/paidadvancerefund/{id} [put]
func (h PaidAdvanceRefundHttp) UpdatePaidAdvanceRefund(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.PaidAdvanceRefund{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdatePaidAdvanceRefund(shopID, id, authUsername, *docReq)

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

// Delete PaidAdvanceRefund godoc
// @Description Delete PaidAdvanceRefund
// @Tags		PaidAdvanceRefund
// @Param		id  path      string  true  "PaidAdvanceRefund ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/paidadvancerefund/{id} [delete]
func (h PaidAdvanceRefundHttp) DeletePaidAdvanceRefund(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeletePaidAdvanceRefund(shopID, id, authUsername)

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

// Delete PaidAdvanceRefund godoc
// @Description Delete PaidAdvanceRefund
// @Tags		PaidAdvanceRefund
// @Param		PaidAdvanceRefund  body      []string  true  "PaidAdvanceRefund GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/paidadvancerefund [delete]
func (h PaidAdvanceRefundHttp) DeletePaidAdvanceRefundByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeletePaidAdvanceRefundByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get PaidAdvanceRefund godoc
// @Description get PaidAdvanceRefund info by guidfixed
// @Tags		PaidAdvanceRefund
// @Param		id  path      string  true  "PaidAdvanceRefund guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/paidadvancerefund/{id} [get]
func (h PaidAdvanceRefundHttp) InfoPaidAdvanceRefund(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get PaidAdvanceRefund %v", id)
	doc, err := h.svc.InfoPaidAdvanceRefund(shopID, id)

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

// Get PaidAdvanceRefund By Code godoc
// @Description get PaidAdvanceRefund info by Code
// @Tags		PaidAdvanceRefund
// @Param		code  path      string  true  "PaidAdvanceRefund Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/paidadvancerefund/code/{code} [get]
func (h PaidAdvanceRefundHttp) InfoPaidAdvanceRefundByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")

	doc, err := h.svc.InfoPaidAdvanceRefundByCode(shopID, code)

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

// List PaidAdvanceRefund step godoc
// @Description get list step
// @Tags		PaidAdvanceRefund
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
// @Router /transaction/paidadvancerefund [get]
func (h PaidAdvanceRefundHttp) SearchPaidAdvanceRefundPage(ctx microservice.IContext) error {
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

	docList, pagination, err := h.svc.SearchPaidAdvanceRefund(shopID, filters, pageable)

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

// List PaidAdvanceRefund godoc
// @Description search limit offset
// @Tags		PaidAdvanceRefund
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
// @Router /transaction/paidadvancerefund/list [get]
func (h PaidAdvanceRefundHttp) SearchPaidAdvanceRefundStep(ctx microservice.IContext) error {
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

	docList, total, err := h.svc.SearchPaidAdvanceRefundStep(shopID, lang, filters, pageableStep)

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

// Create PaidAdvanceRefund Bulk godoc
// @Description Create PaidAdvanceRefund
// @Tags		PaidAdvanceRefund
// @Param		PaidAdvanceRefund  body      []models.PaidAdvanceRefund  true  "PaidAdvanceRefund"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/paidadvancerefund/bulk [post]
func (h PaidAdvanceRefundHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.PaidAdvanceRefund{}
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
