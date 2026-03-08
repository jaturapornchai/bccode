package masterincome

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/masterincome/models"
	"smlcloudplatform/internal/masterincome/repositories"
	"smlcloudplatform/internal/masterincome/services"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IMasterIncomeHttp interface{}

type MasterIncomeHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IMasterIncomeHttpService
}

func NewMasterIncomeHttp(ms *microservice.Microservice, cfg config.IConfig) MasterIncomeHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewMasterIncomeRepository(pst)
	cacheRepo := repositories.NewMasterIncomeCacheRepository(cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewMasterIncomeHttpService(repo, cacheRepo, masterSyncCacheRepo, 15*time.Second)

	return MasterIncomeHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h MasterIncomeHttp) RegisterHttp() {

	h.ms.POST("/master-income/bulk", h.SaveBulk)

	h.ms.GET("/master-income", h.SearchMasterIncomePage)
	h.ms.GET("/master-income/list", h.SearchMasterIncomeStep)
	h.ms.POST("/master-income", h.CreateMasterIncome)
	h.ms.GET("/master-income/:id", h.InfoMasterIncome)
	h.ms.GET("/master-income/code/:code", h.InfoMasterIncomeByCode)
	h.ms.PUT("/master-income/:id", h.UpdateMasterIncome)
	h.ms.DELETE("/master-income/:id", h.DeleteMasterIncome)
	h.ms.DELETE("/master-income", h.DeleteMasterIncomeByGUIDs)
}

// Create MasterIncome godoc
// @Description Create MasterIncome
// @Tags		MasterIncome
// @Param		MasterIncome  body      models.MasterIncome  true  "MasterIncome"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /master-income [post]
func (h MasterIncomeHttp) CreateMasterIncome(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.MasterIncome{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateMasterIncome(shopID, authUsername, *docReq)

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

// Update MasterIncome godoc
// @Description Update MasterIncome
// @Tags		MasterIncome
// @Param		id  path      string  true  "MasterIncome ID"
// @Param		MasterIncome  body      models.MasterIncome  true  "MasterIncome"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /master-income/{id} [put]
func (h MasterIncomeHttp) UpdateMasterIncome(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.MasterIncome{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateMasterIncome(shopID, id, authUsername, *docReq)

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

// Delete MasterIncome godoc
// @Description Delete MasterIncome
// @Tags		MasterIncome
// @Param		id  path      string  true  "MasterIncome ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /master-income/{id} [delete]
func (h MasterIncomeHttp) DeleteMasterIncome(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteMasterIncome(shopID, id, authUsername)

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

// Delete MasterIncome godoc
// @Description Delete MasterIncome
// @Tags		MasterIncome
// @Param		MasterIncome  body      []string  true  "MasterIncome GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /master-income [delete]
func (h MasterIncomeHttp) DeleteMasterIncomeByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteMasterIncomeByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get MasterIncome godoc
// @Description get MasterIncome info by guidfixed
// @Tags		MasterIncome
// @Param		id  path      string  true  "MasterIncome guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /master-income/{id} [get]
func (h MasterIncomeHttp) InfoMasterIncome(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get MasterIncome %v", id)
	doc, err := h.svc.InfoMasterIncome(shopID, id)

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

// Get MasterIncome By Code godoc
// @Description get MasterIncome info by Code
// @Tags		MasterIncome
// @Param		code  path      string  true  "MasterIncome Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /master-income/code/{code} [get]
func (h MasterIncomeHttp) InfoMasterIncomeByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")

	doc, err := h.svc.InfoMasterIncomeByCode(shopID, code)

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

// List MasterIncome step godoc
// @Description get list step
// @Tags		MasterIncome
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /master-income [get]
func (h MasterIncomeHttp) SearchMasterIncomePage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{})

	docList, pagination, err := h.svc.SearchMasterIncome(shopID, filters, pageable)

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

// List MasterIncome godoc
// @Description search limit offset
// @Tags		MasterIncome
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /master-income/list [get]
func (h MasterIncomeHttp) SearchMasterIncomeStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{})

	docList, total, err := h.svc.SearchMasterIncomeStep(shopID, lang, filters, pageableStep)

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

// Create MasterIncome Bulk godoc
// @Description Create MasterIncome
// @Tags		MasterIncome
// @Param		MasterIncome  body      []models.MasterIncome  true  "MasterIncome"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /master-income/bulk [post]
func (h MasterIncomeHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.MasterIncome{}
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
