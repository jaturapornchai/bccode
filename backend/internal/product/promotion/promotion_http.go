package promotion

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/product/promotion/models"
	"smlcloudplatform/internal/product/promotion/repositories"
	"smlcloudplatform/internal/product/promotion/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

type IPromotionHttp interface{}

type PromotionHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IPromotionHttpService
}

func NewPromotionHttp(ms *microservice.Microservice, cfg config.IConfig) PromotionHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewPromotionRepository(pst)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewPromotionHttpService(repo, masterSyncCacheRepo)

	return PromotionHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h PromotionHttp) RegisterHttp() {

	h.ms.POST("/product/promotion/bulk", h.SaveBulk)

	h.ms.GET("/product/promotion", h.SearchPromotionPage)
	h.ms.GET("/product/promotion/list", h.SearchPromotionStep)
	h.ms.POST("/product/promotion", h.CreatePromotion)
	h.ms.GET("/product/promotion/:id", h.InfoPromotion)
	h.ms.GET("/product/promotion/code/:code", h.InfoPromotionByCode)
	h.ms.PUT("/product/promotion/:id", h.UpdatePromotion)
	h.ms.DELETE("/product/promotion/:id", h.DeletePromotion)
	h.ms.DELETE("/product/promotion", h.DeletePromotionByGUIDs)
}

// Create Promotion godoc
// @Description Create Promotion
// @Tags		Promotion
// @Param		Promotion  body      models.Promotion  true  "Promotion"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/promotion [post]
func (h PromotionHttp) CreatePromotion(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.Promotion{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreatePromotion(shopID, authUsername, *docReq)

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

// Update Promotion godoc
// @Description Update Promotion
// @Tags		Promotion
// @Param		id  path      string  true  "Promotion ID"
// @Param		Promotion  body      models.Promotion  true  "Promotion"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/promotion/{id} [put]
func (h PromotionHttp) UpdatePromotion(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.Promotion{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdatePromotion(shopID, id, authUsername, *docReq)

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

// Delete Promotion godoc
// @Description Delete Promotion
// @Tags		Promotion
// @Param		id  path      string  true  "Promotion ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/promotion/{id} [delete]
func (h PromotionHttp) DeletePromotion(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeletePromotion(shopID, id, authUsername)

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

// Delete Promotion godoc
// @Description Delete Promotion
// @Tags		Promotion
// @Param		Promotion  body      []string  true  "Promotion GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/promotion [delete]
func (h PromotionHttp) DeletePromotionByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeletePromotionByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get Promotion godoc
// @Description get Promotion info by guidfixed
// @Tags		Promotion
// @Param		id  path      string  true  "Promotion guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/promotion/{id} [get]
func (h PromotionHttp) InfoPromotion(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get Promotion %v", id)
	doc, err := h.svc.InfoPromotion(shopID, id)

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

// Get Promotion By Code godoc
// @Description get Promotion info by Code
// @Tags		Promotion
// @Param		code  path      string  true  "Promotion Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/promotion/code/{code} [get]
func (h PromotionHttp) InfoPromotionByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")

	doc, err := h.svc.InfoPromotionByCode(shopID, code)

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

// List Promotion step godoc
// @Description get list step
// @Tags		Promotion
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/promotion [get]
func (h PromotionHttp) SearchPromotionPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	// filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
	// 	{
	// 		Param: "promotiontype",
	// 		Field: "promotiontype",
	// 		Type:  requestfilter.FieldTypeInt,
	// 	},
	// })

	docList, pagination, err := h.svc.SearchPromotion(shopID, pageable)

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

// List Promotion godoc
// @Description search limit offset
// @Tags		Promotion
// @Param		q		query	string		false  "Search Value"
// @Param		promotiontype	query	int		false  "promotiontype"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/promotion/list [get]
func (h PromotionHttp) SearchPromotionStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "promotiontype",
			Field: "promotiontype",
			Type:  requestfilter.FieldTypeInt,
		},
	})

	docList, total, err := h.svc.SearchPromotionStep(shopID, lang, filters, pageableStep)

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

// Create Promotion Bulk godoc
// @Description Create Promotion
// @Tags		Promotion
// @Param		Promotion  body      []models.Promotion  true  "Promotion"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/promotion/bulk [post]
func (h PromotionHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.Promotion{}
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
