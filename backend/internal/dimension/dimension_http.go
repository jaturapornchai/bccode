package dimension

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/dimension/models"
	"smlcloudplatform/internal/dimension/repositories"
	"smlcloudplatform/internal/dimension/services"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IDimensionHttp interface{}

type DimensionHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IDimensionHttpService
}

func NewDimensionHttp(ms *microservice.Microservice, cfg config.IConfig) DimensionHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewDimensionRepository(pst)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewDimensionHttpService(repo, masterSyncCacheRepo, 15*time.Second)

	return DimensionHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h DimensionHttp) RegisterHttp() {

	h.ms.GET("/dimension", h.SearchDimensionPage)
	h.ms.GET("/dimension/list", h.SearchDimensionStep)
	h.ms.POST("/dimension", h.CreateDimension)
	h.ms.GET("/dimension/:id", h.InfoDimension)
	h.ms.PUT("/dimension/:id", h.UpdateDimension)
	h.ms.DELETE("/dimension/:id", h.DeleteDimension)
	h.ms.DELETE("/dimension", h.DeleteDimensionByGUIDs)
}

// Create Dimension godoc
// @Description Create Dimension
// @Tags		Dimension
// @Param		Dimension  body      models.Dimension  true  "Dimension"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /dimension [post]
func (h DimensionHttp) CreateDimension(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.Dimension{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateDimension(shopID, authUsername, *docReq)

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

// Update Dimension godoc
// @Description Update Dimension
// @Tags		Dimension
// @Param		id  path      string  true  "Dimension ID"
// @Param		Dimension  body      models.Dimension  true  "Dimension"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /dimension/{id} [put]
func (h DimensionHttp) UpdateDimension(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.Dimension{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateDimension(shopID, id, authUsername, *docReq)

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

// Delete Dimension godoc
// @Description Delete Dimension
// @Tags		Dimension
// @Param		id  path      string  true  "Dimension ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /dimension/{id} [delete]
func (h DimensionHttp) DeleteDimension(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteDimension(shopID, id, authUsername)

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

// Delete Dimension godoc
// @Description Delete Dimension
// @Tags		Dimension
// @Param		Dimension  body      []string  true  "Dimension GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /dimension [delete]
func (h DimensionHttp) DeleteDimensionByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteDimensionByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get Dimension godoc
// @Description get Dimension info by guidfixed
// @Tags		Dimension
// @Param		id  path      string  true  "Dimension guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /dimension/{id} [get]
func (h DimensionHttp) InfoDimension(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get Dimension %v", id)
	doc, err := h.svc.InfoDimension(shopID, id)

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

// List Dimension step godoc
// @Description get list step
// @Tags		Dimension
// @Param		disabled		query	boolean		false  "disabled"
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /dimension [get]
func (h DimensionHttp) SearchDimensionPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "disabled",
			Field: "isdisabled",
			Type:  requestfilter.FieldTypeBoolean,
		},
	})

	docList, pagination, err := h.svc.SearchDimension(shopID, filters, pageable)

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

// List Dimension godoc
// @Description search limit offset
// @Tags		Dimension
// @Param		disabled		query	boolean		false  "disabled"
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /dimension/list [get]
func (h DimensionHttp) SearchDimensionStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "isdisabled",
			Field: "isdisabled",
			Type:  requestfilter.FieldTypeBoolean,
		},
	})

	docList, total, err := h.svc.SearchDimensionStep(shopID, lang, filters, pageableStep)

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
