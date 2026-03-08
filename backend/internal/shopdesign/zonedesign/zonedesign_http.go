package zonedesign

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/shopdesign/zonedesign/models"
	"smlcloudplatform/internal/shopdesign/zonedesign/repositories"
	"smlcloudplatform/internal/shopdesign/zonedesign/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"

	common "smlcloudplatform/internal/models"
)

type ZoneDesignHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IZoneDesignService
}

func NewZoneDesignHttp(ms *microservice.Microservice, cfg config.IConfig) ZoneDesignHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())

	repo := repositories.NewZoneDesignRepository(pst)
	svc := services.NewZoneDesignService(repo)

	return ZoneDesignHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h ZoneDesignHttp) RegisterHttp() {

	h.ms.GET("/zone-design", h.SearchZoneDesign)
	h.ms.POST("/zone-design", h.CreateZoneDesign)
	h.ms.GET("/zone-design/:id", h.InfoZoneDesign)
	h.ms.PUT("/zone-design/:id", h.UpdateZoneDesign)
	h.ms.DELETE("/zone-design/:id", h.DeleteZoneDesign)
}

func (h ZoneDesignHttp) CreateZoneDesign(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.ZoneDesign{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateZoneDesign(shopID, authUsername, *docReq)

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

func (h ZoneDesignHttp) UpdateZoneDesign(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.ZoneDesign{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateZoneDesign(shopID, id, authUsername, *docReq)

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

func (h ZoneDesignHttp) DeleteZoneDesign(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteZoneDesign(shopID, id, authUsername)

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

func (h ZoneDesignHttp) InfoZoneDesign(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get ZoneDesign %v", id)
	doc, err := h.svc.InfoZoneDesign(shopID, id)

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

func (h ZoneDesignHttp) SearchZoneDesign(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)
	docList, pagination, err := h.svc.SearchZoneDesign(shopID, pageable)

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
