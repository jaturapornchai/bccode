package products

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"smlcloudplatform/internal/config"
	creditorepo "smlcloudplatform/internal/debtaccount/creditor/repositories"
	build "smlcloudplatform/internal/goapi/process/build"
	"smlcloudplatform/internal/logger"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/product/product/models"
	"smlcloudplatform/internal/product/product/outbox"
	"smlcloudplatform/internal/product/product/repositories"
	"smlcloudplatform/internal/product/product/services"
	productBarcodeRepo "smlcloudplatform/internal/product/productbarcode/repositories"
	unitRepo "smlcloudplatform/internal/product/unit/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
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
	indexContext, cancelIndexes := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelIndexes()
	if err := repo.EnsureIndexes(indexContext); err != nil {
		logger.GetLogger().Errorf("ensure product indexes: %v", err)
	}
	repoUnit := unitRepo.NewUnitRepository(pstmg)
	repomgCreditor := creditorepo.NewCreditorRepository(pstmg)
	repomgProductBarcode := productBarcodeRepo.NewProductBarcodeRepository(pstmg, cache)
	mq := cfg.MQConfig()
	prod := microservice.NewProducerWithTimeout(mq.URI(), mq.SecurityProtocol(), mq.SSLCAFile(), mq.SSLKeyFile(), mq.SSLCertFile(), ms.Logger, 30*time.Second)
	eventOutbox := outbox.New(pstmg)
	ms.RegisterBackgroundWorker(func(ctx context.Context) {
		defer prod.Close()
		eventOutbox.Run(ctx, prod.SendMessage, func(error) {
			logger.GetLogger().Warnf("Product outbox delivery pending; inspect pending event IDs and retry status")
		})
	})
	svc := services.NewProductHttpService(repo, repoUnit, *repomgCreditor, *repomgProductBarcode, eventOutbox)

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
	h.ms.POST("/product/resync", h.ResyncProduct)
	h.ms.GET("/product/:guid", h.InfoProduct)
	h.ms.PUT("/product/:guid", h.UpdateProduct)
	h.ms.DELETE("/product/:guid", h.DeleteProduct)
}

func requireProductBusinessCode(ctx microservice.IContext) (string, error) {
	businessCode := utils.NormalizeBusinessCode(ctx.UserInfo().BusinessCode)
	if businessCode == "" {
		return "", apperr.Respond(ctx, apperr.New(
			"COMPANY_REQUIRED",
			http.StatusConflict,
			"an active company is required",
			"กรุณาเลือกบริษัทก่อนใช้งานข้อมูลสินค้า",
		).WithField("businesscode"))
	}
	return businessCode, nil
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
	businessCode, err := requireProductBusinessCode(ctx)
	if err != nil {
		return err
	}
	pageable := utils.GetPageable(ctx.QueryParam)

	filters := h.searchFilter(ctx.QueryParam)

	docList, pagination, err := h.svc.ProductList(holdingCode, businessCode, filters, pageable)

	if err != nil {
		return apperr.RespondErr(ctx, err)
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
	businessCode, err := requireProductBusinessCode(ctx)
	if err != nil {
		return err
	}

	input := strings.TrimSpace(ctx.ReadInput())
	if input == "" {
		return apperr.Respond(ctx, apperr.ErrBadRequest.WithMessage("Invalid input: Empty request body"))
	}

	// ✅ แปลง JSON เป็น struct
	newProduct := &models.ProductDoc{}
	err = json.Unmarshal([]byte(input), newProduct)
	if err != nil {
		return apperr.Respond(ctx, apperr.ErrBadRequest.Withf("Invalid JSON format: %v", err))
	}

	// ✅ กำหนดค่า `HoldingCode` และ `GuidFixed`
	newProduct.HoldingCode = holdingCode
	newProduct.BusinessCode = businessCode
	newProduct.GuidFixed = utils.NewGUID()

	// ✅ ตรวจสอบ Validation
	if err = ctx.Validate(newProduct); err != nil {
		return apperr.Respond(ctx, apperr.ErrValidation.WithWrap(err).Withf("Validation failed: %v", err))
	}

	// ✅ กำหนดค่า `CreatedBy` และ `CreatedAt`
	newProduct.CreatedBy = userInfo.Username
	newProduct.CreatedAt = time.Now()

	// ✅ Debug
	fmt.Println("Creating Product:", newProduct)

	// ✅ เรียก Service เพื่อสร้าง Product
	err = h.svc.Create(newProduct)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return apperr.Respond(ctx, apperr.ErrDuplicate.
				WithField("code").
				WithMessage("product code already exists").
				WithThaiMessage("รหัสสินค้านี้ถูกใช้งานแล้ว กรุณาใช้รหัสอื่น").
				WithWrap(err))
		}
		return apperr.RespondErr(ctx, err)
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
	businessCode, err := requireProductBusinessCode(ctx)
	if err != nil {
		return err
	}

	if code == "" {
		return apperr.Respond(ctx, apperr.ErrValidation.WithField("guid").WithMessage("Product Code is required"))
	}

	product, err := h.svc.GetProduct(holdingCode, businessCode, code)
	if err != nil {
		return apperr.Respond(ctx, apperr.NotFound("Product").WithWrap(err))
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
	businessCode, err := requireProductBusinessCode(ctx)
	if err != nil {
		return err
	}

	if code == "" {
		return apperr.Respond(ctx, apperr.ErrValidation.WithField("guid").WithMessage("Product Code is required"))
	}

	input := strings.TrimSpace(ctx.ReadInput())
	if input == "" {
		return apperr.Respond(ctx, apperr.ErrBadRequest.WithMessage("Invalid input: Empty request body"))
	}

	// ✅ แปลง JSON เป็น struct
	updateData := &models.ProductDoc{}
	err = json.Unmarshal([]byte(input), updateData)
	if err != nil {
		return apperr.Respond(ctx, apperr.ErrBadRequest.Withf("Invalid JSON format: %v", err))
	}

	updateData.UpdatedBy = userInfo.Username
	updateData.UpdatedAt = time.Now()

	// ✅ Debug
	fmt.Println("Updating Product:", updateData)

	// ✅ อัปเดต Product
	docData, err := h.svc.Update(holdingCode, businessCode, code, userInfo.Username, updateData)
	if err != nil {
		return apperr.RespondErr(ctx, err)
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
	businessCode, err := requireProductBusinessCode(ctx)
	if err != nil {
		return err
	}

	if code == "" {
		return apperr.Respond(ctx, apperr.ErrValidation.WithField("guid").WithMessage("Product Code is required"))
	}

	err = h.svc.Delete(holdingCode, businessCode, code, userInfo.Username)
	if err != nil {
		return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Message: "Product deleted successfully",
	})
	return nil
}

// @Summary		Resync products into PostgreSQL projection
// @Description Replace the PostgreSQL product projection from active MongoDB products, then queue delivery through the outbox
// @Tags		Product
// @Produce 	json
// @Success		200 {object} common.ApiResponse
// @Failure		400 {object} common.ApiResponse
// @Security	AccessToken
// @Router		/product/resync [post]
func (h ProductHttp) ResyncProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	businessCode, err := requireProductBusinessCode(ctx)
	if err != nil {
		return err
	}

	rebuilt, err := build.ProcessProductRebuildCompany(holdingCode, businessCode)
	if err != nil {
		return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err).WithMessage("product rebuild failed"))
	}
	queued, err := h.svc.Resync(holdingCode, businessCode)
	if err != nil {
		return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err).WithMessage("product resync failed"))
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Message: fmt.Sprintf("rebuilt %d products and queued %d events for delivery", rebuilt, queued),
		// Retain the legacy field: this request makes no synchronous Kafka sends.
		Data: map[string]int{"rebuilt": rebuilt, "queued": queued, "published": 0},
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
			Param: "itemtype",
			Field: "itemtype",
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
