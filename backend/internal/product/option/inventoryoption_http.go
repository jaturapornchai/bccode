package option

import (
	"encoding/json"
	"net/http"
	"net/url"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/product/option/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
)

type OptionHttp struct {
	ms         *microservice.Microservice
	cfg        config.IConfig
	optService IOptionService
}

func NewOptionHttp(ms *microservice.Microservice, cfg config.IConfig) *OptionHttp {

	pst := ms.MongoPersister(cfg.MongoPersisterConfig())

	optRepo := NewOptionRepository(pst)
	optService := NewOptionService(optRepo)

	return &OptionHttp{
		ms:         ms,
		cfg:        cfg,
		optService: optService,
	}
}

func (h OptionHttp) RegisterHttp() {

	h.ms.GET("/option/:id", h.InfoInventoryOptionMain)
	h.ms.GET("/option/by-code", h.InfoArray)
	h.ms.GET("/option", h.SearchInventoryOptionMain)
	h.ms.POST("/option", h.CreateInventoryOptionMain)
	h.ms.PUT("/option/:id", h.UpdateInventoryOptionMain)
	h.ms.DELETE("/option/:id", h.DeleteInventoryOptionMain)

}

// Create Inventory Option godoc
// @Description Create Inventory Option
// @Tags		Inventory
// @Param		Option  body      models.InventoryOptionMain  true  "Option"
// @Accept 		json
// @Success		201	{object}	models.ResponseSuccessWithID
// @Failure		401 {object}	models.AuthResponseFailed
// @Security     AccessToken
// @Router /option [post]
func (h *OptionHttp) CreateInventoryOptionMain(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.InventoryOptionMain{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.optService.CreateOption(shopID, authUsername, *docReq)

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

// Update Option godoc
// @Description Update Option
// @Tags		Inventory
// @Param		id  path      string  true  "Option ID"
// @Param		Option  body      models.InventoryOptionMain  true  "Option"
// @Accept 		json
// @Success		200	{object}	models.ResponseSuccessWithID
// @Failure		401 {object}	models.AuthResponseFailed
// @Security     AccessToken
// @Router /option/{id} [put]
func (h *OptionHttp) UpdateInventoryOptionMain(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.InventoryOptionMain{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.optService.UpdateOption(shopID, id, authUsername, *docReq)

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

// Delete Option godoc
// @Description Delete Option
// @Tags		Inventory
// @Param		id  path      string  true  "Option ID"
// @Accept 		json
// @Success		200	{object}	models.ResponseSuccessWithID
// @Failure		401 {object}	models.AuthResponseFailed
// @Security     AccessToken
// @Router /option/{id} [delete]
func (h *OptionHttp) DeleteInventoryOptionMain(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.optService.DeleteOption(shopID, id, authUsername)

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

// Get Inventory Option Infomation godoc
// @Description Get Inventory Option
// @Tags		Inventory
// @Param		id  path      string  true  "Option Id"
// @Accept 		json
// @Success		200	{object}	models.InventoryOptionMainInfo
// @Failure		401 {object}	models.AuthResponseFailed
// @Security     AccessToken
// @Router /option/{id} [get]
func (h *OptionHttp) InfoInventoryOptionMain(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	doc, err := h.optService.InfoOption(shopID, id)

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

// Get Inventory Option By code array godoc
// @Description get Inventory Option by code array
// @Tags		Unit
// @Param		codes	query	string		false  "Code filter, json array encode "
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /option/by-code [get]
func (h OptionHttp) InfoArray(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	codesReq, err := url.QueryUnescape(ctx.QueryParam("codes"))

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	docReq := []string{}
	err = json.Unmarshal([]byte(codesReq), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	// where to filter array
	doc, err := h.optService.InfoWTFArray(shopID, docReq)

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

// List Inventory Option godoc
// @Description List Inventory Option
// @Tags		Inventory
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{object}	models.InventoryOptionPageResponse
// @Failure		401 {object}	models.AuthResponseFailed
// @Security     AccessToken
// @Router /option [get]
func (h *OptionHttp) SearchInventoryOptionMain(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.optService.SearchOption(shopID, pageable)

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
