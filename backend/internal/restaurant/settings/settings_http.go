package settings

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/restaurant/settings/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"

	mastersync "smlcloudplatform/internal/mastersync/repositories"
)

type IRestaurantSettingsHttp interface{}

type RestaurantSettingsHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc IRestaurantSettingsService
}

func NewRestaurantSettingsHttp(ms *microservice.Microservice, cfg config.IConfig) RestaurantSettingsHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := NewRestaurantSettingsRepository(pst)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := NewRestaurantSettingsService(repo, masterSyncCacheRepo)

	return RestaurantSettingsHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h RestaurantSettingsHttp) RegisterHttp() {

	h.ms.POST("/restaurant/settings/bulk", h.SaveBulk)

	h.ms.GET("/restaurant/settings", h.SearchRestaurantSettings)
	h.ms.POST("/restaurant/settings", h.CreateRestaurantSettings)
	h.ms.GET("/restaurant/settings/:id", h.InfoRestaurantSettings)
	h.ms.GET("/restaurant/settings/code/:code", h.InfoRestaurantSettingsByCode)
	h.ms.PUT("/restaurant/settings/:id", h.UpdateRestaurantSettings)
	h.ms.DELETE("/restaurant/settings/:id", h.DeleteRestaurantSettings)
	h.ms.DELETE("/restaurant/settings", h.DeleteByGUIDs)

}

// Create Restaurant Settings godoc
// @Description Restaurant Settings
// @Tags		Restaurant
// @Param		Settings  body      models.RestaurantSettings  true  "Settings"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /restaurant/settings [post]
func (h RestaurantSettingsHttp) CreateRestaurantSettings(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.RestaurantSettings{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateRestaurantSettings(shopID, authUsername, *docReq)

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

// Update Restaurant Settings godoc
// @Description Restaurant Settings
// @Tags		Restaurant
// @Param		id  path      string  true  "Settings ID"
// @Param		Settings  body      models.RestaurantSettings  true  "Settings"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /restaurant/settings/{id} [put]
func (h RestaurantSettingsHttp) UpdateRestaurantSettings(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.RestaurantSettings{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateRestaurantSettings(shopID, id, authUsername, *docReq)

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

// Delete Restaurant Settings godoc
// @Description Restaurant Settings
// @Tags		Restaurant
// @Param		id  path      string  true  "RestaurantSettings ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /restaurant/settings/{id} [delete]
func (h RestaurantSettingsHttp) DeleteRestaurantSettings(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteRestaurantSettings(shopID, id, authUsername)

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

// Delete Restaurant Settings godoc
// @Description Delete Restaurant Settings
// @Tags		Restaurant
// @Param		RestaurantSettings  body      []string  true  "Restaurant Settings GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /restaurant/settings [delete]
func (h RestaurantSettingsHttp) DeleteByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get Restaurant Settings Infomation godoc
// @Description Get Restaurant Settings
// @Tags		Restaurant
// @Param		id  path      string  true  "RestaurantSettings Id"
// @Accept 		json
// @Success		200	{object}	models.RestaurantSettingsInfoResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /restaurant/settings/{id} [get]
func (h RestaurantSettingsHttp) InfoRestaurantSettings(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get RestaurantSettings %v", id)
	doc, err := h.svc.InfoRestaurantSettings(shopID, id)

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

// Get Restaurant Settings By Code Infomation godoc
// @Description Get Restaurant Settings By Code
// @Tags		Restaurant
// @Param		code  path      string  true  "RestaurantSettings Code"
// @Accept 		json
// @Success		200	{object}	models.RestaurantSettingsInfoResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /restaurant/settings/code/{code} [get]
func (h RestaurantSettingsHttp) InfoRestaurantSettingsByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")

	h.ms.Logger.Debugf("Get RestaurantSettings %v", code)
	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.svc.ListRestaurantSettingsByCode(shopID, code, pageable)

	if err != nil {
		h.ms.Logger.Errorf("Error getting document %v: %v", code, err)
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

// List Restaurant Settings godoc
// @Description List Restaurant Settings Category
// @Tags		Restaurant
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Size"
// @Accept 		json
// @Success		200	{object}	models.RestaurantSettingsPageResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /restaurant/settings [get]
func (h RestaurantSettingsHttp) SearchRestaurantSettings(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.svc.SearchRestaurantSettings(shopID, pageable)

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

// Create RestaurantSettings Bulk godoc
// @Description Create RestaurantSettings
// @Tags		Restaurant
// @Param		RestaurantSettings  body      []models.RestaurantSettings  true  "RestaurantSettings"
// @Accept 		json
// @Success		201	{object}	common.BulkInsertResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /restaurant/settings/bulk [post]
func (h RestaurantSettingsHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.RestaurantSettings{}
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
