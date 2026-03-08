package chequedisqualified

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/chequedisqualified/models"
	"smlcloudplatform/internal/transaction/chequedisqualified/repositories"
	"smlcloudplatform/internal/transaction/chequedisqualified/services"
	trancache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

type IChequeDisqualifiedHttp interface{}

type ChequeDisqualifiedHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IChequeDisqualifiedHttpService
}

func NewChequeDisqualifiedHttp(ms *microservice.Microservice, cfg config.IConfig) ChequeDisqualifiedHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	producer := ms.Producer(cfg.MQConfig())

	repo := repositories.NewChequeDisqualifiedRepository(pst)
	repoMq := repositories.NewChequeDisqualifiedMessageQueueRepository(producer)

	transRepo := trancache.NewCacheRepository(cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewChequeDisqualifiedHttpService(repo, transRepo, repoMq, masterSyncCacheRepo)

	return ChequeDisqualifiedHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h ChequeDisqualifiedHttp) RegisterHttp() {

	h.ms.POST("/transaction/chequereceive/chequedisqualified/bulk", h.SaveBulk)

	h.ms.GET("/transaction/chequereceive/chequedisqualified", h.SearchChequeDisqualifiedPage)
	h.ms.GET("/transaction/chequereceive/chequedisqualified/list", h.SearchChequeDisqualifiedStep)
	h.ms.POST("/transaction/chequereceive/chequedisqualified", h.CreateChequeDisqualified)
	h.ms.GET("/transaction/chequereceive/chequedisqualified/:id", h.InfoChequeDisqualified)
	h.ms.GET("/transaction/chequereceive/chequedisqualified/code/:code", h.InfoChequeDisqualifiedByCode)
	h.ms.PUT("/transaction/chequereceive/chequedisqualified/:id", h.UpdateChequeDisqualified)
	h.ms.DELETE("/transaction/chequereceive/chequedisqualified/:id", h.DeleteChequeDisqualified)
	h.ms.DELETE("/transaction/chequereceive/chequedisqualified", h.DeleteChequeDisqualifiedByGUIDs)
}

// Create ChequeDisqualified godoc
// @Description Create ChequeDisqualified
// @Tags		ChequeDisqualified
// @Param		ChequeDisqualified  body      models.ChequeDisqualified  true  "ChequeDisqualified"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequereceive/chequedisqualified [post]
func (h ChequeDisqualifiedHttp) CreateChequeDisqualified(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.ChequeDisqualified{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, docNo, err := h.svc.CreateChequeDisqualified(shopID, authUsername, *docReq)

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

// Update ChequeDisqualified godoc
// @Description Update ChequeDisqualified
// @Tags		ChequeDisqualified
// @Param		id  path      string  true  "ChequeDisqualified ID"
// @Param		ChequeDisqualified  body      models.ChequeDisqualified  true  "ChequeDisqualified"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequereceive/chequedisqualified/{id} [put]
func (h ChequeDisqualifiedHttp) UpdateChequeDisqualified(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.ChequeDisqualified{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateChequeDisqualified(shopID, id, authUsername, *docReq)

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

// Delete ChequeDisqualified godoc
// @Description Delete ChequeDisqualified
// @Tags		ChequeDisqualified
// @Param		id  path      string  true  "ChequeDisqualified ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequereceive/chequedisqualified/{id} [delete]
func (h ChequeDisqualifiedHttp) DeleteChequeDisqualified(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteChequeDisqualified(shopID, id, authUsername)

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

// Delete ChequeDisqualified godoc
// @Description Delete ChequeDisqualified
// @Tags		ChequeDisqualified
// @Param		ChequeDisqualified  body      []string  true  "ChequeDisqualified GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequereceive/chequedisqualified [delete]
func (h ChequeDisqualifiedHttp) DeleteChequeDisqualifiedByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteChequeDisqualifiedByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get ChequeDisqualified godoc
// @Description get ChequeDisqualified info by guidfixed
// @Tags		ChequeDisqualified
// @Param		id  path      string  true  "ChequeDisqualified guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequereceive/chequedisqualified/{id} [get]
func (h ChequeDisqualifiedHttp) InfoChequeDisqualified(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get ChequeDisqualified %v", id)
	doc, err := h.svc.InfoChequeDisqualified(shopID, id)

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

// Get ChequeDisqualified By Code godoc
// @Description get ChequeDisqualified info by Code
// @Tags		ChequeDisqualified
// @Param		code  path      string  true  "ChequeDisqualified Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequereceive/chequedisqualified/code/{code} [get]
func (h ChequeDisqualifiedHttp) InfoChequeDisqualifiedByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")

	doc, err := h.svc.InfoChequeDisqualifiedByCode(shopID, code)

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

// List ChequeDisqualified step godoc
// @Description get list step
// @Tags		ChequeDisqualified
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
// @Router /transaction/chequereceive/chequedisqualified [get]
func (h ChequeDisqualifiedHttp) SearchChequeDisqualifiedPage(ctx microservice.IContext) error {
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

	docList, pagination, err := h.svc.SearchChequeDisqualified(shopID, filters, pageable)

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

// List ChequeDisqualified godoc
// @Description search limit offset
// @Tags		ChequeDisqualified
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
// @Router /transaction/chequereceive/chequedisqualified/list [get]
func (h ChequeDisqualifiedHttp) SearchChequeDisqualifiedStep(ctx microservice.IContext) error {
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

	docList, total, err := h.svc.SearchChequeDisqualifiedStep(shopID, lang, filters, pageableStep)

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

// Create ChequeDisqualified Bulk godoc
// @Description Create ChequeDisqualified
// @Tags		ChequeDisqualified
// @Param		ChequeDisqualified  body      []models.ChequeDisqualified  true  "ChequeDisqualified"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/chequereceive/chequedisqualified/bulk [post]
func (h ChequeDisqualifiedHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.ChequeDisqualified{}
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
