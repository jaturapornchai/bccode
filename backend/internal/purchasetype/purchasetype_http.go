package purchasetype

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/purchasetype/models"
	"smlcloudplatform/internal/purchasetype/repositories"
	"smlcloudplatform/internal/purchasetype/services"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IPurchaseTypeHttp interface{}

type PurchaseTypeHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IPurchaseTypeHttpService
}

func NewPurchaseTypeHttp(ms *microservice.Microservice, cfg config.IConfig) PurchaseTypeHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewPurchaseTypeRepository(pst)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewPurchaseTypeHttpService(repo, masterSyncCacheRepo, 15*time.Second)

	return PurchaseTypeHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h PurchaseTypeHttp) RegisterHttp() {

	h.ms.GET("/purchase-type", h.SearchPurchaseTypePage)
	h.ms.GET("/purchase-type/list", h.SearchPurchaseTypeStep)
	h.ms.POST("/purchase-type", h.CreatePurchaseType)
	h.ms.GET("/purchase-type/:id", h.InfoPurchaseType)
	h.ms.GET("/purchase-type/code/:code", h.InfoPurchaseTypeByCode)
	h.ms.PUT("/purchase-type/:id", h.UpdatePurchaseType)
	h.ms.DELETE("/purchase-type/:id", h.DeletePurchaseType)
	h.ms.DELETE("/purchase-type", h.DeletePurchaseTypeByGUIDs)
}

// Create PurchaseType godoc
// @Description Create PurchaseType
// @Tags		PurchaseType
// @Param		PurchaseType  body      models.PurchaseType  true  "PurchaseType"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /purchase-type [post]
func (h PurchaseTypeHttp) CreatePurchaseType(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.PurchaseType{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreatePurchaseType(shopID, authUsername, *docReq)

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

// Update PurchaseType godoc
// @Description Update PurchaseType
// @Tags		PurchaseType
// @Param		id  path      string  true  "PurchaseType ID"
// @Param		PurchaseType  body      models.PurchaseType  true  "PurchaseType"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /purchase-type/{id} [put]
func (h PurchaseTypeHttp) UpdatePurchaseType(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.PurchaseType{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdatePurchaseType(shopID, id, authUsername, *docReq)

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

// Delete PurchaseType godoc
// @Description Delete PurchaseType
// @Tags		PurchaseType
// @Param		id  path      string  true  "PurchaseType ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /purchase-type/{id} [delete]
func (h PurchaseTypeHttp) DeletePurchaseType(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeletePurchaseType(shopID, id, authUsername)

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

// Delete PurchaseType godoc
// @Description Delete PurchaseType
// @Tags		PurchaseType
// @Param		PurchaseType  body      []string  true  "PurchaseType GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /purchase-type [delete]
func (h PurchaseTypeHttp) DeletePurchaseTypeByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeletePurchaseTypeByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get PurchaseType godoc
// @Description get PurchaseType info by guidfixed
// @Tags		PurchaseType
// @Param		id  path      string  true  "PurchaseType guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /purchase-type/{id} [get]
func (h PurchaseTypeHttp) InfoPurchaseType(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get PurchaseType %v", id)
	doc, err := h.svc.InfoPurchaseType(shopID, id)

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

// Get PurchaseType By Code godoc
// @Description get PurchaseType info by Code
// @Tags		PurchaseType
// @Param		code  path      string  true  "PurchaseType Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /purchase-type/code/{code} [get]
func (h PurchaseTypeHttp) InfoPurchaseTypeByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")

	doc, err := h.svc.InfoPurchaseTypeByCode(shopID, code)

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

// List PurchaseType step godoc
// @Description get list step
// @Tags		PurchaseType
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /purchase-type [get]
func (h PurchaseTypeHttp) SearchPurchaseTypePage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{})

	docList, pagination, err := h.svc.SearchPurchaseType(shopID, filters, pageable)

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

// List PurchaseType godoc
// @Description search limit offset
// @Tags		PurchaseType
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /purchase-type/list [get]
func (h PurchaseTypeHttp) SearchPurchaseTypeStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{})

	docList, total, err := h.svc.SearchPurchaseTypeStep(shopID, lang, filters, pageableStep)

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
