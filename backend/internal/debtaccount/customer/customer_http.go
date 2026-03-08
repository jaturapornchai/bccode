package customer

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/debtaccount/customer/models"
	"smlcloudplatform/internal/debtaccount/customer/repositories"
	"smlcloudplatform/internal/debtaccount/customer/services"
	repositoriesGroup "smlcloudplatform/internal/debtaccount/customergroup/repositories"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
)

type ICustomerHttp interface{}

type CustomerHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.ICustomerHttpService
}

func NewCustomerHttp(ms *microservice.Microservice, cfg config.IConfig) CustomerHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewCustomerRepository(pst)
	repoGroup := repositoriesGroup.NewCustomerGroupRepository(pst)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewCustomerHttpService(repo, repoGroup, masterSyncCacheRepo)

	return CustomerHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h CustomerHttp) RegisterHttp() {

	h.ms.POST("/debtaccount/customer/bulk", h.SaveBulk)

	h.ms.GET("/debtaccount/customer", h.SearchCustomerPage)
	h.ms.GET("/debtaccount/customer/list", h.SearchCustomerStep)
	h.ms.POST("/debtaccount/customer", h.CreateCustomer)
	h.ms.GET("/debtaccount/customer/:id", h.InfoCustomer)
	h.ms.GET("/debtaccount/customer/code/:code", h.InfoCustomerByCode)
	h.ms.PUT("/debtaccount/customer/:id", h.UpdateCustomer)
	h.ms.DELETE("/debtaccount/customer/:id", h.DeleteCustomer)
	h.ms.DELETE("/debtaccount/customer", h.DeleteCustomerByGUIDs)
}

// Create Customer godoc
// @Description Create Customer
// @Tags		Customer
// @Param		Customer  body      models.CustomerRequest  true  "Customer"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/customer [post]
func (h CustomerHttp) CreateCustomer(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.CustomerRequest{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateCustomer(shopID, authUsername, *docReq)

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

// Update Customer godoc
// @Description Update Customer
// @Tags		Customer
// @Param		id  path      string  true  "Customer ID"
// @Param		Customer  body      models.CustomerRequest  true  "Customer"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/customer/{id} [put]
func (h CustomerHttp) UpdateCustomer(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.CustomerRequest{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateCustomer(shopID, id, authUsername, *docReq)

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

// Delete Customer godoc
// @Description Delete Customer
// @Tags		Customer
// @Param		id  path      string  true  "Customer ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/customer/{id} [delete]
func (h CustomerHttp) DeleteCustomer(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteCustomer(shopID, id, authUsername)

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

// Delete Customer godoc
// @Description Delete Customer
// @Tags		Customer
// @Param		Customer  body      []string  true  "Customer GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/customer [delete]
func (h CustomerHttp) DeleteCustomerByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteCustomerByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get Customer godoc
// @Description get info Customer by id
// @Tags		Customer
// @Param		id  path      string  true  "Customer ID"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/customer/{id} [get]
func (h CustomerHttp) InfoCustomer(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get Customer %v", id)
	doc, err := h.svc.InfoCustomer(shopID, id)

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

// Get Customer By Code godoc
// @Description Get Customer By Code
// @Tags		Customer
// @Param		code  path      string  true  "Customer Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/customer/code/{code} [get]
func (h CustomerHttp) InfoCustomerByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")

	doc, err := h.svc.InfoCustomerByCode(shopID, code)

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

// List Customer godoc
// @Description get struct array by ID
// @Tags		Customer
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "page"
// @Param		limit	query	integer		false  "limit"
// @Param		iscreditor	query	bool		false  "creditor"
// @Param		isdebtor	query	bool		false  "debtor"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/customer [get]
func (h CustomerHttp) SearchCustomerPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	filters := map[string]interface{}{}
	iscreditor := ctx.QueryParam("iscreditor")
	isdebtor := ctx.QueryParam("isdebtor")

	if iscreditor != "" && iscreditor != "-1" {
		filters["iscreditor"] = iscreditor == "true" || iscreditor == "1"
	}

	if isdebtor != "" && isdebtor != "-1" {
		filters["isdebtor"] = isdebtor == "true" || isdebtor == "1"
	}

	docList, pagination, err := h.svc.SearchCustomer(shopID, filters, pageable)

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

// List Customer godoc
// @Description search limit offset
// @Tags		Customer
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Param		iscreditor	query	bool		false  "creditor"
// @Param		isdebtor	query	bool		false  "debtor"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/customer/list [get]
func (h CustomerHttp) SearchCustomerStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	filters := map[string]interface{}{}
	iscreditor := ctx.QueryParam("iscreditor")
	isdebtor := ctx.QueryParam("isdebtor")

	if iscreditor != "" && iscreditor != "-1" {
		filters["iscreditor"] = iscreditor == "true" || iscreditor == "1"
	}

	if isdebtor != "" && isdebtor != "-1" {
		filters["isdebtor"] = isdebtor == "true" || isdebtor == "1"
	}

	docList, total, err := h.svc.SearchCustomerStep(shopID, lang, filters, pageableStep)

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

// Create Customer Bulk godoc
// @Description Create Customer
// @Tags		Customer
// @Param		Customer  body      []models.CustomerRequest  true  "Customer"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/customer/bulk [post]
func (h CustomerHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.CustomerRequest{}
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
