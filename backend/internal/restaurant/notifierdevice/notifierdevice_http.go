package notifierdevice

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/restaurant/notifierdevice/models"
	"smlcloudplatform/internal/restaurant/notifierdevice/repositories"
	"smlcloudplatform/internal/restaurant/notifierdevice/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type INotifierDeviceHttp interface{}

type NotifierDeviceHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.INotifierDeviceHttpService
}

func NewNotifierDeviceHttp(ms *microservice.Microservice, cfg config.IConfig) NotifierDeviceHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewNotifierDeviceRepository(pst)
	cacheRepo := repositories.NewNotifierDeviceCacheRepository(cache)

	svc := services.NewNotifierDeviceHttpService(repo, cacheRepo, utils.RandStringBytesMaskImprSrcUnsafe, utils.RandNumberX, 15*time.Second)

	return NotifierDeviceHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h NotifierDeviceHttp) RegisterHttp() {

	h.ms.POST("/restaurant/notifier-device/ref", h.CreateNotifierAuthRefCode)
	h.ms.POST("/restaurant/notifier-device/ref-confirm", h.ConfirmNotifierAuthRefCode)

	h.ms.GET("/restaurant/notifier-device", h.SearchNotifierPage)
	h.ms.GET("/restaurant/notifier-device/:id", h.InfoNotifierDevice)
	h.ms.PUT("/restaurant/notifier-device/:id", h.UpdateNotifierDevice)
	h.ms.DELETE("/restaurant/notifier-device/:id", h.DeleteNotifierDevice)
	h.ms.DELETE("/restaurant/notifier-device", h.DeleteNotifierDeviceByGUIDs)
}

// Ref Notifier godoc
// @Description Ref Notifier
// @Tags		Notifier
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /restaurant/notifier-device/ref [post]
func (h NotifierDeviceHttp) CreateNotifierAuthRefCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	notifierAuth, err := h.svc.CreateAuthCode(shopID, authUsername)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, map[string]interface{}{
		"success": true,
		"refcode": notifierAuth.RefCode,
	})
	return nil
}

// Ref confirm Notifier godoc
// @Description Ref confirm Notifier
// @Tags		Notifier
// @Param		NotifierDeviceConfirmAuthPayload  body      models.NotifierDeviceConfirmAuthPayload  true  "NotifierDeviceConfirmAuthPayload"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /restaurant/notifier-device/ref-confirm [post]
func (h NotifierDeviceHttp) ConfirmNotifierAuthRefCode(ctx microservice.IContext) error {

	input := ctx.ReadInput()

	docReq := &models.NotifierDeviceConfirmAuthPayload{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	isAuthPass, err := h.svc.ConfirmAuthCode(*docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, map[string]interface{}{
		"success": true,
		"pass":    isAuthPass,
	})
	return nil
}

// Update Notifier godoc
// @Description Update Notifier
// @Tags		Notifier
// @Param		id  path      string  true  "Notifier ID"
// @Param		Notifier  body      models.Notifier  true  "Notifier"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /restaurant/notifier-device/{id} [put]
func (h NotifierDeviceHttp) UpdateNotifierDevice(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.NotifierDevice{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateNotifierDevice(shopID, id, authUsername, *docReq)

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

// Delete Notifier godoc
// @Description Delete Notifier
// @Tags		Notifier
// @Param		id  path      string  true  "Notifier ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /restaurant/notifier-device/{id} [delete]
func (h NotifierDeviceHttp) DeleteNotifierDevice(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteNotifierDevice(shopID, id, authUsername)

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

// Delete Notifier godoc
// @Description Delete Notifier
// @Tags		Notifier
// @Param		Notifier  body      []string  true  "Notifier GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /restaurant/notifier-device [delete]
func (h NotifierDeviceHttp) DeleteNotifierDeviceByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteNotifierDeviceByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get Notifier godoc
// @Description get Notifier info by guidfixed
// @Tags		Notifier
// @Param		id  path      string  true  "Notifier guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /restaurant/notifier-device/{id} [get]
func (h NotifierDeviceHttp) InfoNotifierDevice(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get Notifier %v", id)
	doc, err := h.svc.InfoNotifierDevice(shopID, id)

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

// List Notifier step godoc
// @Description get list step
// @Tags		Notifier
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /restaurant/notifier-device [get]
func (h NotifierDeviceHttp) SearchNotifierPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.svc.SearchNotifierDevice(shopID, map[string]interface{}{}, pageable)

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
