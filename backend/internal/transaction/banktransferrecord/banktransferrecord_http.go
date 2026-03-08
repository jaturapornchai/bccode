package banktransferrecord

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/banktransferrecord/models"
	"smlcloudplatform/internal/transaction/banktransferrecord/repositories"
	"smlcloudplatform/internal/transaction/banktransferrecord/services"
	trancache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

type IBankTransferRecordHttp interface{}

type BankTransferRecordHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IBankTransferRecordHttpService
}

func NewBankTransferRecordHttp(ms *microservice.Microservice, cfg config.IConfig) BankTransferRecordHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	producer := ms.Producer(cfg.MQConfig())

	repo := repositories.NewBankTransferRecordRepository(pst)
	repoMq := repositories.NewBankTransferRecordMessageQueueRepository(producer)

	transRepo := trancache.NewCacheRepository(cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewBankTransferRecordHttpService(repo, transRepo, repoMq, masterSyncCacheRepo)

	return BankTransferRecordHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h BankTransferRecordHttp) RegisterHttp() {

	h.ms.POST("/transaction/bank/banktransferrecord/bulk", h.SaveBulk)

	h.ms.GET("/transaction/bank/banktransferrecord", h.SearchBankTransferRecordPage)
	h.ms.GET("/transaction/bank/banktransferrecord/list", h.SearchBankTransferRecordStep)
	h.ms.POST("/transaction/bank/banktransferrecord", h.CreateBankTransferRecord)
	h.ms.GET("/transaction/bank/banktransferrecord/:id", h.InfoBankTransferRecord)
	h.ms.GET("/transaction/bank/banktransferrecord/code/:code", h.InfoBankTransferRecordByCode)
	h.ms.PUT("/transaction/bank/banktransferrecord/:id", h.UpdateBankTransferRecord)
	h.ms.DELETE("/transaction/bank/banktransferrecord/:id", h.DeleteBankTransferRecord)
	h.ms.DELETE("/transaction/bank/banktransferrecord", h.DeleteBankTransferRecordByGUIDs)
}

// Create BankTransferRecord godoc
// @Description Create BankTransferRecord
// @Tags		BankTransferRecord
// @Param		BankTransferRecord  body      models.BankTransferRecord  true  "BankTransferRecord"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/banktransferrecord [post]
func (h BankTransferRecordHttp) CreateBankTransferRecord(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.BankTransferRecord{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, docNo, err := h.svc.CreateBankTransferRecord(shopID, authUsername, *docReq)

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

// Update BankTransferRecord godoc
// @Description Update BankTransferRecord
// @Tags		BankTransferRecord
// @Param		id  path      string  true  "BankTransferRecord ID"
// @Param		BankTransferRecord  body      models.BankTransferRecord  true  "BankTransferRecord"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/banktransferrecord/{id} [put]
func (h BankTransferRecordHttp) UpdateBankTransferRecord(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.BankTransferRecord{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateBankTransferRecord(shopID, id, authUsername, *docReq)

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

// Delete BankTransferRecord godoc
// @Description Delete BankTransferRecord
// @Tags		BankTransferRecord
// @Param		id  path      string  true  "BankTransferRecord ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/banktransferrecord/{id} [delete]
func (h BankTransferRecordHttp) DeleteBankTransferRecord(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteBankTransferRecord(shopID, id, authUsername)

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

// Delete BankTransferRecord godoc
// @Description Delete BankTransferRecord
// @Tags		BankTransferRecord
// @Param		BankTransferRecord  body      []string  true  "BankTransferRecord GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/banktransferrecord [delete]
func (h BankTransferRecordHttp) DeleteBankTransferRecordByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteBankTransferRecordByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get BankTransferRecord godoc
// @Description get BankTransferRecord info by guidfixed
// @Tags		BankTransferRecord
// @Param		id  path      string  true  "BankTransferRecord guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/banktransferrecord/{id} [get]
func (h BankTransferRecordHttp) InfoBankTransferRecord(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get BankTransferRecord %v", id)
	doc, err := h.svc.InfoBankTransferRecord(shopID, id)

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

// Get BankTransferRecord By Code godoc
// @Description get BankTransferRecord info by Code
// @Tags		BankTransferRecord
// @Param		code  path      string  true  "BankTransferRecord Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/banktransferrecord/code/{code} [get]
func (h BankTransferRecordHttp) InfoBankTransferRecordByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")

	doc, err := h.svc.InfoBankTransferRecordByCode(shopID, code)

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

// List BankTransferRecord step godoc
// @Description get list step
// @Tags		BankTransferRecord
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
// @Router /transaction/bank/banktransferrecord [get]
func (h BankTransferRecordHttp) SearchBankTransferRecordPage(ctx microservice.IContext) error {
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

	docList, pagination, err := h.svc.SearchBankTransferRecord(shopID, filters, pageable)

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

// List BankTransferRecord godoc
// @Description search limit offset
// @Tags		BankTransferRecord
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
// @Router /transaction/bank/banktransferrecord/list [get]
func (h BankTransferRecordHttp) SearchBankTransferRecordStep(ctx microservice.IContext) error {
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

	docList, total, err := h.svc.SearchBankTransferRecordStep(shopID, lang, filters, pageableStep)

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

// Create BankTransferRecord Bulk godoc
// @Description Create BankTransferRecord
// @Tags		BankTransferRecord
// @Param		BankTransferRecord  body      []models.BankTransferRecord  true  "BankTransferRecord"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/bank/banktransferrecord/bulk [post]
func (h BankTransferRecordHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.BankTransferRecord{}
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
