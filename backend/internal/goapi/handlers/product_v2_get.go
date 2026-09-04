package handlers

// POST /goapi/product/v2/item/get — อ่านสินค้า 1 ตัว ทั้งชั้นบัญชี (อ่านอย่างเดียว) และชั้นลงขาย + ตัวเลือกทั้งหมด (§5.1)

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/models"
	productmodels "smlcloudplatform/internal/product/product/models"
	barcodemodels "smlcloudplatform/internal/product/productbarcode/models"
	msmodels "smlcloudplatform/pkg/microservice/models"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// productV2ModelDoc — เฉพาะฟิลด์ของ productbarcode ที่ item/get ต้องใช้
type productV2ModelDoc struct {
	Barcode       string                               `bson:"barcode"`
	ItemUnitCode  string                               `bson:"itemunitcode"`
	Prices        []productV2ModelPrice                `bson:"prices"`
	Listing       *barcodemodels.ProductBarcodeListing `bson:"listing"`
	ImageURIThumb string                               `bson:"imageurithumb"`
	PackageWeight float64                              `bson:"packageweight"`
	PackageLength float64                              `bson:"packagelength"`
	PackageWidth  float64                              `bson:"packagewidth"`
	PackageHeight float64                              `bson:"packageheight"`
}

func ProductV2ItemGetHandler(c echo.Context) error {
	productV2LimitBody(c)
	var req productV2ItemGetRequest
	if err := c.Bind(&req); err != nil {
		if productV2IsBodyTooLarge(err) {
			return productV2Error(c, http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE", productV2MsgBodyTooLarge, nil)
		}
		return productV2Error(c, http.StatusBadRequest, "INVALID_JSON", productV2MsgInvalidJSON, nil)
	}
	holdingCode, businessCode, cerr := authenticatedCompanyContext(c, req.HoldingCode, req.BusinessCode)
	if cerr != nil {
		return productV2CompanyError(c, cerr)
	}
	userInfo, _ := c.Get("UserInfo").(msmodels.UserInfo)
	req.ItemCode = strings.TrimSpace(req.ItemCode)
	if req.ItemCode == "" {
		return productV2Error(c, http.StatusUnprocessableEntity, "VALIDATION_FAILED", productV2MsgValidationFailed, []productV2FieldError{{"itemcode", "REQUIRED", "กรุณาระบุรหัสสินค้า"}})
	}

	ctx, cancel := context.WithTimeout(context.Background(), productV2Timeout)
	defer cancel()

	allowed, err := productV2CheckPermission(ctx, userInfo, productV2PermissionKey)
	if err != nil {
		logger.Error("ProductV2ItemGet: permission: %v", err)
		return productV2Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", productV2MsgPermissionLoad, nil)
	}
	if !allowed {
		return productV2Error(c, http.StatusForbidden, "FORBIDDEN", productV2MsgForbidden, nil)
	}

	_, db := GetAtlasConnection()
	if db == nil {
		return productV2Error(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", productV2MsgDBUnavailable, nil)
	}

	var doc productV2ProductDoc
	if err := db.Collection(productV2ProductsCollection).FindOne(ctx, productV2ItemFilter(holdingCode, businessCode, req.ItemCode)).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return productV2NotFound(c)
		}
		logger.Error("ProductV2ItemGet: load product: %v", err)
		return productV2Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", productV2MsgInternal, nil)
	}

	cursor, err := db.Collection(productV2BarcodesCollection).Find(ctx, bson.M{
		"holdingcode":  holdingCode,
		"businesscode": businessCode,
		"itemcode":     req.ItemCode,
		"deletedat":    nil,
	}, options.Find().
		SetProjection(bson.M{"barcode": 1, "itemunitcode": 1, "prices": 1, "listing": 1, "imageurithumb": 1, "packageweight": 1, "packagelength": 1, "packagewidth": 1, "packageheight": 1}).
		SetSort(bson.D{{Key: "barcode", Value: 1}}).
		SetLimit(500))
	if err != nil {
		logger.Error("ProductV2ItemGet: load models: %v", err)
		return productV2Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", productV2MsgInternal, nil)
	}
	var modelDocs []productV2ModelDoc
	if err := cursor.All(ctx, &modelDocs); err != nil {
		logger.Error("ProductV2ItemGet: decode models: %v", err)
		return productV2Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", productV2MsgInternal, nil)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"data":    buildProductV2GetResponse(doc, modelDocs),
	})
}

// buildProductV2GetResponse — ประกอบ data ตาม §5.1 (ตรรกะล้วน ทดสอบได้)
func buildProductV2GetResponse(doc productV2ProductDoc, modelDocs []productV2ModelDoc) map[string]any {
	product := doc.ProductDoc
	listing := mergeListing(product.Listing, nil)
	images := []productmodels.ProductImage{}
	if product.Images != nil {
		images = *product.Images
	}
	videos := []productmodels.ProductVideo{}
	if product.Videos != nil {
		videos = *product.Videos
	}

	modelsOut := make([]map[string]any, 0, len(modelDocs))
	for _, m := range modelDocs {
		tierIndex := []int{}
		gtin, isForSale := "", false
		if m.Listing != nil {
			if m.Listing.TierIndex != nil {
				tierIndex = m.Listing.TierIndex
			}
			gtin, isForSale = m.Listing.GTIN, m.Listing.IsForSale
		}
		price := ""
		if value, ok := productV2PickNormalPrice(m.Prices); ok {
			price = fmt.Sprintf("%.2f", value)
		}
		var pkg any
		if m.PackageWeight != 0 || m.PackageLength != 0 || m.PackageWidth != 0 || m.PackageHeight != 0 {
			pkg = map[string]any{"weight": m.PackageWeight, "length": m.PackageLength, "width": m.PackageWidth, "height": m.PackageHeight}
		}
		modelsOut = append(modelsOut, map[string]any{
			"barcode":       m.Barcode,
			"tierindex":     tierIndex,
			"itemunitcode":  m.ItemUnitCode,
			"price":         price,
			"gtin":          gtin,
			"isforsale":     isForSale,
			"imageurithumb": m.ImageURIThumb,
			"package":       pkg,
		})
	}

	return map[string]any{
		"accounting": map[string]any{
			"code":         product.Code,
			"names":        productV2Names(product.Names),
			"unitcode":     product.UnitCode,
			"itemtype":     product.ItemType,
			"categorycode": product.CategoryCode,
			"brandcode":    product.BrandCode,
			"isactive":     productV2IsActive(doc),
			"readonly":     true,
		},
		"listing": listing,
		"package": map[string]any{
			"weight": product.PackageWeight,
			"length": product.PackageLength,
			"width":  product.PackageWidth,
			"height": product.PackageHeight,
		},
		"media": map[string]any{
			"imageuri":      product.ImageURI,
			"imageurithumb": product.ImageURIThumb,
			"images":        images,
			"videos":        videos,
		},
		"models": modelsOut,
		"__v":    doc.Version,
	}
}

// productV2IsActive — ใช้ isactive ในเอกสาร (Product struct ไม่มีฟิลด์นี้) ถ้าไม่มีใช้สถานะยังไม่ถูกลบแทน
func productV2IsActive(doc productV2ProductDoc) bool {
	if doc.IsActive != nil {
		return *doc.IsActive
	}
	return doc.DeletedAt.IsZero()
}

func productV2Names(names *[]models.NameX) []models.NameX {
	if names == nil {
		return []models.NameX{}
	}
	return *names
}
