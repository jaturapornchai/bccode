package ordertype

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/product/ordertype/models"
	"smlcloudplatform/internal/product/ordertype/repositories"
	"smlcloudplatform/internal/product/ordertype/services"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
)

type IOrderTypeHttp interface{}

type OrderTypeHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IOrderTypeHttpService
}

func NewOrderTypeHttp(ms *microservice.Microservice, cfg config.IConfig) OrderTypeHttp {
	prod := ms.Producer(cfg.MQConfig())
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewOrderTypeRepository(pst)
	repoMessageQueue := repositories.NewOrderTypeMessageQueueRepository(prod)
	repoProductBarcode := productbarcode_repositories.NewProductBarcodeRepository(pst, cache)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewOrderTypeHttpService(repo, repoMessageQueue, repoProductBarcode, masterSyncCacheRepo)

	return OrderTypeHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h OrderTypeHttp) RegisterHttp() {

	h.ms.POST("/product/order-type/bulk", h.SaveBulk)

	h.ms.GET("/product/order-type", h.SearchOrderTypePage)
	h.ms.GET("/product/order-type/list", h.SearchOrderTypeStep)
	h.ms.POST("/product/order-type", h.CreateOrderType)
	h.ms.GET("/product/order-type/:id", h.InfoOrderType)
	h.ms.GET("/product/order-type/code/:code", h.InfoOrderTypeByCode)
	h.ms.PUT("/product/order-type/:id", h.UpdateOrderType)
	h.ms.DELETE("/product/order-type/:id", h.DeleteOrderType)
	h.ms.DELETE("/product/order-type", h.DeleteOrderTypeByGUIDs)
}

// Create OrderType godoc
// @Description Create OrderType
// @Tags		OrderType
// @Param		OrderType  body      models.OrderType  true  "OrderType"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/order-type [post]
func (h OrderTypeHttp) CreateOrderType(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.OrderType{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateOrderType(shopID, authUsername, *docReq)

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

// Update OrderType godoc
// @Description Update OrderType
// @Tags		OrderType
// @Param		id  path      string  true  "OrderType ID"
// @Param		OrderType  body      models.OrderType  true  "OrderType"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/order-type/{id} [put]
func (h OrderTypeHttp) UpdateOrderType(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.OrderType{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateOrderType(shopID, id, authUsername, *docReq)

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

// Delete OrderType godoc
// @Description Delete OrderType
// @Tags		OrderType
// @Param		id  path      string  true  "OrderType ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/order-type/{id} [delete]
func (h OrderTypeHttp) DeleteOrderType(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteOrderType(shopID, id, authUsername)

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

// Delete OrderType godoc
// @Description Delete OrderType
// @Tags		OrderType
// @Param		OrderType  body      []string  true  "OrderType GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/order-type [delete]
func (h OrderTypeHttp) DeleteOrderTypeByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteOrderTypeByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get OrderType godoc
// @Description get OrderType info by guidfixed
// @Tags		OrderType
// @Param		id  path      string  true  "OrderType guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/order-type/{id} [get]
func (h OrderTypeHttp) InfoOrderType(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get OrderType %v", id)
	doc, err := h.svc.InfoOrderType(shopID, id)

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

// Get OrderType By Code godoc
// @Description get OrderType info by Code
// @Tags		OrderType
// @Param		code  path      string  true  "OrderType Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/order-type/code/{code} [get]
func (h OrderTypeHttp) InfoOrderTypeByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")

	doc, err := h.svc.InfoOrderTypeByCode(shopID, code)

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

// List OrderType step godoc
// @Description get list step
// @Tags		OrderType
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/order-type [get]
func (h OrderTypeHttp) SearchOrderTypePage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.svc.SearchOrderType(shopID, map[string]interface{}{}, pageable)

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

// List OrderType godoc
// @Description search limit offset
// @Tags		OrderType
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/order-type/list [get]
func (h OrderTypeHttp) SearchOrderTypeStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	docList, total, err := h.svc.SearchOrderTypeStep(shopID, lang, pageableStep)

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

// Create OrderType Bulk godoc
// @Description Create OrderType
// @Tags		OrderType
// @Param		OrderType  body      []models.OrderType  true  "OrderType"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/order-type/bulk [post]
func (h OrderTypeHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.OrderType{}
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
