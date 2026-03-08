package optionpattern

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/product/optionpattern/models"
	"smlcloudplatform/internal/product/optionpattern/repositories"
	"smlcloudplatform/internal/product/optionpattern/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
)

type IOptionPatternHttp interface{}

type OptionPatternHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IOptionPatternHttpService
}

func NewOptionPatternHttp(ms *microservice.Microservice, cfg config.IConfig) OptionPatternHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())

	repo := repositories.NewOptionPatternRepository(pst)

	svc := services.NewOptionPatternHttpService(repo)

	return OptionPatternHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h OptionPatternHttp) RegisterHttp() {

	h.ms.POST("/optionpattern/bulk", h.SaveBulk)

	h.ms.GET("/optionpattern", h.SearchOptionPattern)
	h.ms.POST("/optionpattern", h.CreateOptionPattern)
	h.ms.GET("/optionpattern/:id", h.InfoOptionPattern)
	h.ms.PUT("/optionpattern/:id", h.UpdateOptionPattern)
	h.ms.DELETE("/optionpattern/:id", h.DeleteOptionPattern)
}

// Create OptionPattern godoc
// @Description Create OptionPattern
// @Tags		OptionPattern
// @Param		OptionPattern  body      models.OptionPattern  true  "OptionPattern"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /optionpattern [post]
func (h OptionPatternHttp) CreateOptionPattern(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.OptionPattern{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateOptionPattern(shopID, authUsername, *docReq)

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

// Update OptionPattern godoc
// @Description Update OptionPattern
// @Tags		OptionPattern
// @Param		id  path      string  true  "OptionPattern ID"
// @Param		OptionPattern  body      models.OptionPattern  true  "OptionPattern"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /optionpattern/{id} [put]
func (h OptionPatternHttp) UpdateOptionPattern(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.OptionPattern{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateOptionPattern(shopID, id, authUsername, *docReq)

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

// Delete OptionPattern godoc
// @Description Delete OptionPattern
// @Tags		OptionPattern
// @Param		id  path      string  true  "OptionPattern ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /optionpattern/{id} [delete]
func (h OptionPatternHttp) DeleteOptionPattern(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteOptionPattern(shopID, id, authUsername)

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

// Get OptionPattern godoc
// @Description get struct array by ID
// @Tags		OptionPattern
// @Param		id  path      string  true  "OptionPattern ID"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /optionpattern/{id} [get]
func (h OptionPatternHttp) InfoOptionPattern(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get OptionPattern %v", id)
	doc, err := h.svc.InfoOptionPattern(shopID, id)

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

// List OptionPattern godoc
// @Description get struct array by ID
// @Tags		OptionPattern
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /optionpattern [get]
func (h OptionPatternHttp) SearchOptionPattern(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)
	docList, pagination, err := h.svc.SearchOptionPattern(shopID, pageable)

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

// Create OptionPattern Bulk godoc
// @Description Create OptionPattern
// @Tags		OptionPattern
// @Param		OptionPattern  body      []models.OptionPattern  true  "OptionPattern"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /optionpattern/bulk [post]
func (h OptionPatternHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.OptionPattern{}
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
