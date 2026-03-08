package storefront

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/storefront/models"
	"smlcloudplatform/internal/storefront/repositories"
	"smlcloudplatform/internal/storefront/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
)

type IStorefrontHttp interface{}

type StorefrontHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IStorefrontHttpService
}

func NewStorefrontHttp(ms *microservice.Microservice, cfg config.IConfig) StorefrontHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())

	repo := repositories.NewStorefrontRepository(pst)

	svc := services.NewStorefrontHttpService(repo)

	return StorefrontHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h StorefrontHttp) RegisterHttp() {

	h.ms.GET("/storefront", h.SearchStorefront)
	h.ms.POST("/storefront", h.CreateStorefront)
	h.ms.GET("/storefront/:id", h.InfoStorefront)
	h.ms.PUT("/storefront/:id", h.UpdateStorefront)
	h.ms.DELETE("/storefront/:id", h.DeleteStorefront)
}

// Create Storefront godoc
// @Description Create Storefront
// @Tags		Storefront
// @Param		Storefront  body      models.Storefront  true  "Storefront"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /storefront [post]
func (h StorefrontHttp) CreateStorefront(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.Storefront{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateStorefront(shopID, authUsername, *docReq)

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

// Update Storefront godoc
// @Description Update Storefront
// @Tags		Storefront
// @Param		id  path      string  true  "Storefront ID"
// @Param		Storefront  body      models.Storefront  true  "Storefront"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /storefront/{id} [put]
func (h StorefrontHttp) UpdateStorefront(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.Storefront{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateStorefront(shopID, id, authUsername, *docReq)

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

// Delete Storefront godoc
// @Description Delete Storefront
// @Tags		Storefront
// @Param		id  path      string  true  "Storefront ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /storefront/{id} [delete]
func (h StorefrontHttp) DeleteStorefront(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteStorefront(shopID, id, authUsername)

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

// Get Storefront godoc
// @Description get struct array by ID
// @Tags		Storefront
// @Param		id  path      string  true  "Storefront ID"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /storefront/{id} [get]
func (h StorefrontHttp) InfoStorefront(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get Storefront %v", id)
	doc, err := h.svc.InfoStorefront(shopID, id)

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

// List Storefront godoc
// @Description get struct array by ID
// @Tags		Storefront
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /storefront [get]
func (h StorefrontHttp) SearchStorefront(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)
	docList, pagination, err := h.svc.SearchStorefront(shopID, pageable)

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
