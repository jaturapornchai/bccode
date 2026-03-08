// categoryproduct_http.go
package categoryproduct

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/internal/smlaiproduct/categoryproduct/models"
	categoryrepo "smlcloudplatform/internal/smlaiproduct/categoryproduct/repositories"
	categorysvc "smlcloudplatform/internal/smlaiproduct/categoryproduct/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type ICategoryProductHttp interface{}

type CategoryProductHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc categorysvc.ICategoryProductHttpService
}

func NewCategoryProductHttp(ms *microservice.Microservice, cfg config.IConfig) CategoryProductHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := categoryrepo.NewCategoryProductRepository(pst)
	repoProductBarcode := productbarcode_repositories.NewProductBarcodeRepository(pst, cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)

	svc := categorysvc.NewCategoryProductHttpService(repo, repoProductBarcode, masterSyncCacheRepo, 15*time.Second)

	return CategoryProductHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h CategoryProductHttp) RegisterHttp() {
	h.ms.POST("/aicloud/category/bulk", h.SaveBulk)
	h.ms.GET("/aicloud/category", h.SearchCategoryProductPage)
	h.ms.GET("/aicloud/category/list", h.SearchCategoryProductStep)
	h.ms.POST("/aicloud/category", h.CreateCategoryProduct)
	h.ms.GET("/aicloud/category/:id", h.InfoCategoryProduct)
	h.ms.GET("/aicloud/category/code/:code", h.InfoCategoryProductByCode)
	h.ms.PUT("/aicloud/category/:id", h.UpdateCategoryProduct)
	h.ms.DELETE("/aicloud/category/:id", h.DeleteCategoryProduct)
	h.ms.DELETE("/aicloud/category", h.DeleteCategoryProductByGUIDs)
}

// Create CategoryProduct godoc
// @Description Create CategoryProduct
// @Tags		AICloudProductCategory
// @Param		CategoryProduct	body	models.CategoryProduct	true	"CategoryProduct"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401	{object}	common.AuthResponseFailed
// @Security		AccessToken
// @Router /aicloud/category [post]
func (h CategoryProductHttp) CreateCategoryProduct(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()
	docReq := &models.CategoryProduct{}
	if err := json.Unmarshal([]byte(input), docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err := ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateCategoryProduct(shopID, authUsername, *docReq)
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

// Update CategoryProduct godoc
// @Description Update CategoryProduct
// @Tags		AICloudProductCategory
// @Param		id	path	string	true	"CategoryProduct ID"
// @Param		CategoryProduct	body	models.CategoryProduct	true	"CategoryProduct"
// @Accept		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401	{object}	common.AuthResponseFailed
// @Security		AccessToken
// @Router /aicloud/category/{id} [put]
func (h CategoryProductHttp) UpdateCategoryProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()
	docReq := &models.CategoryProduct{}
	if err := json.Unmarshal([]byte(input), docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err := ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err := h.svc.UpdateCategoryProduct(shopID, id, authUsername, *docReq)
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

// Delete CategoryProduct godoc
// @Description Delete CategoryProduct
// @Tags		AICloudProductCategory
// @Param		id  path      string  true  "CategoryProduct ID"
// @Accept 		json
// @Success		200	{object} common.ResponseSuccessWithID
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/category/{id} [delete]
func (h CategoryProductHttp) DeleteCategoryProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteCategoryProduct(shopID, id, authUsername)
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

// Delete CategoryProduct godoc
// @Description Delete CategoryProduct
// @Tags		AICloudProductCategory
// @Param		CategoryProduct  body      []string  true  "CategoryProduct GUIDs"
// @Accept 		json
// @Success		200	{object} common.ResponseSuccessWithID
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/category [delete]
func (h CategoryProductHttp) DeleteCategoryProductByGUIDs(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	input := ctx.ReadInput()
	docReq := []string{}
	if err := json.Unmarshal([]byte(input), &docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err := h.svc.DeleteCategoryProductByGUIDs(shopID, authUsername, docReq)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})
	return nil
}

// Get CategoryProduct godoc
// @Description get CategoryProduct info by guidfixed
// @Tags		AICloudProductCategory
// @Param		id  path      string  true  "CategoryProduct guidfixed"
// @Accept 		json
// @Success		200	{object} common.ApiResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/category/{id} [get]
func (h CategoryProductHttp) InfoCategoryProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get CategoryProduct %v", id)
	doc, err := h.svc.InfoCategoryProduct(shopID, id)
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

// Get CategoryProduct By Code godoc
// @Description get CategoryProduct info by Code
// @Tags		AICloudProductCategory
// @Param		code  path      string  true  "CategoryProduct Code"
// @Accept 		json
// @Success		200	{object} common.ApiResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/category/code/{code} [get]
func (h CategoryProductHttp) InfoCategoryProductByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")
	doc, err := h.svc.InfoCategoryProductByCode(shopID, code)
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

// List CategoryProduct step godoc
// @Description get list step
// @Tags		AICloudProductCategory
// @Param		q	query	string	false  "Search Value"
// @Param		page	query	integer	false  "Page"
// @Param		limit	query	integer	false  "Limit"
// @Accept 		json
// @Success		200	{array} common.ApiResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/category [get]
func (h CategoryProductHttp) SearchCategoryProductPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)
	docList, pagination, err := h.svc.SearchCategoryProduct(shopID, map[string]interface{}{}, pageable)
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

// List CategoryProduct godoc
// @Description search limit offset
// @Tags		AICloudProductCategory
// @Param		q	query	string	false  "Search Value"
// @Param		offset	query	integer	false  "offset"
// @Param		limit	query	integer	false  "limit"
// @Param		lang	query	string	false  "lang"
// @Accept 		json
// @Success		200	{array} common.ApiResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/category/list [get]
func (h CategoryProductHttp) SearchCategoryProductStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)
	lang := ctx.QueryParam("lang")
	docList, total, err := h.svc.SearchCategoryProductStep(shopID, lang, pageableStep)
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

// Create CategoryProduct Bulk godoc
// @Description Create CategoryProduct
// @Tags		AICloudProductCategory
// @Param		CategoryProduct  body      []models.CategoryProduct  true  "CategoryProduct"
// @Accept 		json
// @Success		201	{object} common.BulkResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/category/bulk [post]
func (h CategoryProductHttp) SaveBulk(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()
	dataReq := []models.CategoryProduct{}
	if err := json.Unmarshal([]byte(input), &dataReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	bulkResponse, err := h.svc.SaveInBatch(shopID, authUsername, dataReq)
	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.BulkResponse{
		Success:    true,
		BulkImport: bulkResponse,
	})
	return nil
}
