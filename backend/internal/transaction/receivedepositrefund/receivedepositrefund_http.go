package receivedepositrefund

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/receivedepositrefund/models"
	"smlcloudplatform/internal/transaction/receivedepositrefund/repositories"
	"smlcloudplatform/internal/transaction/receivedepositrefund/services"
	trancache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

type IReceiveDepositRefundHttp interface{}

type ReceiveDepositRefundHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IReceiveDepositRefundHttpService
}

func NewReceiveDepositRefundHttp(ms *microservice.Microservice, cfg config.IConfig) ReceiveDepositRefundHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	producer := ms.Producer(cfg.MQConfig())

	repo := repositories.NewReceiveDepositRefundRepository(pst)
	repoMq := repositories.NewReceiveDepositRefundMessageQueueRepository(producer)

	transRepo := trancache.NewCacheRepository(cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewReceiveDepositRefundHttpService(repo, transRepo, repoMq, masterSyncCacheRepo)

	return ReceiveDepositRefundHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h ReceiveDepositRefundHttp) RegisterHttp() {

	h.ms.POST("/transaction/receivedepositrefund/bulk", h.SaveBulk)

	h.ms.GET("/transaction/receivedepositrefund", h.SearchReceiveDepositRefundPage)
	h.ms.GET("/transaction/receivedepositrefund/list", h.SearchReceiveDepositRefundStep)
	h.ms.POST("/transaction/receivedepositrefund", h.CreateReceiveDepositRefund)
	h.ms.GET("/transaction/receivedepositrefund/:id", h.InfoReceiveDepositRefund)
	h.ms.GET("/transaction/receivedepositrefund/code/:code", h.InfoReceiveDepositRefundByCode)
	h.ms.PUT("/transaction/receivedepositrefund/:id", h.UpdateReceiveDepositRefund)
	h.ms.DELETE("/transaction/receivedepositrefund/:id", h.DeleteReceiveDepositRefund)
	h.ms.DELETE("/transaction/receivedepositrefund", h.DeleteReceiveDepositRefundByGUIDs)
}

// Create ReceiveDepositRefund godoc
// @Description Create ReceiveDepositRefund
// @Tags		ReceiveDepositRefund
// @Param		ReceiveDepositRefund  body      models.ReceiveDepositRefund  true  "ReceiveDepositRefund"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/receivedepositrefund [post]
func (h ReceiveDepositRefundHttp) CreateReceiveDepositRefund(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	docReq := &models.ReceiveDepositRefund{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, docNo, err := h.svc.CreateReceiveDepositRefund(holdingCode, authUsername, *docReq)

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

// Update ReceiveDepositRefund godoc
// @Description Update ReceiveDepositRefund
// @Tags		ReceiveDepositRefund
// @Param		id  path      string  true  "ReceiveDepositRefund ID"
// @Param		ReceiveDepositRefund  body      models.ReceiveDepositRefund  true  "ReceiveDepositRefund"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/receivedepositrefund/{id} [put]
func (h ReceiveDepositRefundHttp) UpdateReceiveDepositRefund(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.ReceiveDepositRefund{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateReceiveDepositRefund(holdingCode, id, authUsername, *docReq)

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

// Delete ReceiveDepositRefund godoc
// @Description Delete ReceiveDepositRefund
// @Tags		ReceiveDepositRefund
// @Param		id  path      string  true  "ReceiveDepositRefund ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/receivedepositrefund/{id} [delete]
func (h ReceiveDepositRefundHttp) DeleteReceiveDepositRefund(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteReceiveDepositRefund(holdingCode, id, authUsername)

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

// Delete ReceiveDepositRefund godoc
// @Description Delete ReceiveDepositRefund
// @Tags		ReceiveDepositRefund
// @Param		ReceiveDepositRefund  body      []string  true  "ReceiveDepositRefund GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/receivedepositrefund [delete]
func (h ReceiveDepositRefundHttp) DeleteReceiveDepositRefundByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteReceiveDepositRefundByGUIDs(holdingCode, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get ReceiveDepositRefund godoc
// @Description get ReceiveDepositRefund info by guidfixed
// @Tags		ReceiveDepositRefund
// @Param		id  path      string  true  "ReceiveDepositRefund guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/receivedepositrefund/{id} [get]
func (h ReceiveDepositRefundHttp) InfoReceiveDepositRefund(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get ReceiveDepositRefund %v", id)
	doc, err := h.svc.InfoReceiveDepositRefund(holdingCode, id)

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

// Get ReceiveDepositRefund By Code godoc
// @Description get ReceiveDepositRefund info by Code
// @Tags		ReceiveDepositRefund
// @Param		code  path      string  true  "ReceiveDepositRefund Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/receivedepositrefund/code/{code} [get]
func (h ReceiveDepositRefundHttp) InfoReceiveDepositRefundByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	code := ctx.Param("code")

	doc, err := h.svc.InfoReceiveDepositRefundByCode(holdingCode, code)

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

// List ReceiveDepositRefund step godoc
// @Description get list step
// @Tags		ReceiveDepositRefund
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
// @Router /transaction/receivedepositrefund [get]
func (h ReceiveDepositRefundHttp) SearchReceiveDepositRefundPage(ctx microservice.IContext) error {
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

	docList, pagination, err := h.svc.SearchReceiveDepositRefund(holdingCode, filters, pageable)

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

// List ReceiveDepositRefund godoc
// @Description search limit offset
// @Tags		ReceiveDepositRefund
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
// @Router /transaction/receivedepositrefund/list [get]
func (h ReceiveDepositRefundHttp) SearchReceiveDepositRefundStep(ctx microservice.IContext) error {
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

	docList, total, err := h.svc.SearchReceiveDepositRefundStep(holdingCode, lang, filters, pageableStep)

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

// Create ReceiveDepositRefund Bulk godoc
// @Description Create ReceiveDepositRefund
// @Tags		ReceiveDepositRefund
// @Param		ReceiveDepositRefund  body      []models.ReceiveDepositRefund  true  "ReceiveDepositRefund"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/receivedepositrefund/bulk [post]
func (h ReceiveDepositRefundHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()

	dataReq := []models.ReceiveDepositRefund{}
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
