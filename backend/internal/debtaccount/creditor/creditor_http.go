package creditor

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/debtaccount/creditor/models"
	"smlcloudplatform/internal/debtaccount/creditor/repositories"
	"smlcloudplatform/internal/debtaccount/creditor/services"
	repositoriesGroup "smlcloudplatform/internal/debtaccount/creditorgroup/repositories"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

type ICreditorHttp interface{}

type CreditorHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.ICreditorHttpService
}

func NewCreditorHttp(ms *microservice.Microservice, cfg config.IConfig) CreditorHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	prod := ms.Producer(cfg.MQConfig())

	repo := repositories.NewCreditorRepository(pst)
	repoMq := repositories.NewCreditorMessageQueueRepository(prod)
	repoGroup := repositoriesGroup.NewCreditorGroupRepository(pst)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewCreditorHttpService(repo, repoMq, repoGroup, masterSyncCacheRepo)

	return CreditorHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h CreditorHttp) RegisterHttp() {

	h.ms.POST("/debtaccount/creditor/bulk", h.SaveBulk)

	h.ms.GET("/debtaccount/creditor", h.SearchCreditorPage)
	h.ms.GET("/debtaccount/creditor/list", h.SearchCreditorStep)
	h.ms.POST("/debtaccount/creditor", h.CreateCreditor)
	h.ms.GET("/debtaccount/creditor/:id", h.InfoCreditor)
	h.ms.GET("/debtaccount/creditor/code/:code", h.InfoCreditorByCode)
	h.ms.PUT("/debtaccount/creditor/:id", h.UpdateCreditor)
	h.ms.DELETE("/debtaccount/creditor/:id", h.DeleteCreditor)
	h.ms.DELETE("/debtaccount/creditor", h.DeleteCreditorByGUIDs)
}

// Create Creditor godoc
// @Description Create Creditor
// @Tags		Creditor
// @Param		Creditor  body      models.CreditorRequest  true  "Creditor"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/creditor [post]
func (h CreditorHttp) CreateCreditor(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.CreditorRequest{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateCreditor(shopID, authUsername, *docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      idx,
	})
	return nil
}

// Update Creditor godoc
// @Description Update Creditor
// @Tags		Creditor
// @Param		id  path      string  true  "Creditor ID"
// @Param		Creditor  body      models.CreditorRequest  true  "Creditor"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/creditor/{id} [put]
func (h CreditorHttp) UpdateCreditor(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.CreditorRequest{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateCreditor(shopID, id, authUsername, *docReq)

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

// Delete Creditor godoc
// @Description Delete Creditor
// @Tags		Creditor
// @Param		id  path      string  true  "Creditor ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/creditor/{id} [delete]
func (h CreditorHttp) DeleteCreditor(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteCreditor(shopID, id, authUsername)

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

// Delete Creditor godoc
// @Description Delete Creditor
// @Tags		Creditor
// @Param		Creditor  body      []string  true  "Creditor GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/creditor [delete]
func (h CreditorHttp) DeleteCreditorByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteCreditorByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get Creditor godoc
// @Description get struct array by ID
// @Tags		Creditor
// @Param		id  path      string  true  "Creditor ID"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/creditor/{id} [get]
func (h CreditorHttp) InfoCreditor(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get Creditor %v", id)
	doc, err := h.svc.InfoCreditor(shopID, id)

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

// Get Creditor By Code godoc
// @Description Get Creditor by code
// @Tags		Creditor
// @Param		code  path      string  true  "Creditor Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/creditor/code/{code} [get]
func (h CreditorHttp) InfoCreditorByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")

	doc, err := h.svc.InfoCreditorByCode(shopID, code)

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

// List Creditor godoc
// @Description get struct array by ID
// @Tags		Creditor
// @Param		q		query	string		false  "Search Value"
// @Param		groups		query	string		false  "groups guidfixed"
// @Param		page	query	integer		false  "page"
// @Param		limit	query	integer		false  "limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/creditor [get]
func (h CreditorHttp) SearchCreditorPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "groups",
			Field: "groups",
			Type:  requestfilter.FieldTypeString,
		},
	})

	docList, pagination, err := h.svc.SearchCreditor(shopID, filters, pageable)

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

// List Creditor godoc
// @Description search limit offset
// @Tags		Creditor
// @Param		q		query	string		false  "Search Value"
// @Param		groups		query	string		false  "groups guidfixed"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/creditor/list [get]
func (h CreditorHttp) SearchCreditorStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "groups",
			Field: "groups",
			Type:  requestfilter.FieldTypeString,
		},
	})

	docList, total, err := h.svc.SearchCreditorStep(shopID, lang, filters, pageableStep)

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

// Create Creditor Bulk godoc
// @Description Create Creditor
// @Tags		Creditor
// @Param		Creditor  body      []models.CreditorRequest  true  "Creditor"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/creditor/bulk [post]
func (h CreditorHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.CreditorRequest{}
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
