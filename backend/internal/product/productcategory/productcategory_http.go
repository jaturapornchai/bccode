package productcategory

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcodeRepo "smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/internal/product/productcategory/models"
	"smlcloudplatform/internal/product/productcategory/repositories"
	"smlcloudplatform/internal/product/productcategory/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

type IProductCategoryHttp interface{}

type ProductCategoryHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IProductCategoryHttpService
}

func NewProductCategoryHttp(ms *microservice.Microservice, cfg config.IConfig) ProductCategoryHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewProductCategoryRepository(pst)
	productBarcodeRepository := productbarcodeRepo.NewProductBarcodeRepository(pst, cache)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewProductCategoryHttpService(repo, masterSyncCacheRepo, productBarcodeRepository)

	return ProductCategoryHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h ProductCategoryHttp) RegisterHttp() {

	h.ms.POST("/product/category/bulk", h.SaveBulk)

	h.ms.GET("/product/category", h.SearchProductCategoryPage)
	h.ms.GET("/product/category/list", h.SearchProductCategoryLimit)
	h.ms.POST("/product/category", h.CreateProductCategory)
	h.ms.GET("/product/category/:id", h.InfoProductCategory)
	h.ms.PUT("/product/category/xsort", h.UpdateProductCategoryXSort)
	h.ms.PUT("/product/category/barcodes", h.UpdateProductCategoryBarcodes)
	h.ms.PUT("/product/category/:id", h.UpdateProductCategory)
	h.ms.DELETE("/product/category/:id", h.DeleteProductCategory)
	h.ms.DELETE("/product/category", h.DeleteProductCategoryByGUIDs)
}

// Create ProductCategory godoc
// @Description Create ProductCategory
// @Tags		ProductCategory
// @Param		ProductCategory  body      models.ProductCategory  true  "ProductCategory"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/category [post]
func (h ProductCategoryHttp) CreateProductCategory(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	docReq := &models.ProductCategory{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if docReq.XSorts == nil {
		docReq.XSorts = &[]common.XSort{}
	}

	if docReq.CodeList == nil {
		docReq.CodeList = &[]models.CodeXSort{}
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateProductCategory(holdingCode, authUsername, *docReq)

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

// Update ProductCategory godoc
// @Description Update ProductCategory
// @Tags		ProductCategory
// @Param		id  path      string  true  "ProductCategory ID"
// @Param		ProductCategory  body      models.ProductCategory  true  "ProductCategory"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/category/{id} [put]
func (h ProductCategoryHttp) UpdateProductCategory(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.ProductCategory{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if docReq.XSorts == nil {
		docReq.XSorts = &[]common.XSort{}
	}

	if docReq.CodeList == nil {
		docReq.CodeList = &[]models.CodeXSort{}
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateProductCategory(holdingCode, id, authUsername, *docReq)

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

// Update XSort	 Category godoc
// @Description Update XSort Category
// @Tags		ProductCategory
// @Param		XSort  body      []common.XSortModifyReqesut  true  "XSort"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/category/xsort [put]
func (h ProductCategoryHttp) UpdateProductCategoryXSort(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()

	req := &[]common.XSortModifyReqesut{}
	err := json.Unmarshal([]byte(input), &req)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(req); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.XSortsSave(holdingCode, authUsername, *req)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Update Barcodes	 Category godoc
// @Description Update Barcodes Category
// @Tags		ProductCategory
// @Param		Barcodes  body      []models.BarcodesModifyReqesut  true  "Barcodes"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/category/barcodes [put]
func (h ProductCategoryHttp) UpdateProductCategoryBarcodes(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()

	req := &[]common.XSortModifyReqesut{}
	err := json.Unmarshal([]byte(input), &req)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(req); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.XBarcodesSave(holdingCode, authUsername, *req)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Delete ProductCategory godoc
// @Description Delete ProductCategory
// @Tags		ProductCategory
// @Param		id  path      string  true  "ProductCategory ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/category/{id} [delete]
func (h ProductCategoryHttp) DeleteProductCategory(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteProductCategory(holdingCode, id, authUsername)

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

// Get ProductCategory godoc
// @Description get struct array by ID
// @Tags		ProductCategory
// @Param		id  path      string  true  "ProductCategory ID"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/category/{id} [get]
func (h ProductCategoryHttp) InfoProductCategory(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get ProductCategory %v", id)
	doc, err := h.svc.InfoProductCategory(holdingCode, id)

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

// List ProductCategory godoc
// @Description get struct array by ID
// @Tags		ProductCategory
// @Param		q		query	string		false  "Search Value"
// @Param		group-number		query	int		false  "group number"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/category [get]
func (h ProductCategoryHttp) SearchProductCategoryPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageable := utils.GetPageable(ctx.QueryParam)
	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "group-number",
			Field: "group_number",
			Type:  requestfilter.FieldTypeInt,
		},
	})

	docList, pagination, err := h.svc.SearchProductCategory(holdingCode, filters, pageable)

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

// List ProductCategory godoc
// @Description search limit offset
// @Tags		ProductCategory
// @Param		q		query	string		false  "Search Value"
// @Param		group-number		query	int		false  "group number"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/category/list [get]
func (h ProductCategoryHttp) SearchProductCategoryLimit(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "group-number",
			Field: "group_number",
			Type:  requestfilter.FieldTypeInt,
		},
	})

	lang := ctx.QueryParam("lang")

	docList, total, err := h.svc.SearchProductCategoryStep(holdingCode, lang, filters, pageableStep)

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

// Create ProductCategory Bulk godoc
// @Description Create ProductCategory
// @Tags		ProductCategory
// @Param		ProductCategory  body      []models.ProductCategory  true  "ProductCategory"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/category/bulk [post]
func (h ProductCategoryHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()

	dataReq := []models.ProductCategory{}
	err := json.Unmarshal([]byte(input), &dataReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.SaveInBatch(holdingCode, authUsername, dataReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	ctx.Response(
		http.StatusCreated,
		common.ApiResponse{
			Success: true,
		},
	)

	return nil
}

// Delete ProductCategory By GUIDs godoc
// @Description Delete ProductCategory
// @Tags		ProductCategory
// @Param		ProductCategory  body      []string  true  "ProductCategory GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/category [delete]
func (h ProductCategoryHttp) DeleteProductCategoryByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteProductCategoryByGUIDs(holdingCode, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}
