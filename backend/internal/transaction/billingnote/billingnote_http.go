package billingnote

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/billingnote/models"
	"smlcloudplatform/internal/transaction/billingnote/repositories"
	"smlcloudplatform/internal/transaction/billingnote/services"
	trancache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

type IBillingNoteHttp interface{}

type BillingNoteHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IBillingNoteHttpService
}

func NewBillingNoteHttp(ms *microservice.Microservice, cfg config.IConfig) BillingNoteHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewBillingNoteRepository(pst)
	repoMq := repositories.NewBillingNoteMessageQueueRepository(ms.Producer(cfg.MQConfig()))

	transRepo := trancache.NewCacheRepository(cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewBillingNoteHttpService(repo, repoMq, transRepo, masterSyncCacheRepo)

	return BillingNoteHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h BillingNoteHttp) RegisterHttp() {

	h.ms.POST("/transaction/billingnote/bulk", h.SaveBulk)

	h.ms.GET("/transaction/billingnote", h.SearchBillingNotePage)
	h.ms.GET("/transaction/billingnote/list", h.SearchBillingNoteStep)
	h.ms.POST("/transaction/billingnote", h.CreateBillingNote)
	h.ms.GET("/transaction/billingnote/:id", h.InfoBillingNote)
	h.ms.GET("/transaction/billingnote/code/:code", h.InfoBillingNoteByCode)
	h.ms.PUT("/transaction/billingnote/:id", h.UpdateBillingNote)
	h.ms.DELETE("/transaction/billingnote/:id", h.DeleteBillingNote)
	h.ms.DELETE("/transaction/billingnote", h.DeleteBillingNoteByGUIDs)
}

// Create BillingNote godoc
// @Description Create BillingNote
// @Tags		BillingNote
// @Param		BillingNote  body      models.BillingNote  true  "BillingNote"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/billingnote [post]
func (h BillingNoteHttp) CreateBillingNote(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	docReq := &models.BillingNote{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	// The company comes from the selected shop, never from the request body.
	docReq.BusinessCode = ctx.UserInfo().BusinessCode

	idx, docNo, err := h.svc.CreateBillingNote(holdingCode, authUsername, *docReq)

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

// Update BillingNote godoc
// @Description Update BillingNote
// @Tags		BillingNote
// @Param		id  path      string  true  "BillingNote ID"
// @Param		BillingNote  body      models.BillingNote  true  "BillingNote"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/billingnote/{id} [put]
func (h BillingNoteHttp) UpdateBillingNote(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.BillingNote{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	// Preserve businesscode from the user's active session.
	docReq.BusinessCode = ctx.UserInfo().BusinessCode

	err = h.svc.UpdateBillingNote(holdingCode, id, authUsername, *docReq)

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

// Delete BillingNote godoc
// @Description Delete BillingNote
// @Tags		BillingNote
// @Param		id  path      string  true  "BillingNote ID"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/billingnote/{id} [delete]
func (h BillingNoteHttp) DeleteBillingNote(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	err := h.svc.DeleteBillingNote(holdingCode, id, authUsername)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      id,
		Message: "Success",
	})
	return nil
}

// Delete BillingNote by GUIDs godoc
// @Description Delete BillingNote by GUIDs
// @Tags		BillingNote
// @Param		BillingNote  body      []string  true  "BillingNote GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/billingnote [delete]
func (h BillingNoteHttp) DeleteBillingNoteByGUIDs(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()

	docReq := []string{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.DeleteBillingNoteByGUIDs(holdingCode, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Message: "Success",
	})
	return nil
}

// Info BillingNote godoc
// @Description Info BillingNote
// @Tags		BillingNote
// @Param		id  path      string  true  "BillingNote ID"
// @Accept 		json
// @Success		200	{object}	models.BillingNoteInfo
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/billingnote/{id} [get]
func (h BillingNoteHttp) InfoBillingNote(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	doc, err := h.svc.InfoBillingNote(holdingCode, id)

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

// Info BillingNote by code godoc
// @Description Info BillingNote by code
// @Tags		BillingNote
// @Param		code  path      string  true  "DocNo"
// @Accept 		json
// @Success		200	{object}	models.BillingNoteInfo
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/billingnote/code/{code} [get]
func (h BillingNoteHttp) InfoBillingNoteByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	code := ctx.Param("code")

	doc, err := h.svc.InfoBillingNoteByCode(holdingCode, code)

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

// Search BillingNote godoc
// @Description Search BillingNote
// @Tags		BillingNote
// @Param		q		query	string		false  "Search Search"
// @Param		page	query	int			false  "Page"
// @Param		limit	query	int			false  "Limit"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/billingnote [get]
func (h BillingNoteHttp) SearchBillingNotePage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageable := utils.GetPageable(ctx.QueryParam)

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "docno",
			Field: "docno",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "custcode",
			Field: "custcode",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "fromdate",
			Field: "docdatetime",
			Type:  requestfilter.FieldTypeDate,
		},
		{
			Param: "todate",
			Field: "docdatetime",
			Type:  requestfilter.FieldTypeDate,
		},
	})

	docList, pagination, err := h.svc.SearchBillingNote(holdingCode, filters, pageable)

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

// Search BillingNote Step godoc
// @Description Search BillingNote Step
// @Tags		BillingNote
// @Param		q		query	string		false  "Search Search"
// @Param		page	query	int			false  "Page"
// @Param		limit	query	int			false  "Limit"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/billingnote/list [get]
func (h BillingNoteHttp) SearchBillingNoteStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "docno",
			Field: "docno",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "custcode",
			Field: "custcode",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "fromdate",
			Field: "docdatetime",
			Type:  requestfilter.FieldTypeDate,
		},
		{
			Param: "todate",
			Field: "docdatetime",
			Type:  requestfilter.FieldTypeDate,
		},
	})

	docList, total, err := h.svc.SearchBillingNoteStep(holdingCode, lang, filters, pageableStep)

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

// Save Bulk BillingNote godoc
// @Description Save Bulk BillingNote
// @Tags		BillingNote
// @Param		BillingNote  body      []models.BillingNote  true  "BillingNote"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/billingnote/bulk [post]
func (h BillingNoteHttp) SaveBulk(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()

	dataDocs := []models.BillingNote{}
	err := json.Unmarshal([]byte(input), &dataDocs)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	res, err := h.svc.SaveInBatch(holdingCode, authUsername, dataDocs)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.BulkResponse{
		Success:    true,
		BulkImport: res,
	})
	return nil
}
