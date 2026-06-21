package productbarcode

import (
	"context"
	"net/http"
	authmodels "smlcloudplatform/internal/authentication/models"
	common "smlcloudplatform/internal/models"
	phmodels "smlcloudplatform/internal/product/productbarcode/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	docs = h.enrichPriceHistoryAvatars(docs)

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

	docs = h.enrichPriceHistoryAvatars(docs)

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
			Param: "keynumber",
			Field: "keynumber",
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

// enrichPriceHistoryAvatars เติม CreatedByAvatarThumb ให้แต่ละ record
// จาก users collection (lookup ครั้งเดียวแบบ batch ด้วย $in). best-effort:
// ถ้า lookup ล้มเหลวจะคืน docs เดิมโดยไม่มี avatar (ไม่ทำให้ request พัง).
func (h ProductBarcodeHttp) enrichPriceHistoryAvatars(docs []phmodels.ProductPriceHistoryInfo) []phmodels.ProductPriceHistoryInfo {
	if len(docs) == 0 {
		return docs
	}

	usernameSet := map[string]struct{}{}
	for _, d := range docs {
		if d.CreatedBy != "" {
			usernameSet[d.CreatedBy] = struct{}{}
		}
	}
	if len(usernameSet) == 0 {
		return docs
	}

	usernames := make([]string, 0, len(usernameSet))
	for u := range usernameSet {
		usernames = append(usernames, u)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pst := h.ms.MongoPersister(h.cfg.MongoPersisterConfig())
	var users []struct {
		Username    string `bson:"username"`
		Avatar      string `bson:"avatar"`
		AvatarThumb string `bson:"avatarthumb"`
	}
	projection := options.Find().SetProjection(bson.M{"username": 1, "avatar": 1, "avatarthumb": 1})
	if err := pst.Find(ctx, &authmodels.UserDoc{}, bson.M{"username": bson.M{"$in": usernames}}, &users, projection); err != nil {
		return docs
	}

	avatarByUsername := make(map[string]string, len(users))
	for _, u := range users {
		thumb := u.AvatarThumb
		if thumb == "" {
			thumb = u.Avatar
		}
		avatarByUsername[u.Username] = thumb
	}

	for i := range docs {
		docs[i].CreatedByAvatarThumb = avatarByUsername[docs[i].CreatedBy]
	}

	return docs
}
