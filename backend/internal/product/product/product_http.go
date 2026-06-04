package products

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"smlcloudplatform/internal/config"
	creditorepo "smlcloudplatform/internal/debtaccount/creditor/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/product/product/models"
	"smlcloudplatform/internal/product/product/repositories"
	"smlcloudplatform/internal/product/product/services"
	productBarcodeRepo "smlcloudplatform/internal/product/productbarcode/repositories"
	unitRepo "smlcloudplatform/internal/product/unit/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
	"strings"
	"time"
)

type IProductHttp interface{}

type ProductHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IProductHttpService
}

// ✅ **สร้าง New ProductHttp**
func NewProductHttp(ms *microservice.Microservice, cfg config.IConfig) ProductHttp {

	pstmg := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	repo := repositories.NewProductRepository(pstmg)
	repoUnit := unitRepo.NewUnitRepository(pstmg)
	repomgCreditor := creditorepo.NewCreditorRepository(pstmg)
	repomgProductBarcode := productBarcodeRepo.NewProductBarcodeRepository(pstmg, cache)
	svc := services.NewProductHttpService(repo, repoUnit, *repomgCreditor, *repomgProductBarcode)

	return ProductHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

// ✅ **Register API Routes**
func (h ProductHttp) RegisterHttp() {
	h.ms.GET("/product", h.SearchProduct)
	h.ms.POST("/product", h.CreateProduct)
	h.ms.GET("/product/:guid", h.InfoProduct)
	h.ms.PUT("/product/:guid", h.UpdateProduct)
	h.ms.DELETE("/product/:guid", h.DeleteProduct)
}

// @Summary		Search products
// @Description Search products with pagination
// @Tags		Product
// @Accept 		json
// @Produce 	json
// @Param		q query string false "Keyword to search"
// @Param		page query int false "Page number"
// @Param		limit query int false "Items per page"
// @Success		200 {object} common.ApiResponse{data=[]models.ProductInfo}
// @Failure		400 {object} common.ApiResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security	AccessToken
// @Router		/product [get]
func (h ProductHttp) SearchProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	pageable := utils.GetPageable(ctx.QueryParam)

	filters := h.searchFilter(ctx.QueryParam)

	docList, pagination, err := h.svc.ProductList(holdingCode, filters, pageable)

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

// @Summary		Create a new product
// @Description Create a new product with details
// @Tags		Product
// @Accept 		json
// @Produce 	json
// @Param		Product body models.ProductDoc true "Product data"
// @Success		201 {object} common.ApiResponse{data=models.ProductDoc}
// @Failure		400 {object} common.ApiResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security	AccessToken
// @Router		/product [post]
func (h ProductHttp) CreateProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := strings.TrimSpace(userInfo.HoldingCode)

	input := strings.TrimSpace(ctx.ReadInput())
	if input == "" {
		ctx.ResponseError(http.StatusBadRequest, "Invalid input: Empty request body")
		return errors.New("Invalid input: Empty request body")
	}

	// ✅ แปลง JSON เป็น struct
	newProduct := &models.ProductDoc{}
	err := json.Unmarshal([]byte(input), newProduct)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, "Invalid JSON format: "+err.Error())
		return err
	}

	// ✅ กำหนดค่า `HoldingCode` และ `GuidFixed`
	newProduct.HoldingCode = holdingCode
	newProduct.GuidFixed = utils.NewGUID()

	// ✅ ตรวจสอบ Validation
	if err = ctx.Validate(newProduct); err != nil {
		ctx.ResponseError(http.StatusBadRequest, "Validation failed: "+err.Error())
		return err
	}

	// ✅ กำหนดค่า `CreatedBy` และ `CreatedAt`
	newProduct.CreatedBy = userInfo.Username
	newProduct.CreatedAt = time.Now()

	// ✅ Debug
	fmt.Println("Creating Product:", newProduct)

	// ✅ เรียก Service เพื่อสร้าง Product
	err = h.svc.Create(newProduct)
	if err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		Message: "Product created successfully",
		Data:    newProduct,
	})
	return nil
}

// @Summary		Get product details
// @Description Get product details by code
// @Tags		Product
// @Accept 		json
// @Produce 	json
// @Param		guid path string true "Product guid"
// @Success		200 {object} common.ApiResponse{data=models.ProductDoc}
// @Failure		400 {object} common.ApiResponse
// @Failure		404 {object} common.ApiResponse
// @Security	AccessToken
// @Router		/product/{guid} [get]
func (h ProductHttp) InfoProduct(ctx microservice.IContext) error {
	code := strings.TrimSpace(ctx.Param("guid"))
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	if code == "" {
		ctx.ResponseError(http.StatusBadRequest, "Product Code is required")
		return errors.New("Product Code is required")
	}

	product, err := h.svc.GetProduct(holdingCode, code)
	if err != nil {
		ctx.ResponseError(http.StatusNotFound, "Product not found")
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    product,
	})
	return nil
}

// @Summary		Update an existing product
// @Description Update an existing product by guid
// @Tags		Product
// @Accept 		json
// @Produce 	json
// @Param		guid path string true "Product Guid"
// @Param		Product body models.ProductDoc true "Updated product data"
// @Success		200 {object} common.ApiResponse{data=models.ProductDoc}
// @Failure		400 {object} common.ApiResponse
// @Failure		404 {object} common.ApiResponse
// @Security	AccessToken
// @Router		/product/{guid} [put]
func (h ProductHttp) UpdateProduct(ctx microservice.IContext) error {
	code := strings.TrimSpace(ctx.Param("guid"))
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	if code == "" {
		ctx.ResponseError(http.StatusBadRequest, "Product Code is required")
		return errors.New("Product Code is required")
	}

	input := strings.TrimSpace(ctx.ReadInput())
	if input == "" {
		ctx.ResponseError(http.StatusBadRequest, "Invalid input: Empty request body")
		return errors.New("Invalid input: Empty request body")
	}

	// ✅ แปลง JSON เป็น struct
	updateData := &models.ProductDoc{}
	err := json.Unmarshal([]byte(input), updateData)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, "Invalid JSON format: "+err.Error())
		return err
	}

	updateData.UpdatedBy = userInfo.Username
	updateData.UpdatedAt = time.Now()

	// ✅ Debug
	fmt.Println("Updating Product:", updateData)

	// ✅ อัปเดต Product
	docData, err := h.svc.Update(holdingCode, code, userInfo.Username, updateData)
	if err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Message: "Product updated successfully",
		Data:    docData,
	})
	return nil
}

// @Summary		Delete a product
// @Description Delete a product by guid
// @Tags		Product
// @Accept 		json
// @Produce 	json
// @Param		guid path string true "Product Code"
// @Success		200 {object} common.ApiResponse
// @Failure		400 {object} common.ApiResponse
// @Failure		404 {object} common.ApiResponse
// @Security	AccessToken
// @Router		/product/{guid} [delete]
func (h ProductHttp) DeleteProduct(ctx microservice.IContext) error {
	code := strings.TrimSpace(ctx.Param("guid"))
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	if code == "" {
		ctx.ResponseError(http.StatusBadRequest, "Product Code is required")
		return errors.New("Product Code is required")
	}

	err := h.svc.Delete(holdingCode, code, userInfo.Username)
	if err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Message: "Product deleted successfully",
	})
	return nil
}

func (h ProductHttp) searchFilter(queryParam func(string) string) map[string]interface{} {
	filters := requestfilter.GenerateFilters(queryParam, []requestfilter.FilterRequest{
		{
			Param: "code",
			Field: "code",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "names",
			Field: "names.name",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "item_type",
			Field: "item_type",
			Type:  requestfilter.FieldTypeInt,
		},
		{
			Param: "materialtype",
			Field: "materialtype",
			Type:  requestfilter.FieldTypeInt,
		},
	})

	return filters
}
