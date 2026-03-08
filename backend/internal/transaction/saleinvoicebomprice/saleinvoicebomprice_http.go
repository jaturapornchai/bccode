package saleinvoicebomprice

import (
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/saleinvoicebomprice/repositories"
	"smlcloudplatform/internal/transaction/saleinvoicebomprice/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

type ISaleInvoiceBomPriceHttp interface{}

type SaleInvoiceBomPriceHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.ISaleInvoiceBomPriceService
}

func NewSaleInvoiceBomPriceHttp(ms *microservice.Microservice, cfg config.IConfig) SaleInvoiceBomPriceHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())

	repo := repositories.NewSaleInvoiceBomPriceRepository(pst)
	repoMq := repositories.NewSaleInvoiceBOMPriceMessageQueueRepository(ms.Producer(cfg.MQConfig()))

	svc := services.NewSaleInvoiceBomPriceService(repoMq, repo)

	return SaleInvoiceBomPriceHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h SaleInvoiceBomPriceHttp) RegisterHttp() {

	h.ms.GET("/transaction/sale-invoice-price", h.SearchSaleInvoiceBomPricePage)
	h.ms.GET("/transaction/sale-invoice-price/list", h.SearchSaleInvoiceBomPriceStep)
	h.ms.GET("/transaction/sale-invoice-price/:id", h.InfoSaleInvoiceBomPrice)
	h.ms.GET("/transaction/sale-invoice-price/docno/:docno", h.InfoSaleInvoiceBomPriceByDocNo)
}

// Get SaleInvoiceBomPrice godoc
// @Description get SaleInvoiceBomPrice info by guidfixed
// @Tags		SaleInvoiceBomPrice
// @Param		id  path      string  true  "SaleInvoiceBomPrice guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/sale-invoice-price/{id} [get]
func (h SaleInvoiceBomPriceHttp) InfoSaleInvoiceBomPrice(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get SaleInvoiceBomPrice %v", id)
	doc, err := h.svc.InfoSaleInvoiceBomPrice(shopID, id)

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

// Get SaleInvoiceBomPrice godoc
// @Description get SaleInvoiceBomPrice info by guidfixed
// @Tags		SaleInvoiceBomPrice
// @Param		id  path      string  true  "SaleInvoiceBomPrice guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/sale-invoice-price/docno/{docno} [get]
func (h SaleInvoiceBomPriceHttp) InfoSaleInvoiceBomPriceByDocNo(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	docNo := ctx.Param("docno")

	docs, err := h.svc.InfoSaleInvoiceBomPriceByDocNo(shopID, docNo)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    docs,
	})
	return nil
}

// List SaleInvoiceBomPrice step godoc
// @Description get list step
// @Tags		SaleInvoiceBomPrice
// @Param		barcode		query	string		false  "Barcode"
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/sale-invoice-price [get]
func (h SaleInvoiceBomPriceHttp) SearchSaleInvoiceBomPricePage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	filters := h.searchFilter(ctx.QueryParam)

	docList, pagination, err := h.svc.SearchSaleInvoiceBomPrice(shopID, filters, pageable)

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

// List SaleInvoiceBomPrice godoc
// @Description search limit offset
// @Tags		SaleInvoiceBomPrice
// @Param		barcode		query	string		false  "Barcode"
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/sale-invoice-price/list [get]
func (h SaleInvoiceBomPriceHttp) SearchSaleInvoiceBomPriceStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	filters := h.searchFilter(ctx.QueryParam)

	docList, total, err := h.svc.SearchSaleInvoiceBomPriceStep(shopID, lang, filters, pageableStep)

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

func (h SaleInvoiceBomPriceHttp) searchFilter(queryParam func(string) string) map[string]interface{} {
	filters := requestfilter.GenerateFilters(queryParam, []requestfilter.FilterRequest{
		{
			Param: "barcode",
			Field: "barcode",
			Type:  requestfilter.FieldTypeString,
		},
	})

	return filters
}
