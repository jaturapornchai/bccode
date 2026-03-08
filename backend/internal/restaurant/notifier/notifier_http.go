package notifier

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/restaurant/notifier/models"
	"smlcloudplatform/internal/restaurant/notifier/repositories"
	"smlcloudplatform/internal/restaurant/notifier/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type INotifierHttp interface{}

type NotifierHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.INotifierHttpService
}

func NewNotifierHttp(ms *microservice.Microservice, cfg config.IConfig) NotifierHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())

	repo := repositories.NewNotifierRepository(pst)

	svc := services.NewNotifierHttpService(repo, utils.RandStringBytesMaskImprSrcUnsafe, utils.RandNumberX, 15*time.Second)

	return NotifierHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h NotifierHttp) RegisterHttp() {

	h.ms.GET("/restaurant/notifier", h.SearchNotifierPage)
	h.ms.POST("/restaurant/notifier", h.CreateNotifier)
	h.ms.GET("/restaurant/notifier/:id", h.InfoNotifier)
	h.ms.GET("/restaurant/notifier/code/:code", h.InfoNotifierByCode)
	h.ms.PUT("/restaurant/notifier/:id", h.UpdateNotifier)
	h.ms.DELETE("/restaurant/notifier/:id", h.DeleteNotifier)
	h.ms.DELETE("/restaurant/notifier", h.DeleteNotifierByGUIDs)
}

// Create Notifier godoc
// @Description Create Notifier
// @Tags		Notifier
// @Param		Notifier  body      models.Notifier  true  "Notifier"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /restaurant/notifier [post]
func (h NotifierHttp) CreateNotifier(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.Notifier{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateNotifier(shopID, authUsername, *docReq)

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

// Update Notifier godoc
// @Description Update Notifier
// @Tags		Notifier
// @Param		id  path      string  true  "Notifier ID"
// @Param		Notifier  body      models.Notifier  true  "Notifier"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /restaurant/notifier/{id} [put]
func (h NotifierHttp) UpdateNotifier(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.Notifier{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateNotifier(shopID, id, authUsername, *docReq)

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
// @Router /restaurant/notifier/{id} [delete]
func (h NotifierHttp) DeleteNotifier(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteNotifier(shopID, id, authUsername)

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
// @Router /restaurant/notifier [delete]
func (h NotifierHttp) DeleteNotifierByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteNotifierByGUIDs(shopID, authUsername, docReq)

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
// @Router /restaurant/notifier/{id} [get]
func (h NotifierHttp) InfoNotifier(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get Notifier %v", id)
	doc, err := h.svc.InfoNotifier(shopID, id)

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

// Get Notifier By Code godoc
// @Description get Notifier info by Code
// @Tags		Notifier
// @Param		code  path      string  true  "Notifier Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /restaurant/notifier/code/{code} [get]
func (h NotifierHttp) InfoNotifierByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")

	doc, err := h.svc.InfoNotifierByCode(shopID, code)

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
// @Router /restaurant/notifier [get]
func (h NotifierHttp) SearchNotifierPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.svc.SearchNotifier(shopID, map[string]interface{}{}, pageable)

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
