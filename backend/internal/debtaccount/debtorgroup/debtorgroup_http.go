package debtorgroup

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/debtaccount/debtorgroup/models"
	"smlcloudplatform/internal/debtaccount/debtorgroup/repositories"
	"smlcloudplatform/internal/debtaccount/debtorgroup/services"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
)

type IDebtorGroupHttp interface{}

type DebtorGroupHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IDebtorGroupHttpService
}

func NewDebtorGroupHttp(ms *microservice.Microservice, cfg config.IConfig) DebtorGroupHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewDebtorGroupRepository(pst)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewDebtorGroupHttpService(repo, masterSyncCacheRepo)

	return DebtorGroupHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h DebtorGroupHttp) RegisterHttp() {

	h.ms.POST("/debtaccount/debtor-group/bulk", h.SaveBulk)

	h.ms.GET("/debtaccount/debtor-group", h.SearchDebtorGroupPage)
	h.ms.GET("/debtaccount/debtor-group/list", h.SearchDebtorGroupStep)
	h.ms.POST("/debtaccount/debtor-group", h.CreateDebtorGroup)
	h.ms.GET("/debtaccount/debtor-group/:id", h.InfoDebtorGroup)
	h.ms.PUT("/debtaccount/debtor-group/:id", h.UpdateDebtorGroup)
	h.ms.DELETE("/debtaccount/debtor-group/:id", h.DeleteDebtorGroup)
	h.ms.DELETE("/debtaccount/debtor-group", h.DeleteDebtorGroupByGUIDs)
}

// Create DebtorGroup godoc
// @Description Create DebtorGroup
// @Tags		DebtorGroup
// @Param		DebtorGroup  body      models.DebtorGroup  true  "DebtorGroup"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/debtor-group [post]
func (h DebtorGroupHttp) CreateDebtorGroup(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.DebtorGroup{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateDebtorGroup(shopID, authUsername, *docReq)

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

// Update DebtorGroup godoc
// @Description Update DebtorGroup
// @Tags		DebtorGroup
// @Param		id  path      string  true  "DebtorGroup ID"
// @Param		DebtorGroup  body      models.DebtorGroup  true  "DebtorGroup"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/debtor-group/{id} [put]
func (h DebtorGroupHttp) UpdateDebtorGroup(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.DebtorGroup{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateDebtorGroup(shopID, id, authUsername, *docReq)

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

// Delete DebtorGroup godoc
// @Description Delete DebtorGroup
// @Tags		DebtorGroup
// @Param		id  path      string  true  "DebtorGroup ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/debtor-group/{id} [delete]
func (h DebtorGroupHttp) DeleteDebtorGroup(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteDebtorGroup(shopID, id, authUsername)

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

// Delete DebtorGroup godoc
// @Description Delete DebtorGroup
// @Tags		DebtorGroup
// @Param		DebtorGroup  body      []string  true  "DebtorGroup GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/debtor-group [delete]
func (h DebtorGroupHttp) DeleteDebtorGroupByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteDebtorGroupByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get DebtorGroup godoc
// @Description get struct array by ID
// @Tags		DebtorGroup
// @Param		id  path      string  true  "DebtorGroup ID"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/debtor-group/{id} [get]
func (h DebtorGroupHttp) InfoDebtorGroup(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get DebtorGroup %v", id)
	doc, err := h.svc.InfoDebtorGroup(shopID, id)

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

// List DebtorGroup godoc
// @Description get struct array by ID
// @Tags		DebtorGroup
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "page"
// @Param		limit	query	integer		false  "limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/debtor-group [get]
func (h DebtorGroupHttp) SearchDebtorGroupPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.svc.SearchDebtorGroup(shopID, map[string]interface{}{}, pageable)

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

// List DebtorGroup godoc
// @Description search limit offset
// @Tags		DebtorGroup
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/debtor-group/list [get]
func (h DebtorGroupHttp) SearchDebtorGroupStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	docList, total, err := h.svc.SearchDebtorGroupStep(shopID, lang, pageableStep)

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

// Create DebtorGroup Bulk godoc
// @Description Create DebtorGroup
// @Tags		DebtorGroup
// @Param		DebtorGroup  body      []models.DebtorGroup  true  "DebtorGroup"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/debtor-group/bulk [post]
func (h DebtorGroupHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.DebtorGroup{}
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
