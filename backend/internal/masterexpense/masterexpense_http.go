package masterexpense

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/masterexpense/models"
	"smlcloudplatform/internal/masterexpense/repositories"
	"smlcloudplatform/internal/masterexpense/services"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IMasterExpenseHttp interface{}

type MasterExpenseHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IMasterExpenseHttpService
}

func NewMasterExpenseHttp(ms *microservice.Microservice, cfg config.IConfig) MasterExpenseHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewMasterExpenseRepository(pst)
	cacheRepo := repositories.NewMasterExpenseCacheRepository(cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewMasterExpenseHttpService(repo, cacheRepo, masterSyncCacheRepo, 15*time.Second)

	return MasterExpenseHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h MasterExpenseHttp) RegisterHttp() {

	h.ms.POST("/master-expense/bulk", h.SaveBulk)

	h.ms.GET("/master-expense", h.SearchMasterExpensePage)
	h.ms.GET("/master-expense/list", h.SearchMasterExpenseStep)
	h.ms.POST("/master-expense", h.CreateMasterExpense)
	h.ms.GET("/master-expense/:id", h.InfoMasterExpense)
	h.ms.GET("/master-expense/code/:code", h.InfoMasterExpenseByCode)
	h.ms.PUT("/master-expense/:id", h.UpdateMasterExpense)
	h.ms.DELETE("/master-expense/:id", h.DeleteMasterExpense)
	h.ms.DELETE("/master-expense", h.DeleteMasterExpenseByGUIDs)
}

// Create MasterExpense godoc
// @Description Create MasterExpense
// @Tags		MasterExpense
// @Param		MasterExpense  body      models.MasterExpense  true  "MasterExpense"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /master-expense [post]
func (h MasterExpenseHttp) CreateMasterExpense(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.MasterExpense{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateMasterExpense(shopID, authUsername, *docReq)

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

// Update MasterExpense godoc
// @Description Update MasterExpense
// @Tags		MasterExpense
// @Param		id  path      string  true  "MasterExpense ID"
// @Param		MasterExpense  body      models.MasterExpense  true  "MasterExpense"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /master-expense/{id} [put]
func (h MasterExpenseHttp) UpdateMasterExpense(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.MasterExpense{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateMasterExpense(shopID, id, authUsername, *docReq)

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

// Delete MasterExpense godoc
// @Description Delete MasterExpense
// @Tags		MasterExpense
// @Param		id  path      string  true  "MasterExpense ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /master-expense/{id} [delete]
func (h MasterExpenseHttp) DeleteMasterExpense(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteMasterExpense(shopID, id, authUsername)

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

// Delete MasterExpense godoc
// @Description Delete MasterExpense
// @Tags		MasterExpense
// @Param		MasterExpense  body      []string  true  "MasterExpense GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /master-expense [delete]
func (h MasterExpenseHttp) DeleteMasterExpenseByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteMasterExpenseByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get MasterExpense godoc
// @Description get MasterExpense info by guidfixed
// @Tags		MasterExpense
// @Param		id  path      string  true  "MasterExpense guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /master-expense/{id} [get]
func (h MasterExpenseHttp) InfoMasterExpense(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get MasterExpense %v", id)
	doc, err := h.svc.InfoMasterExpense(shopID, id)

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

// Get MasterExpense By Code godoc
// @Description get MasterExpense info by Code
// @Tags		MasterExpense
// @Param		code  path      string  true  "MasterExpense Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /master-expense/code/{code} [get]
func (h MasterExpenseHttp) InfoMasterExpenseByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")

	doc, err := h.svc.InfoMasterExpenseByCode(shopID, code)

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

// List MasterExpense step godoc
// @Description get list step
// @Tags		MasterExpense
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /master-expense [get]
func (h MasterExpenseHttp) SearchMasterExpensePage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{})

	docList, pagination, err := h.svc.SearchMasterExpense(shopID, filters, pageable)

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

// List MasterExpense godoc
// @Description search limit offset
// @Tags		MasterExpense
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /master-expense/list [get]
func (h MasterExpenseHttp) SearchMasterExpenseStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{})

	docList, total, err := h.svc.SearchMasterExpenseStep(shopID, lang, filters, pageableStep)

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

// Create MasterExpense Bulk godoc
// @Description Create MasterExpense
// @Tags		MasterExpense
// @Param		MasterExpense  body      []models.MasterExpense  true  "MasterExpense"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /master-expense/bulk [post]
func (h MasterExpenseHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.MasterExpense{}
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
