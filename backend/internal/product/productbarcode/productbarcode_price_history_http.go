package productbarcode

import (
	"net/http"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

// Get ProductBarcode Price History godoc
// @Description Get ProductBarcode Price History
// @Tags		ProductBarcode
// @Param		barcode		query	string		false  "Barcode filter"
// @Param		keynumber	query	string		false  "Key number filter"
// @Param		action		query	string		false  "Action filter (create/update)"
// @Param		createdby	query	string		false  "Created by filter"
// @Param		fromdate	query	string		false  "From date (YYYY-MM-DD)"
// @Param		todate		query	string		false  "To date (YYYY-MM-DD)"
// @Param		page		query	integer		false  "Page"
// @Param		limit		query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/barcode/price-history [get]
func (h ProductBarcodeHttp) GetPriceHistory(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageable := utils.GetPageable(ctx.QueryParam)

	filters := h.priceHistoryFilter(ctx.QueryParam)

	docs, pagination, err := h.svc.GetPriceHistory(holdingCode, filters, pageable)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success:    true,
		Data:       docs,
		Pagination: pagination,
	})
	return nil
}

// Get ProductBarcode Price History By Barcode godoc
// @Description Get ProductBarcode Price History By Barcode
// @Tags		ProductBarcode
// @Param		barcode		path	string		true   "Barcode"
// @Param		page		query	integer		false  "Page"
// @Param		limit		query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /product/barcode/price-history/{barcode} [get]
func (h ProductBarcodeHttp) GetPriceHistoryByBarcode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	barcode := ctx.Param("barcode")
	pageable := utils.GetPageable(ctx.QueryParam)

	docs, pagination, err := h.svc.GetPriceHistoryByBarcode(holdingCode, barcode, pageable)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success:    true,
		Data:       docs,
		Pagination: pagination,
	})
	return nil
}

func (h ProductBarcodeHttp) priceHistoryFilter(queryParam func(string) string) map[string]interface{} {
	filters := requestfilter.GenerateFilters(queryParam, []requestfilter.FilterRequest{
		{
			Param: "barcode",
			Field: "barcode",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "key_number",
			Field: "key_number",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "action",
			Field: "action",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "createdby",
			Field: "createdby",
			Type:  requestfilter.FieldTypeString,
		},
	})

	// Date range filters
	if fromDate := queryParam("fromdate"); fromDate != "" {
		if toDate := queryParam("todate"); toDate != "" {
			filters["createdat"] = map[string]interface{}{
				"$gte": fromDate + "T00:00:00Z",
				"$lte": toDate + "T23:59:59Z",
			}
		} else {
			filters["createdat"] = map[string]interface{}{
				"$gte": fromDate + "T00:00:00Z",
			}
		}
	} else if toDate := queryParam("todate"); toDate != "" {
		filters["createdat"] = map[string]interface{}{
			"$lte": toDate + "T23:59:59Z",
		}
	}

	return filters
}
