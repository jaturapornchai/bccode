package zone

import (
	"encoding/json"
	"net/http"
	"net/url"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/restaurant/zone/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"

	mastersync "smlcloudplatform/internal/mastersync/repositories"
)

type IZoneHttp interface{}

type ZoneHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc IZoneService
}

func NewZoneHttp(ms *microservice.Microservice, cfg config.IConfig) ZoneHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := NewZoneRepository(pst)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := NewZoneService(repo, masterSyncCacheRepo)

	return ZoneHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h ZoneHttp) RegisterHttp() {

	h.ms.POST("/restaurant/zone/bulk", h.SaveBulk)

	h.ms.GET("/restaurant/zone", h.SearchZone)
	h.ms.POST("/restaurant/zone", h.CreateZone)
	h.ms.GET("/restaurant/zone/:id", h.InfoZone)
	h.ms.GET("/restaurant/zone/by-code", h.InfoArray)
	h.ms.PUT("/restaurant/zone/:id", h.UpdateZone)
	h.ms.DELETE("/restaurant/zone/:id", h.DeleteZone)
	h.ms.DELETE("/restaurant/zone", h.DeleteByGUIDs)

}

// Create Restaurant Zone godoc
// @Description Restaurant Zone
// @Tags		Restaurant
// @Param		Zone  body      models.Zone  true  "Zone"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /restaurant/zone [post]
func (h ZoneHttp) CreateZone(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.Zone{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateZone(shopID, authUsername, *docReq)

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

// Update Restaurant Zone godoc
// @Description Restaurant Zone
// @Tags		Restaurant
// @Param		id  path      string  true  "Zone ID"
// @Param		Zone  body      models.Zone  true  "Zone"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /restaurant/zone/{id} [put]
func (h ZoneHttp) UpdateZone(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.Zone{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateZone(shopID, id, authUsername, *docReq)

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

// Delete Restaurant Zone godoc
// @Description Restaurant Zone
// @Tags		Restaurant
// @Param		id  path      string  true  "Zone ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /restaurant/zone/{id} [delete]
func (h ZoneHttp) DeleteZone(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteZone(shopID, id, authUsername)

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

// Delete Restaurant Zone godoc
// @Description Delete Restaurant Zone
// @Tags		Restaurant
// @Param		Zone  body      []string  true  "Zone GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /restaurant/zone [delete]
func (h ZoneHttp) DeleteByGUIDs(ctx microservice.IContext) error {
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

// Get Restaurant Zone Infomation godoc
// @Description Get Restaurant Zone
// @Tags		Restaurant
// @Param		id  path      string  true  "Zone Id"
// @Accept 		json
// @Success		200	{object}	models.ZoneInfoResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /restaurant/zone/{id} [get]
func (h ZoneHttp) InfoZone(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get Zone %v", id)
	doc, err := h.svc.InfoZone(shopID, id)

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

// Get Restaurant Zone By code array godoc
// @Description get Restaurant Zone by code array
// @Tags		Restaurant
// @Param		codes	query	string		false  "Code filter, json array encode "
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /restaurant/zone/by-code [get]
func (h ZoneHttp) InfoArray(ctx microservice.IContext) error {
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
	doc, err := h.svc.InfoWTFArray(shopID, docReq)

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

// List Restaurant Zone godoc
// @Description List Restaurant Zone Category
// @Tags		Restaurant
// @Param		group-number	query	integer		false  "Group Number"
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Size"
// @Accept 		json
// @Success		200	{object}	models.ZonePageResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /restaurant/zone [get]
func (h ZoneHttp) SearchZone(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)
	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "group-number",
			Field: "groupnumber",
			Type:  requestfilter.FieldTypeInt,
		},
	})

	docList, pagination, err := h.svc.SearchZone(shopID, filters, pageable)

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

// Create Zone Bulk godoc
// @Description Create Zone
// @Tags		Restaurant
// @Param		Zone  body      []models.Zone  true  "Zone"
// @Accept 		json
// @Success		201	{object}	common.BulkInsertResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /restaurant/zone/bulk [post]
func (h ZoneHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.Zone{}
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
