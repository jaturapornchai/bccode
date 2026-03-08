package shopcoupon

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/shopcoupon/models"
	"smlcloudplatform/internal/shopcoupon/repositories"
	"smlcloudplatform/internal/shopcoupon/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
	"strconv"
	"strings"
)

type IShopCouponHttp interface{}

type ShopCouponHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IShopCouponHttpService
}

func NewShopCouponHttp(ms *microservice.Microservice, cfg config.IConfig) ShopCouponHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())

	repo := repositories.NewShopCouponRepository(pst)

	svc := services.NewShopCouponHttpService(repo)

	return ShopCouponHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h ShopCouponHttp) RegisterHttp() {

	h.ms.GET("/shopcoupon", h.SearchShopCoupon)
	h.ms.POST("/shopcoupon", h.CreateShopCoupon)
	h.ms.GET("/shopcoupon/:id", h.InfoShopCoupon)
	h.ms.PUT("/shopcoupon/:id", h.UpdateShopCoupon)
	h.ms.DELETE("/shopcoupon/:id", h.DeleteShopCoupon)
}

func (h ShopCouponHttp) CreateShopCoupon(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.ShopCoupon{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateShopCoupon(shopID, authUsername, *docReq)

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

func (h ShopCouponHttp) UpdateShopCoupon(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.ShopCoupon{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateShopCoupon(shopID, id, authUsername, *docReq)

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

func (h ShopCouponHttp) DeleteShopCoupon(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteShopCoupon(id, shopID, authUsername)

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

func (h ShopCouponHttp) InfoShopCoupon(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get ShopCoupon %v", id)
	doc, err := h.svc.InfoShopCoupon(shopID, id)

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

func (h ShopCouponHttp) SearchShopCoupon(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	filterMatch := map[string]interface{}{}

	paramType := strings.TrimSpace(ctx.QueryParam("type"))

	if len(paramType) > 0 {
		couponType, err := strconv.Atoi(paramType)
		if err == nil {
			filterMatch["type"] = couponType
		}
	}

	docList, pagination, err := h.svc.SearchShopCoupon(shopID, filterMatch, pageable)

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
