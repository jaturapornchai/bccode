package costcenter

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/organization/costcenter/models"
	"smlcloudplatform/internal/organization/costcenter/repositories"
	"smlcloudplatform/internal/organization/costcenter/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
)

type ICostCenterHttp interface{}

type CostCenterHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.ICostCenterHttpService
}

func NewCostCenterHttp(ms *microservice.Microservice, cfg config.IConfig) CostCenterHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	prod := ms.Producer(cfg.MQConfig())

	repo := repositories.NewCostCenterRepository(pst)
	repoMessageQueue := repositories.NewCostCenterMessageQueueRepository(prod)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewCostCenterHttpService(repo, repoMessageQueue, masterSyncCacheRepo)

	return CostCenterHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h CostCenterHttp) RegisterHttp() {

	h.ms.POST("/organization/costcenter/bulk", h.SaveBulk)

	h.ms.GET("/organization/costcenter", h.SearchCostCenterPage)
	h.ms.GET("/organization/costcenter/list", h.SearchCostCenterStep)
	h.ms.POST("/organization/costcenter", h.CreateCostCenter)
	h.ms.GET("/organization/costcenter/:id", h.InfoCostCenter)
	h.ms.GET("/organization/costcenter/:costCenterCode/branch/:branchCode", h.InfoCostCenterByCode)
	h.ms.PUT("/organization/costcenter/:id", h.UpdateCostCenter)
	h.ms.DELETE("/organization/costcenter/:id", h.DeleteCostCenter)
	h.ms.DELETE("/organization/costcenter", h.DeleteCostCenterByGUIDs)
}

// Create CostCenter godoc
// @Description Create CostCenter
// @Tags		CostCenter
// @Param		CostCenter  body      models.CostCenter  true  "CostCenter"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/costcenter [post]
func (h CostCenterHttp) CreateCostCenter(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.CostCenter{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateCostCenter(shopID, authUsername, *docReq)

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

// Update CostCenter godoc
// @Description Update CostCenter
// @Tags		CostCenter
// @Param		id  path      string  true  "CostCenter ID"
// @Param		CostCenter  body      models.CostCenter  true  "CostCenter"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/costcenter/{id} [put]
func (h CostCenterHttp) UpdateCostCenter(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.CostCenter{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateCostCenter(shopID, id, authUsername, *docReq)

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

// Delete CostCenter godoc
// @Description Delete CostCenter
// @Tags		CostCenter
// @Param		id  path      string  true  "CostCenter ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/costcenter/{id} [delete]
func (h CostCenterHttp) DeleteCostCenter(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteCostCenter(shopID, id, authUsername)

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

// Delete CostCenter godoc
// @Description Delete CostCenter
// @Tags		CostCenter
// @Param		CostCenter  body      []string  true  "CostCenter GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/costcenter [delete]
func (h CostCenterHttp) DeleteCostCenterByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteCostCenterByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get CostCenter godoc
// @Description get CostCenter info by guidfixed
// @Tags		CostCenter
// @Param		id  path      string  true  "CostCenter guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/costcenter/{id} [get]
func (h CostCenterHttp) InfoCostCenter(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get CostCenter %v", id)
	doc, err := h.svc.InfoCostCenter(shopID, id)

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

// Get CostCenter By Code godoc
// @Description get CostCenter info by Code
// @Tags		CostCenter
// @Param		costCenterCode  path      string  true  "CostCenter Code"
// @Param		branchCode  path      string  true  "Branch Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/costcenter/{costCenterCode}/branch/{branchCode} [get]
func (h CostCenterHttp) InfoCostCenterByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	branchCode := ctx.Param("branchCode")
	costCenterCode := ctx.Param("costCenterCode")

	doc, err := h.svc.InfoCostCenterByCode(shopID, branchCode, costCenterCode)

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

// List CostCenter step godoc
// @Description get list step
// @Tags		CostCenter
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/costcenter [get]
func (h CostCenterHttp) SearchCostCenterPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.svc.SearchCostCenter(shopID, map[string]interface{}{}, pageable)

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

// List CostCenter godoc
// @Description search limit offset
// @Tags		CostCenter
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/costcenter/list [get]
func (h CostCenterHttp) SearchCostCenterStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	docList, total, err := h.svc.SearchCostCenterStep(shopID, lang, pageableStep)

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

// Create CostCenter Bulk godoc
// @Description Create CostCenter
// @Tags		CostCenter
// @Param		CostCenter  body      []models.CostCenter  true  "CostCenter"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/costcenter/bulk [post]
func (h CostCenterHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.CostCenter{}
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
