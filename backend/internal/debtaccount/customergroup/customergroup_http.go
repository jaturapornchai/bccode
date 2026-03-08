package customergroup

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/debtaccount/customergroup/models"
	"smlcloudplatform/internal/debtaccount/customergroup/repositories"
	"smlcloudplatform/internal/debtaccount/customergroup/services"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
)

type ICustomerGroupHttp interface{}

type CustomerGroupHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.ICustomerGroupHttpService
}

func NewCustomerGroupHttp(ms *microservice.Microservice, cfg config.IConfig) CustomerGroupHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewCustomerGroupRepository(pst)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewCustomerGroupHttpService(repo, masterSyncCacheRepo)

	return CustomerGroupHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h CustomerGroupHttp) RegisterHttp() {

	h.ms.POST("/debtaccount/customer-group/bulk", h.SaveBulk)

	h.ms.GET("/debtaccount/customer-group", h.SearchCustomerGroupPage)
	h.ms.GET("/debtaccount/customer-group/list", h.SearchCustomerGroupStep)
	h.ms.POST("/debtaccount/customer-group", h.CreateCustomerGroup)
	h.ms.GET("/debtaccount/customer-group/:id", h.InfoCustomerGroup)
	h.ms.PUT("/debtaccount/customer-group/:id", h.UpdateCustomerGroup)
	h.ms.DELETE("/debtaccount/customer-group/:id", h.DeleteCustomerGroup)
	h.ms.DELETE("/debtaccount/customer-group", h.DeleteCustomerGroupByGUIDs)
}

// Create CustomerGroup godoc
// @Description Create CustomerGroup
// @Tags		CustomerGroup
// @Param		CustomerGroup  body      models.CustomerGroup  true  "CustomerGroup"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/customer-group [post]
func (h CustomerGroupHttp) CreateCustomerGroup(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.CustomerGroup{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateCustomerGroup(shopID, authUsername, *docReq)

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

// Update CustomerGroup godoc
// @Description Update CustomerGroup
// @Tags		CustomerGroup
// @Param		id  path      string  true  "CustomerGroup ID"
// @Param		CustomerGroup  body      models.CustomerGroup  true  "CustomerGroup"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/customer-group/{id} [put]
func (h CustomerGroupHttp) UpdateCustomerGroup(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.CustomerGroup{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateCustomerGroup(shopID, id, authUsername, *docReq)

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

// Delete CustomerGroup godoc
// @Description Delete CustomerGroup
// @Tags		CustomerGroup
// @Param		id  path      string  true  "CustomerGroup ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/customer-group/{id} [delete]
func (h CustomerGroupHttp) DeleteCustomerGroup(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteCustomerGroup(shopID, id, authUsername)

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

// Delete CustomerGroup godoc
// @Description Delete CustomerGroup
// @Tags		CustomerGroup
// @Param		CustomerGroup  body      []string  true  "CustomerGroup GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/customer-group [delete]
func (h CustomerGroupHttp) DeleteCustomerGroupByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteCustomerGroupByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get CustomerGroup godoc
// @Description get struct array by ID
// @Tags		CustomerGroup
// @Param		id  path      string  true  "CustomerGroup ID"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/customer-group/{id} [get]
func (h CustomerGroupHttp) InfoCustomerGroup(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get CustomerGroup %v", id)
	doc, err := h.svc.InfoCustomerGroup(shopID, id)

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

// List CustomerGroup godoc
// @Description get struct array by ID
// @Tags		CustomerGroup
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "page"
// @Param		limit	query	integer		false  "limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/customer-group [get]
func (h CustomerGroupHttp) SearchCustomerGroupPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.svc.SearchCustomerGroup(shopID, map[string]interface{}{}, pageable)

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

// List CustomerGroup godoc
// @Description search limit offset
// @Tags		CustomerGroup
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/customer-group/list [get]
func (h CustomerGroupHttp) SearchCustomerGroupStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	docList, total, err := h.svc.SearchCustomerGroupStep(shopID, lang, pageableStep)

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

// Create CustomerGroup Bulk godoc
// @Description Create CustomerGroup
// @Tags		CustomerGroup
// @Param		CustomerGroup  body      []models.CustomerGroup  true  "CustomerGroup"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/customer-group/bulk [post]
func (h CustomerGroupHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.CustomerGroup{}
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
