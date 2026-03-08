package creditorgroup

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/debtaccount/creditorgroup/models"
	"smlcloudplatform/internal/debtaccount/creditorgroup/repositories"
	"smlcloudplatform/internal/debtaccount/creditorgroup/services"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
)

type ICreditorGroupHttp interface{}

type CreditorGroupHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.ICreditorGroupHttpService
}

func NewCreditorGroupHttp(ms *microservice.Microservice, cfg config.IConfig) CreditorGroupHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewCreditorGroupRepository(pst)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewCreditorGroupHttpService(repo, masterSyncCacheRepo)

	return CreditorGroupHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h CreditorGroupHttp) RegisterHttp() {

	h.ms.POST("/debtaccount/creditor-group/bulk", h.SaveBulk)

	h.ms.GET("/debtaccount/creditor-group", h.SearchCreditorGroupPage)
	h.ms.GET("/debtaccount/creditor-group/list", h.SearchCreditorGroupStep)
	h.ms.POST("/debtaccount/creditor-group", h.CreateCreditorGroup)
	h.ms.GET("/debtaccount/creditor-group/:id", h.InfoCreditorGroup)
	h.ms.PUT("/debtaccount/creditor-group/:id", h.UpdateCreditorGroup)
	h.ms.DELETE("/debtaccount/creditor-group/:id", h.DeleteCreditorGroup)
	h.ms.DELETE("/debtaccount/creditor-group", h.DeleteCreditorGroupByGUIDs)
}

// Create CreditorGroup godoc
// @Description Create CreditorGroup
// @Tags		CreditorGroup
// @Param		CreditorGroup  body      models.CreditorGroup  true  "CreditorGroup"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/creditor-group [post]
func (h CreditorGroupHttp) CreateCreditorGroup(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.CreditorGroup{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateCreditorGroup(shopID, authUsername, *docReq)

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

// Update CreditorGroup godoc
// @Description Update CreditorGroup
// @Tags		CreditorGroup
// @Param		id  path      string  true  "CreditorGroup ID"
// @Param		CreditorGroup  body      models.CreditorGroup  true  "CreditorGroup"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/creditor-group/{id} [put]
func (h CreditorGroupHttp) UpdateCreditorGroup(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.CreditorGroup{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateCreditorGroup(shopID, id, authUsername, *docReq)

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

// Delete CreditorGroup godoc
// @Description Delete CreditorGroup
// @Tags		CreditorGroup
// @Param		id  path      string  true  "CreditorGroup ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/creditor-group/{id} [delete]
func (h CreditorGroupHttp) DeleteCreditorGroup(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteCreditorGroup(shopID, id, authUsername)

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

// Delete CreditorGroup godoc
// @Description Delete CreditorGroup
// @Tags		CreditorGroup
// @Param		CreditorGroup  body      []string  true  "CreditorGroup GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/creditor-group [delete]
func (h CreditorGroupHttp) DeleteCreditorGroupByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteCreditorGroupByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get CreditorGroup godoc
// @Description get struct array by ID
// @Tags		CreditorGroup
// @Param		id  path      string  true  "CreditorGroup ID"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/creditor-group/{id} [get]
func (h CreditorGroupHttp) InfoCreditorGroup(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get CreditorGroup %v", id)
	doc, err := h.svc.InfoCreditorGroup(shopID, id)

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

// List CreditorGroup godoc
// @Description get struct array by ID
// @Tags		CreditorGroup
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "page"
// @Param		limit	query	integer		false  "limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/creditor-group [get]
func (h CreditorGroupHttp) SearchCreditorGroupPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.svc.SearchCreditorGroup(shopID, map[string]interface{}{}, pageable)

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

// List CreditorGroup godoc
// @Description search limit offset
// @Tags		CreditorGroup
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/creditor-group/list [get]
func (h CreditorGroupHttp) SearchCreditorGroupStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	docList, total, err := h.svc.SearchCreditorGroupStep(shopID, lang, pageableStep)

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

// Create CreditorGroup Bulk godoc
// @Description Create CreditorGroup
// @Tags		CreditorGroup
// @Param		CreditorGroup  body      []models.CreditorGroup  true  "CreditorGroup"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/creditor-group/bulk [post]
func (h CreditorGroupHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.CreditorGroup{}
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
