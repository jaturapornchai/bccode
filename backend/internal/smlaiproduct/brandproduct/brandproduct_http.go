package brandproduct

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/internal/smlaiproduct/brandproduct/models"
	"smlcloudplatform/internal/smlaiproduct/brandproduct/repositories"
	"smlcloudplatform/internal/smlaiproduct/brandproduct/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IBrandProductHttp interface{}

type BrandProductHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IBrandProductHttpService
}

func NewBrandProductHttp(ms *microservice.Microservice, cfg config.IConfig) BrandProductHttp {

	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewBrandProductRepository(pst)
	repoProductBarcode := productbarcode_repositories.NewProductBarcodeRepository(pst, cache)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewBrandProductHttpService(repo, repoProductBarcode, masterSyncCacheRepo, 15*time.Second)

	return BrandProductHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h BrandProductHttp) RegisterHttp() {

	h.ms.POST("/aicloud/brand/bulk", h.SaveBulk)

	h.ms.GET("/aicloud/brand", h.SearchBrandProductPage)
	h.ms.GET("/aicloud/brand/list", h.SearchBrandProductStep)
	h.ms.POST("/aicloud/brand", h.CreateBrandProduct)
	h.ms.GET("/aicloud/brand/:id", h.InfoBrandProduct)
	h.ms.GET("/aicloud/brand/code/:code", h.InfoBrandProductByCode)
	h.ms.PUT("/aicloud/brand/:id", h.UpdateBrandProduct)
	h.ms.DELETE("/aicloud/brand/:id", h.DeleteBrandProduct)
	h.ms.DELETE("/aicloud/brand", h.DeleteBrandProductByGUIDs)
}

// Create BrandProduct godoc
// @Description Create BrandProduct
// @Tags		AICloudProductBrand
// @Param		BrandProduct  body      models.BrandProduct  true  "BrandProduct"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/brand [post]
func (h BrandProductHttp) CreateBrandProduct(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	docReq := &models.BrandProduct{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateBrandProduct(holdingCode, authUsername, *docReq)

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

// Update BrandProduct godoc
// @Description Update BrandProduct
// @Tags		AICloudProductBrand
// @Param		id  path      string  true  "BrandProduct ID"
// @Param		BrandProduct  body      models.BrandProduct  true  "BrandProduct"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/brand/{id} [put]
func (h BrandProductHttp) UpdateBrandProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.BrandProduct{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateBrandProduct(holdingCode, id, authUsername, *docReq)

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

// Delete BrandProduct godoc
// @Description Delete BrandProduct
// @Tags		AICloudProductBrand
// @Param		id  path      string  true  "BrandProduct ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/brand/{id} [delete]
func (h BrandProductHttp) DeleteBrandProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteBrandProduct(holdingCode, id, authUsername)

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

// Delete BrandProduct godoc
// @Description Delete BrandProduct
// @Tags		AICloudProductBrand
// @Param		BrandProduct  body      []string  true  "BrandProduct GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/brand [delete]
func (h BrandProductHttp) DeleteBrandProductByGUIDs(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	input := ctx.ReadInput()

	docReq := []string{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.DeleteBrandProductByGUIDs(holdingCode, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get BrandProduct godoc
// @Description get BrandProduct info by guidfixed
// @Tags		AICloudProductBrand
// @Param		id  path      string  true  "BrandProduct guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/brand/{id} [get]
func (h BrandProductHttp) InfoBrandProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get BrandProduct %v", id)
	doc, err := h.svc.InfoBrandProduct(holdingCode, id)

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

// Get BrandProduct By Code godoc
// @Description get BrandProduct info by Code
// @Tags		AICloudProductBrand
// @Param		code  path      string  true  "BrandProduct Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/brand/code/{code} [get]
func (h BrandProductHttp) InfoBrandProductByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	code := ctx.Param("code")

	doc, err := h.svc.InfoBrandProductByCode(holdingCode, code)

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

// List BrandProduct step godoc
// @Description get list step
// @Tags		AICloudProductBrand
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/brand [get]
func (h BrandProductHttp) SearchBrandProductPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.svc.SearchBrandProduct(holdingCode, map[string]interface{}{}, pageable)

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

// List BrandProduct godoc
// @Description search limit offset
// @Tags		AICloudProductBrand
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/brand/list [get]
func (h BrandProductHttp) SearchBrandProductStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	docList, total, err := h.svc.SearchBrandProductStep(holdingCode, lang, pageableStep)

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

// Create BrandProduct Bulk godoc
// @Description Create BrandProduct
// @Tags		AICloudProductBrand
// @Param		BrandProduct  body      []models.BrandProduct  true  "BrandProduct"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/brand/bulk [post]
func (h BrandProductHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()

	dataReq := []models.BrandProduct{}
	err := json.Unmarshal([]byte(input), &dataReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	bulkResponse, err := h.svc.SaveInBatch(holdingCode, authUsername, dataReq)

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
