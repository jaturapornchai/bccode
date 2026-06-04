package withdrawalrecord

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	trancache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/transaction/withdrawalrecord/models"
	"smlcloudplatform/internal/transaction/withdrawalrecord/repositories"
	"smlcloudplatform/internal/transaction/withdrawalrecord/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

type IWithdrawalRecordHttp interface{}

type WithdrawalRecordHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IWithdrawalRecordHttpService
}

func NewWithdrawalRecordHttp(ms *microservice.Microservice, cfg config.IConfig) WithdrawalRecordHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	producer := ms.Producer(cfg.MQConfig())

	repo := repositories.NewWithdrawalRecordRepository(pst)
	repoMq := repositories.NewWithdrawalRecordMessageQueueRepository(producer)

	transRepo := trancache.NewCacheRepository(cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewWithdrawalRecordHttpService(repo, transRepo, repoMq, masterSyncCacheRepo)

	return WithdrawalRecordHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h WithdrawalRecordHttp) RegisterHttp() {

	h.ms.POST("/transaction/bank/withdrawalrecord/bulk", h.SaveBulk)

	h.ms.GET("/transaction/bank/withdrawalrecord", h.SearchWithdrawalRecordPage)
	h.ms.GET("/transaction/bank/withdrawalrecord/list", h.SearchWithdrawalRecordStep)
	h.ms.POST("/transaction/bank/withdrawalrecord", h.CreateWithdrawalRecord)
	h.ms.GET("/transaction/bank/withdrawalrecord/:id", h.InfoWithdrawalRecord)
	h.ms.GET("/transaction/bank/withdrawalrecord/code/:code", h.InfoWithdrawalRecordByCode)
	h.ms.PUT("/transaction/bank/withdrawalrecord/:id", h.UpdateWithdrawalRecord)
	h.ms.DELETE("/transaction/bank/withdrawalrecord/:id", h.DeleteWithdrawalRecord)
	h.ms.DELETE("/transaction/bank/withdrawalrecord", h.DeleteWithdrawalRecordByGUIDs)
}

// Create WithdrawalRecord godoc
// @Description Create WithdrawalRecord
// @Tags		WithdrawalRecord
// @Param		WithdrawalRecord  body      models.WithdrawalRecord  true  "WithdrawalRecord"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/withdrawalrecord [post]
func (h WithdrawalRecordHttp) CreateWithdrawalRecord(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	docReq := &models.WithdrawalRecord{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, docNo, err := h.svc.CreateWithdrawalRecord(holdingCode, authUsername, *docReq)

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

// Update WithdrawalRecord godoc
// @Description Update WithdrawalRecord
// @Tags		WithdrawalRecord
// @Param		id  path      string  true  "WithdrawalRecord ID"
// @Param		WithdrawalRecord  body      models.WithdrawalRecord  true  "WithdrawalRecord"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/withdrawalrecord/{id} [put]
func (h WithdrawalRecordHttp) UpdateWithdrawalRecord(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.WithdrawalRecord{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateWithdrawalRecord(holdingCode, id, authUsername, *docReq)

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

// Delete WithdrawalRecord godoc
// @Description Delete WithdrawalRecord
// @Tags		WithdrawalRecord
// @Param		id  path      string  true  "WithdrawalRecord ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/withdrawalrecord/{id} [delete]
func (h WithdrawalRecordHttp) DeleteWithdrawalRecord(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteWithdrawalRecord(holdingCode, id, authUsername)

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

// Delete WithdrawalRecord godoc
// @Description Delete WithdrawalRecord
// @Tags		WithdrawalRecord
// @Param		WithdrawalRecord  body      []string  true  "WithdrawalRecord GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/withdrawalrecord [delete]
func (h WithdrawalRecordHttp) DeleteWithdrawalRecordByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteWithdrawalRecordByGUIDs(holdingCode, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get WithdrawalRecord godoc
// @Description get WithdrawalRecord info by guidfixed
// @Tags		WithdrawalRecord
// @Param		id  path      string  true  "WithdrawalRecord guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/withdrawalrecord/{id} [get]
func (h WithdrawalRecordHttp) InfoWithdrawalRecord(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get WithdrawalRecord %v", id)
	doc, err := h.svc.InfoWithdrawalRecord(holdingCode, id)

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

// Get WithdrawalRecord By Code godoc
// @Description get WithdrawalRecord info by Code
// @Tags		WithdrawalRecord
// @Param		code  path      string  true  "WithdrawalRecord Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/withdrawalrecord/code/{code} [get]
func (h WithdrawalRecordHttp) InfoWithdrawalRecordByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	code := ctx.Param("code")

	doc, err := h.svc.InfoWithdrawalRecordByCode(holdingCode, code)

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

// List WithdrawalRecord step godoc
// @Description get list step
// @Tags		WithdrawalRecord
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
// @Router /transaction/bank/withdrawalrecord [get]
func (h WithdrawalRecordHttp) SearchWithdrawalRecordPage(ctx microservice.IContext) error {
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

	docList, pagination, err := h.svc.SearchWithdrawalRecord(holdingCode, filters, pageable)

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

// List WithdrawalRecord godoc
// @Description search limit offset
// @Tags		WithdrawalRecord
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
// @Router /transaction/bank/withdrawalrecord/list [get]
func (h WithdrawalRecordHttp) SearchWithdrawalRecordStep(ctx microservice.IContext) error {
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

	docList, total, err := h.svc.SearchWithdrawalRecordStep(holdingCode, lang, filters, pageableStep)

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

// Create WithdrawalRecord Bulk godoc
// @Description Create WithdrawalRecord
// @Tags		WithdrawalRecord
// @Param		WithdrawalRecord  body      []models.WithdrawalRecord  true  "WithdrawalRecord"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/withdrawalrecord/bulk [post]
func (h WithdrawalRecordHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()

	dataReq := []models.WithdrawalRecord{}
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
