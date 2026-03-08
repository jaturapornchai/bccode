package bom

import (
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/product/bom/repositories"
	"smlcloudplatform/internal/product/bom/services"
	product_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	saleinvoicebom_repositories "smlcloudplatform/internal/transaction/saleinvoicebomprice/repositories"
	saleinvoicebom_services "smlcloudplatform/internal/transaction/saleinvoicebomprice/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

type IBOMHttp interface{}

type BOMHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IBOMHttpService
}

func NewBOMHttp(ms *microservice.Microservice, cfg config.IConfig) BOMHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	producer := ms.Producer(cfg.MQConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewBomRepository(pst)
	repoMq := repositories.NewBomMessageQueueRepository(producer)
	repoSaleBomMq := saleinvoicebom_repositories.NewSaleInvoiceBOMPriceMessageQueueRepository(producer)

	repoProduct := product_repositories.NewProductBarcodeRepository(pst, cache)

	repoSaleinvoiceBom := saleinvoicebom_repositories.NewSaleInvoiceBomPriceRepository(pst)
	svcSaleinvoiceBom := saleinvoicebom_services.NewSaleInvoiceBomPriceService(repoSaleBomMq, repoSaleinvoiceBom)

	svc := services.NewBOMHttpService(repo, repoMq, repoProduct, svcSaleinvoiceBom)

	return BOMHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h BOMHttp) RegisterHttp() {

	h.ms.GET("/product/bom", h.SearchBOMPage)
	h.ms.GET("/product/bom/list", h.SearchBOMStep)
	h.ms.POST("/product/bom", h.CreateBOM)
	h.ms.GET("/product/bom/:id", h.InfoBOM)
	h.ms.DELETE("/product/bom/:id", h.DeleteBOM)
}

// Create BOM godoc
// @Description Create BOM
// @Tags		BOM
// @Param		barcode		query	string		false  "Barcode"
// @Param		docNo		query	string		false  "DocNo"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/bom [post]
func (h BOMHttp) CreateBOM(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID

	barcode := ctx.QueryParam("barcode")

	if barcode == "" {
		ctx.ResponseError(http.StatusBadRequest, "barcode is empty")
		return nil
	}

	docNo := ctx.QueryParam("docNo")

	if barcode == "" {
		ctx.ResponseError(http.StatusBadRequest, "barcode is empty")
		return nil
	}

	idx, err := h.svc.UpsertBOM(shopID, authUsername, docNo, barcode)

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

// Delete BOM godoc
// @Description Delete BOM
// @Tags		BOM
// @Param		id  path      string  true  "BOM ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/bom/{id} [delete]
func (h BOMHttp) DeleteBOM(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteBOM(shopID, id, authUsername)

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

// Get BOM godoc
// @Description get BOM info by guidfixed
// @Tags		BOM
// @Param		id  path      string  true  "BOM guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/bom/{id} [get]
func (h BOMHttp) InfoBOM(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get BOM %v", id)
	doc, err := h.svc.InfoBOM(shopID, id)

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

// List BOM step godoc
// @Description get list step
// @Tags		BOM
// @Param		barcode		query	string		false  "Barcode"
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/bom [get]
func (h BOMHttp) SearchBOMPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	filters := h.searchFilter(ctx.QueryParam)

	docList, pagination, err := h.svc.SearchBOM(shopID, filters, pageable)

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

// List BOM godoc
// @Description search limit offset
// @Tags		BOM
// @Param		barcode		query	string		false  "Barcode"
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/bom/list [get]
func (h BOMHttp) SearchBOMStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	filters := h.searchFilter(ctx.QueryParam)

	docList, total, err := h.svc.SearchBOMStep(shopID, lang, filters, pageableStep)

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

func (h BOMHttp) searchFilter(queryParam func(string) string) map[string]interface{} {
	filters := requestfilter.GenerateFilters(queryParam, []requestfilter.FilterRequest{
		{
			Param: "barcode",
			Field: "barcode",
			Type:  requestfilter.FieldTypeString,
		},
	})

	return filters
}
