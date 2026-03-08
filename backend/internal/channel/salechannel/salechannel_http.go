package salechannel

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/channel/salechannel/models"
	"smlcloudplatform/internal/channel/salechannel/repositories"
	"smlcloudplatform/internal/channel/salechannel/services"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
)

type ISaleChannelHttp interface{}

type SaleChannelHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.ISaleChannelHttpService
}

func NewSaleChannelHttp(ms *microservice.Microservice, cfg config.IConfig) SaleChannelHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewSaleChannelRepository(pst)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewSaleChannelHttpService(repo, masterSyncCacheRepo)

	return SaleChannelHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h SaleChannelHttp) RegisterHttp() {

	h.ms.POST("/sale-channel/bulk", h.SaveBulk)

	h.ms.GET("/sale-channel", h.SearchSaleChannelPage)
	h.ms.GET("/sale-channel/list", h.SearchSaleChannelStep)
	h.ms.POST("/sale-channel", h.CreateSaleChannel)
	h.ms.GET("/sale-channel/:id", h.InfoSaleChannel)
	h.ms.GET("/sale-channel/code/:code", h.InfoSaleChannelByCode)
	h.ms.PUT("/sale-channel/:id", h.UpdateSaleChannel)
	h.ms.DELETE("/sale-channel/:id", h.DeleteSaleChannel)
	h.ms.DELETE("/sale-channel", h.DeleteSaleChannelByGUIDs)
}

// Create SaleChannel godoc
// @Description Create SaleChannel
// @Tags		SaleChannel
// @Param		SaleChannel  body      models.SaleChannel  true  "SaleChannel"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /sale-channel [post]
func (h SaleChannelHttp) CreateSaleChannel(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.SaleChannel{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateSaleChannel(shopID, authUsername, *docReq)

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

// Update SaleChannel godoc
// @Description Update SaleChannel
// @Tags		SaleChannel
// @Param		id  path      string  true  "SaleChannel ID"
// @Param		SaleChannel  body      models.SaleChannel  true  "SaleChannel"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /sale-channel/{id} [put]
func (h SaleChannelHttp) UpdateSaleChannel(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.SaleChannel{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateSaleChannel(shopID, id, authUsername, *docReq)

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

// Delete SaleChannel godoc
// @Description Delete SaleChannel
// @Tags		SaleChannel
// @Param		id  path      string  true  "SaleChannel ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /sale-channel/{id} [delete]
func (h SaleChannelHttp) DeleteSaleChannel(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteSaleChannel(shopID, id, authUsername)

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

// Delete SaleChannel godoc
// @Description Delete SaleChannel
// @Tags		SaleChannel
// @Param		SaleChannel  body      []string  true  "SaleChannel GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /sale-channel [delete]
func (h SaleChannelHttp) DeleteSaleChannelByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteSaleChannelByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get SaleChannel godoc
// @Description get SaleChannel info by guidfixed
// @Tags		SaleChannel
// @Param		id  path      string  true  "SaleChannel guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /sale-channel/{id} [get]
func (h SaleChannelHttp) InfoSaleChannel(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get SaleChannel %v", id)
	doc, err := h.svc.InfoSaleChannel(shopID, id)

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

// Get SaleChannel By Code godoc
// @Description get SaleChannel info by Code
// @Tags		SaleChannel
// @Param		code  path      string  true  "SaleChannel Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /sale-channel/code/{code} [get]
func (h SaleChannelHttp) InfoSaleChannelByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")

	doc, err := h.svc.InfoSaleChannelByCode(shopID, code)

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

// List SaleChannel step godoc
// @Description get list step
// @Tags		SaleChannel
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /sale-channel [get]
func (h SaleChannelHttp) SearchSaleChannelPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.svc.SearchSaleChannel(shopID, map[string]interface{}{}, pageable)

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

// List SaleChannel godoc
// @Description search limit offset
// @Tags		SaleChannel
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /sale-channel/list [get]
func (h SaleChannelHttp) SearchSaleChannelStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	docList, total, err := h.svc.SearchSaleChannelStep(shopID, lang, pageableStep)

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

// Create SaleChannel Bulk godoc
// @Description Create SaleChannel
// @Tags		SaleChannel
// @Param		SaleChannel  body      []models.SaleChannel  true  "SaleChannel"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /sale-channel/bulk [post]
func (h SaleChannelHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.SaleChannel{}
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
