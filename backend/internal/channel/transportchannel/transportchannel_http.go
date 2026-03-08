package transportchannel

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/channel/transportchannel/models"
	"smlcloudplatform/internal/channel/transportchannel/repositories"
	"smlcloudplatform/internal/channel/transportchannel/services"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
)

type ITransportChannelHttp interface{}

type TransportChannelHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.ITransportChannelHttpService
}

func NewTransportChannelHttp(ms *microservice.Microservice, cfg config.IConfig) TransportChannelHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewTransportChannelRepository(pst)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewTransportChannelHttpService(repo, masterSyncCacheRepo)

	return TransportChannelHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h TransportChannelHttp) RegisterHttp() {

	h.ms.POST("/transport-channel/bulk", h.SaveBulk)

	h.ms.GET("/transport-channel", h.SearchTransportChannelPage)
	h.ms.GET("/transport-channel/list", h.SearchTransportChannelStep)
	h.ms.POST("/transport-channel", h.CreateTransportChannel)
	h.ms.GET("/transport-channel/:id", h.InfoTransportChannel)
	h.ms.GET("/transport-channel/code/:code", h.InfoTransportChannelByCode)
	h.ms.PUT("/transport-channel/:id", h.UpdateTransportChannel)
	h.ms.DELETE("/transport-channel/:id", h.DeleteTransportChannel)
	h.ms.DELETE("/transport-channel", h.DeleteTransportChannelByGUIDs)
}

// Create TransportChannel godoc
// @Description Create TransportChannel
// @Tags		TransportChannel
// @Param		TransportChannel  body      models.TransportChannel  true  "TransportChannel"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transport-channel [post]
func (h TransportChannelHttp) CreateTransportChannel(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.TransportChannel{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateTransportChannel(shopID, authUsername, *docReq)

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

// Update TransportChannel godoc
// @Description Update TransportChannel
// @Tags		TransportChannel
// @Param		id  path      string  true  "TransportChannel ID"
// @Param		TransportChannel  body      models.TransportChannel  true  "TransportChannel"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transport-channel/{id} [put]
func (h TransportChannelHttp) UpdateTransportChannel(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.TransportChannel{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateTransportChannel(shopID, id, authUsername, *docReq)

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

// Delete TransportChannel godoc
// @Description Delete TransportChannel
// @Tags		TransportChannel
// @Param		id  path      string  true  "TransportChannel ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transport-channel/{id} [delete]
func (h TransportChannelHttp) DeleteTransportChannel(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteTransportChannel(shopID, id, authUsername)

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

// Delete TransportChannel godoc
// @Description Delete TransportChannel
// @Tags		TransportChannel
// @Param		TransportChannel  body      []string  true  "TransportChannel GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transport-channel [delete]
func (h TransportChannelHttp) DeleteTransportChannelByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteTransportChannelByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get TransportChannel godoc
// @Description get TransportChannel info by guidfixed
// @Tags		TransportChannel
// @Param		id  path      string  true  "TransportChannel guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transport-channel/{id} [get]
func (h TransportChannelHttp) InfoTransportChannel(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get TransportChannel %v", id)
	doc, err := h.svc.InfoTransportChannel(shopID, id)

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

// Get TransportChannel By Code godoc
// @Description get TransportChannel info by Code
// @Tags		TransportChannel
// @Param		code  path      string  true  "TransportChannel Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transport-channel/code/{code} [get]
func (h TransportChannelHttp) InfoTransportChannelByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")

	doc, err := h.svc.InfoTransportChannelByCode(shopID, code)

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

// List TransportChannel step godoc
// @Description get list step
// @Tags		TransportChannel
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transport-channel [get]
func (h TransportChannelHttp) SearchTransportChannelPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.svc.SearchTransportChannel(shopID, map[string]interface{}{}, pageable)

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

// List TransportChannel godoc
// @Description search limit offset
// @Tags		TransportChannel
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transport-channel/list [get]
func (h TransportChannelHttp) SearchTransportChannelStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	docList, total, err := h.svc.SearchTransportChannelStep(shopID, lang, pageableStep)

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

// Create TransportChannel Bulk godoc
// @Description Create TransportChannel
// @Tags		TransportChannel
// @Param		TransportChannel  body      []models.TransportChannel  true  "TransportChannel"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transport-channel/bulk [post]
func (h TransportChannelHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.TransportChannel{}
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
