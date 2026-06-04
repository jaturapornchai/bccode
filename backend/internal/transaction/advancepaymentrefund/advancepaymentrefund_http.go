package advancepaymentrefund

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/advancepaymentrefund/models"
	"smlcloudplatform/internal/transaction/advancepaymentrefund/repositories"
	"smlcloudplatform/internal/transaction/advancepaymentrefund/services"
	trancache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

type IAdvancePaymentRefundHttp interface{}

type AdvancePaymentRefundHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IAdvancePaymentRefundHttpService
}

func NewAdvancePaymentRefundHttp(ms *microservice.Microservice, cfg config.IConfig) AdvancePaymentRefundHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	producer := ms.Producer(cfg.MQConfig())

	repo := repositories.NewAdvancePaymentRefundRepository(pst)
	repoMq := repositories.NewAdvancePaymentRefundMessageQueueRepository(producer)

	transRepo := trancache.NewCacheRepository(cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewAdvancePaymentRefundHttpService(repo, transRepo, repoMq, masterSyncCacheRepo)

	return AdvancePaymentRefundHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h AdvancePaymentRefundHttp) RegisterHttp() {

	h.ms.POST("/transaction/advancepaymentrefund/bulk", h.SaveBulk)

	h.ms.GET("/transaction/advancepaymentrefund", h.SearchAdvancePaymentRefundPage)
	h.ms.GET("/transaction/advancepaymentrefund/list", h.SearchAdvancePaymentRefundStep)
	h.ms.POST("/transaction/advancepaymentrefund", h.CreateAdvancePaymentRefund)
	h.ms.GET("/transaction/advancepaymentrefund/:id", h.InfoAdvancePaymentRefund)
	h.ms.GET("/transaction/advancepaymentrefund/code/:code", h.InfoAdvancePaymentRefundByCode)
	h.ms.PUT("/transaction/advancepaymentrefund/:id", h.UpdateAdvancePaymentRefund)
	h.ms.DELETE("/transaction/advancepaymentrefund/:id", h.DeleteAdvancePaymentRefund)
	h.ms.DELETE("/transaction/advancepaymentrefund", h.DeleteAdvancePaymentRefundByGUIDs)
}

// Create AdvancePaymentRefund godoc
// @Description Create AdvancePaymentRefund
// @Tags		AdvancePaymentRefund
// @Param		AdvancePaymentRefund  body      models.AdvancePaymentRefund  true  "AdvancePaymentRefund"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/advancepaymentrefund [post]
func (h AdvancePaymentRefundHttp) CreateAdvancePaymentRefund(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	docReq := &models.AdvancePaymentRefund{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, docNo, err := h.svc.CreateAdvancePaymentRefund(holdingCode, authUsername, *docReq)

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

// Update AdvancePaymentRefund godoc
// @Description Update AdvancePaymentRefund
// @Tags		AdvancePaymentRefund
// @Param		id  path      string  true  "AdvancePaymentRefund ID"
// @Param		AdvancePaymentRefund  body      models.AdvancePaymentRefund  true  "AdvancePaymentRefund"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/advancepaymentrefund/{id} [put]
func (h AdvancePaymentRefundHttp) UpdateAdvancePaymentRefund(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.AdvancePaymentRefund{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateAdvancePaymentRefund(holdingCode, id, authUsername, *docReq)

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

// Delete AdvancePaymentRefund godoc
// @Description Delete AdvancePaymentRefund
// @Tags		AdvancePaymentRefund
// @Param		id  path      string  true  "AdvancePaymentRefund ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/advancepaymentrefund/{id} [delete]
func (h AdvancePaymentRefundHttp) DeleteAdvancePaymentRefund(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteAdvancePaymentRefund(holdingCode, id, authUsername)

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

// Delete AdvancePaymentRefund godoc
// @Description Delete AdvancePaymentRefund
// @Tags		AdvancePaymentRefund
// @Param		AdvancePaymentRefund  body      []string  true  "AdvancePaymentRefund GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/advancepaymentrefund [delete]
func (h AdvancePaymentRefundHttp) DeleteAdvancePaymentRefundByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteAdvancePaymentRefundByGUIDs(holdingCode, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get AdvancePaymentRefund godoc
// @Description get AdvancePaymentRefund info by guidfixed
// @Tags		AdvancePaymentRefund
// @Param		id  path      string  true  "AdvancePaymentRefund guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/advancepaymentrefund/{id} [get]
func (h AdvancePaymentRefundHttp) InfoAdvancePaymentRefund(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get AdvancePaymentRefund %v", id)
	doc, err := h.svc.InfoAdvancePaymentRefund(holdingCode, id)

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

// Get AdvancePaymentRefund By Code godoc
// @Description get AdvancePaymentRefund info by Code
// @Tags		AdvancePaymentRefund
// @Param		code  path      string  true  "AdvancePaymentRefund Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/advancepaymentrefund/code/{code} [get]
func (h AdvancePaymentRefundHttp) InfoAdvancePaymentRefundByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	code := ctx.Param("code")

	doc, err := h.svc.InfoAdvancePaymentRefundByCode(holdingCode, code)

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

// List AdvancePaymentRefund step godoc
// @Description get list step
// @Tags		AdvancePaymentRefund
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
// @Router /transaction/advancepaymentrefund [get]
func (h AdvancePaymentRefundHttp) SearchAdvancePaymentRefundPage(ctx microservice.IContext) error {
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

	docList, pagination, err := h.svc.SearchAdvancePaymentRefund(holdingCode, filters, pageable)

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

// List AdvancePaymentRefund godoc
// @Description search limit offset
// @Tags		AdvancePaymentRefund
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
// @Router /transaction/advancepaymentrefund/list [get]
func (h AdvancePaymentRefundHttp) SearchAdvancePaymentRefundStep(ctx microservice.IContext) error {
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

	docList, total, err := h.svc.SearchAdvancePaymentRefundStep(holdingCode, lang, filters, pageableStep)

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

// Create AdvancePaymentRefund Bulk godoc
// @Description Create AdvancePaymentRefund
// @Tags		AdvancePaymentRefund
// @Param		AdvancePaymentRefund  body      []models.AdvancePaymentRefund  true  "AdvancePaymentRefund"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/advancepaymentrefund/bulk [post]
func (h AdvancePaymentRefundHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()

	dataReq := []models.AdvancePaymentRefund{}
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
