package handlers

// POST /goapi/product/v2/item/update-listing — แก้ชั้นลงขายระดับสินค้า (§5.2)
//
// หลักการ: เขียนด้วย $set เฉพาะ path ที่อนุญาต + ตรวจ __v (ไม่เขียนทับทั้งเอกสาร)
// - v2 ไม่ส่ง Kafka (projection PG/ClickHouse มีเฉพาะคอลัมน์ชั้นบัญชี)
// - การบันทึกฝั่งบัญชีเดิมไม่เพิ่ม __v ดังนั้นการแก้ชั้นบัญชีที่แทรกระหว่าง item/get → update-listing จะไม่ถูกจับ
//   (ยอมรับได้ เพราะ v2 แตะเฉพาะฟิลด์ชั้นลงขาย)

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	productmodels "smlcloudplatform/internal/product/product/models"
	msmodels "smlcloudplatform/pkg/microservice/models"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ProductV2ItemUpdateListingHandler(c echo.Context) error {
	productV2LimitBody(c)
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		if productV2IsBodyTooLarge(err) {
			return productV2Error(c, http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE", productV2MsgBodyTooLarge, nil)
		}
		return productV2Error(c, http.StatusBadRequest, "INVALID_JSON", productV2MsgInvalidJSON, nil)
	}
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil || raw == nil {
		return productV2Error(c, http.StatusBadRequest, "INVALID_JSON", productV2MsgInvalidJSON, nil)
	}
	holdingCode, businessCode, cerr := authenticatedCompanyContext(c, productV2MapString(raw, "holdingcode"), productV2MapString(raw, "businesscode"))
	if cerr != nil {
		return productV2CompanyError(c, cerr)
	}
	userInfo, _ := c.Get("UserInfo").(msmodels.UserInfo)

	// E24 ก่อนทุกอย่าง — ฟิลด์ชั้นบัญชีต้องถูกปฏิเสธพร้อมรายชื่อ ห้ามหลุดแบบเงียบ ๆ
	accounting, unknown := scanProductV2Payload(raw)
	if len(accounting) > 0 {
		message, fields := productV2AccountingFieldErrors(accounting)
		return productV2Error(c, http.StatusUnprocessableEntity, "ACCOUNTING_FIELD_READONLY", message, fields)
	}
	if len(unknown) > 0 {
		return productV2Error(c, http.StatusUnprocessableEntity, "VALIDATION_FAILED", productV2MsgValidationFailed, productV2UnknownFieldErrors(unknown))
	}
	if productV2HasListingTiers(raw) {
		return productV2Error(c, http.StatusUnprocessableEntity, "VALIDATION_FAILED", productV2MsgValidationFailed, []productV2FieldError{{"listing.tiers", "READONLY", productV2MsgTiersReadOnly}})
	}

	var req productV2UpdateListingRequest
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		return productV2Error(c, http.StatusUnprocessableEntity, "VALIDATION_FAILED", productV2MsgValidationFailed, []productV2FieldError{productV2DecodeError(err)})
	}
	req.ItemCode = strings.TrimSpace(req.ItemCode)
	if req.ItemCode == "" {
		return productV2Error(c, http.StatusUnprocessableEntity, "VALIDATION_FAILED", productV2MsgValidationFailed, []productV2FieldError{{"itemcode", "REQUIRED", "กรุณาระบุรหัสสินค้า"}})
	}
	if req.Version == nil {
		return productV2Error(c, http.StatusUnprocessableEntity, "VALIDATION_FAILED", productV2MsgValidationFailed, []productV2FieldError{{"__v", "REQUIRED", "กรุณาส่งเลขเวอร์ชันของเอกสาร (__v)"}})
	}

	ctx, cancel := context.WithTimeout(context.Background(), productV2Timeout)
	defer cancel()

	allowed, err := productV2CheckPermission(ctx, userInfo, productV2PermissionKey)
	if err != nil {
		logger.Error("ProductV2ItemUpdateListing: permission: %v", err)
		return productV2Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", productV2MsgPermissionLoad, nil)
	}
	if !allowed {
		return productV2Error(c, http.StatusForbidden, "FORBIDDEN", productV2MsgForbidden, nil)
	}

	_, db := GetAtlasConnection()
	if db == nil {
		return productV2Error(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", productV2MsgDBUnavailable, nil)
	}
	products := db.Collection(productV2ProductsCollection)
	itemFilter := productV2ItemFilter(holdingCode, businessCode, req.ItemCode)

	var current productV2ProductDoc
	if err := products.FindOne(ctx, itemFilter).Decode(&current); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return productV2NotFound(c)
		}
		logger.Error("ProductV2ItemUpdateListing: load product: %v", err)
		return productV2Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", productV2MsgInternal, nil)
	}

	var modelPrices []float64
	if req.Listing != nil && req.Listing.Wholesale != nil {
		modelPrices, err = loadProductV2ModelPrices(ctx, db, holdingCode, businessCode, req.ItemCode)
		if err != nil {
			logger.Error("ProductV2ItemUpdateListing: load model prices: %v", err)
			return productV2Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", productV2MsgInternal, nil)
		}
	}

	if fields := validateProductV2Update(req, modelPrices, productV2DefaultLimits); len(fields) > 0 {
		code, message := "VALIDATION_FAILED", productV2MsgValidationFailed
		if len(fields) == 1 && fields[0].Code == "PACKAGE_INCOMPLETE" {
			code, message = "PACKAGE_INCOMPLETE", productV2MsgPackageIncomplete
		}
		return productV2Error(c, http.StatusUnprocessableEntity, code, message, fields)
	}

	now := time.Now().UTC()
	set := buildProductV2Set(current.ProductDoc, req, userInfo.Username, now)
	filter := bson.M{}
	for k, v := range itemFilter {
		filter[k] = v
	}
	for k, v := range productV2VersionFilter(*req.Version) {
		filter[k] = v
	}
	result, err := products.UpdateOne(ctx, filter, bson.M{"$set": set, "$inc": bson.M{"__v": 1}})
	if err != nil {
		logger.Error("ProductV2ItemUpdateListing: update: %v", err)
		return productV2Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", productV2MsgInternal, nil)
	}
	if result.MatchedCount == 0 {
		// ช่องว่างเล็กน้อยระหว่าง UpdateOne กับ CountDocuments ยอมรับได้ (แค่แยก 404 กับ 409)
		count, err := products.CountDocuments(ctx, itemFilter, options.Count().SetLimit(1))
		if err != nil {
			logger.Error("ProductV2ItemUpdateListing: count: %v", err)
			return productV2Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", productV2MsgInternal, nil)
		}
		status, code := classifyProductV2UpdateResult(result.MatchedCount, count > 0)
		if code == "VERSION_CONFLICT" {
			return productV2Error(c, status, code, productV2MsgVersionConflict, nil)
		}
		return productV2NotFound(c)
	}

	merged := applyProductV2Patch(current.ProductDoc, req)
	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"itemcode":  req.ItemCode,
			"__v":       *req.Version + 1,
			"readiness": map[string]any{"listing": productV2ListingReadiness(merged)},
		},
	})
}

func productV2MapString(raw map[string]any, key string) string {
	if value, ok := raw[key].(string); ok {
		return value
	}
	return ""
}

func productV2ItemFilter(holdingCode, businessCode, itemCode string) bson.M {
	return bson.M{
		"holdingcode":  holdingCode,
		"businesscode": businessCode,
		"code":         itemCode,
		"deletedat":    nil,
	}
}

// productV2VersionFilter — เงื่อนไข __v; เอกสารเก่าที่ยังไม่มี __v ถือเป็นเวอร์ชัน 0
func productV2VersionFilter(version int64) bson.M {
	if version == 0 {
		return bson.M{"$or": bson.A{bson.M{"__v": int64(0)}, bson.M{"__v": bson.M{"$exists": false}}}}
	}
	return bson.M{"__v": version}
}

// classifyProductV2UpdateResult — 200 / 404 ITEM_NOT_FOUND / 409 VERSION_CONFLICT
func classifyProductV2UpdateResult(matched int64, exists bool) (int, string) {
	switch {
	case matched > 0:
		return http.StatusOK, ""
	case exists:
		return http.StatusConflict, "VERSION_CONFLICT"
	default:
		return http.StatusNotFound, "ITEM_NOT_FOUND"
	}
}

// productV2HasListingTiers — listing.tiers แก้ผ่าน tier/* เท่านั้น (ต้อง sync productbarcode.listing.tierindex ใน transaction)
func productV2HasListingTiers(raw map[string]any) bool {
	listing, ok := raw["listing"].(map[string]any)
	if !ok {
		return false
	}
	_, has := listing["tiers"]
	return has
}

// productV2UnknownFieldErrors — key ที่ไม่รู้จัก; packageweight/... ชี้ทางให้ส่งผ่าน package{} แทน
func productV2UnknownFieldErrors(keys []string) []productV2FieldError {
	fields := make([]productV2FieldError, 0, len(keys))
	for _, key := range keys {
		if _, isPackage := productV2PackageTopKeys[key]; isPackage {
			fields = append(fields, productV2FieldError{Field: key, Code: "INVALID_FIELD", Message: productV2MsgPackageViaObject})
			continue
		}
		fields = append(fields, productV2FieldError{Field: key, Code: "UNKNOWN_FIELD", Message: "ไม่รู้จักฟิลด์นี้"})
	}
	return fields
}

// productV2DecodeError — แปลง error ของ json decoder เป็นรายการฟิลด์ (unknown field "x" → UNKNOWN_FIELD)
func productV2DecodeError(err error) productV2FieldError {
	const marker = "unknown field "
	msg := err.Error()
	if idx := strings.Index(msg, marker); idx >= 0 {
		field := strings.Trim(msg[idx+len(marker):], "\"")
		return productV2FieldError{Field: field, Code: "UNKNOWN_FIELD", Message: "ไม่รู้จักฟิลด์นี้"}
	}
	var unmarshalErr *json.UnmarshalTypeError
	if errors.As(err, &unmarshalErr) {
		field := unmarshalErr.Field
		if strings.HasSuffix(field, "unitprice") {
			return productV2FieldError{Field: field, Code: "INVALID_VALUE", Message: "ราคาต้องเป็นตัวเลขทศนิยม เช่น 199.00 หรือ \"199.00\""}
		}
		return productV2FieldError{Field: field, Code: "INVALID_VALUE", Message: "ชนิดข้อมูลไม่ถูกต้อง"}
	}
	if strings.Contains(msg, "Decimal128") {
		return productV2FieldError{Field: "listing.wholesale.unitprice", Code: "INVALID_VALUE", Message: "ราคาต้องเป็นตัวเลขทศนิยม เช่น 199.00 หรือ \"199.00\""}
	}
	return productV2FieldError{Field: "", Code: "INVALID_VALUE", Message: productV2MsgInvalidJSON}
}

// loadProductV2ModelPrices — ราคาปกติ (keynumber 1 ถ้ามี ไม่งั้น keynumber ต่ำสุด) ของทุกบาร์โค้ดในสินค้า ใช้ตรวจ E07
func loadProductV2ModelPrices(ctx context.Context, db *mongo.Database, holdingCode, businessCode, itemCode string) ([]float64, error) {
	cursor, err := db.Collection(productV2BarcodesCollection).Find(ctx, bson.M{
		"holdingcode":  holdingCode,
		"businesscode": businessCode,
		"itemcode":     itemCode,
		"deletedat":    nil,
	}, options.Find().SetProjection(bson.M{"prices": 1}).SetLimit(500))
	if err != nil {
		return nil, err
	}
	var docs []struct {
		Prices []productV2ModelPrice `bson:"prices"`
	}
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	prices := make([]float64, 0, len(docs))
	for _, doc := range docs {
		if price, ok := productV2PickNormalPrice(doc.Prices); ok {
			prices = append(prices, price)
		}
	}
	return prices, nil
}

type productV2ModelPrice struct {
	KeyNumber int     `bson:"keynumber" json:"keynumber"`
	Price     float64 `bson:"price" json:"price"`
}

func productV2PickNormalPrice(prices []productV2ModelPrice) (float64, bool) {
	found, picked := false, productV2ModelPrice{}
	for _, p := range prices {
		if p.KeyNumber == 1 {
			return p.Price, true
		}
		if !found || p.KeyNumber < picked.KeyNumber {
			found, picked = true, p
		}
	}
	return picked.Price, found
}

// mergeListing — รวม patch เข้ากับ listing เดิม (ฟิลด์ที่ไม่ส่งมาคงค่าเดิม); nil current = เริ่มจากค่าว่าง
func mergeListing(current *productmodels.ProductListing, patch *productV2ListingPatch) productmodels.ProductListing {
	merged := productmodels.ProductListing{}
	if current != nil {
		merged = *current
	}
	if patch == nil {
		return productV2NormalizeListing(merged)
	}
	if patch.Title != nil {
		merged.Title = strings.TrimSpace(*patch.Title)
	}
	if patch.Description != nil {
		merged.Description = *patch.Description
	}
	if patch.Condition != nil {
		merged.Condition = *patch.Condition
	}
	if patch.Preorder != nil {
		preorder := *patch.Preorder
		if !preorder.IsPreorder {
			preorder.DaysToShip = 0
		}
		merged.Preorder = &preorder
	}
	if patch.PurchaseLimit != nil {
		limit := *patch.PurchaseLimit
		merged.PurchaseLimit = &limit
	}
	if patch.Wholesale != nil {
		merged.Wholesale = *patch.Wholesale
	}
	if patch.SizeChart != nil {
		chart := *patch.SizeChart
		merged.SizeChart = &chart
	}
	return productV2NormalizeListing(merged)
}

// productV2NormalizeListing — รูปทรงคงที่ตาม §5.1: array ไม่เป็น nil, preorder/purchaselimit มีเสมอ, condition ว่าง → NEW (ค่าเริ่มต้นตามแบบจำลอง)
func productV2NormalizeListing(listing productmodels.ProductListing) productmodels.ProductListing {
	if listing.Condition == "" {
		listing.Condition = productmodels.ListingConditionNew
	}
	if listing.Preorder == nil {
		listing.Preorder = &productmodels.ProductListingPreorder{}
	}
	if listing.PurchaseLimit == nil {
		listing.PurchaseLimit = &productmodels.ProductListingPurchaseLimit{}
	}
	if listing.Wholesale == nil {
		listing.Wholesale = []productmodels.ProductListingWholesale{}
	}
	if listing.Tiers == nil {
		listing.Tiers = []productmodels.ProductListingTier{}
	}
	return listing
}

// applyProductV2Patch — เอกสารหลังแก้ (ใช้คำนวณ readiness ตอบกลับ)
func applyProductV2Patch(current productmodels.ProductDoc, req productV2UpdateListingRequest) productmodels.ProductDoc {
	if req.Listing != nil {
		listing := mergeListing(current.Listing, req.Listing)
		current.Listing = &listing
	}
	if req.Package != nil {
		current.PackageWeight = productV2Float(req.Package.Weight)
		current.PackageLength = productV2Float(req.Package.Length)
		current.PackageWidth = productV2Float(req.Package.Width)
		current.PackageHeight = productV2Float(req.Package.Height)
	}
	if req.Description != nil {
		current.Description = *req.Description
	}
	if req.ImageURI != nil {
		current.ImageURI = *req.ImageURI
	}
	if req.ImageURIThumb != nil {
		current.ImageURIThumb = *req.ImageURIThumb
	}
	if req.Images != nil {
		images := *req.Images
		current.Images = &images
	}
	if req.Videos != nil {
		videos := *req.Videos
		current.Videos = &videos
	}
	return current
}

func productV2Float(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
}

// buildProductV2Set — สร้าง $set เฉพาะ path ที่อนุญาต; `listing` เขียนทั้ง sub-document ที่ merge แล้ว
// (ไม่ใช้ dotted path ใต้ listing เพื่อเลี่ยง error "cannot create field in null parent")
func buildProductV2Set(current productmodels.ProductDoc, req productV2UpdateListingRequest, username string, now time.Time) bson.M {
	set := bson.M{}
	if req.Listing != nil {
		set["listing"] = mergeListing(current.Listing, req.Listing)
	}
	if req.Package != nil {
		set["packageweight"] = productV2Float(req.Package.Weight)
		set["packagelength"] = productV2Float(req.Package.Length)
		set["packagewidth"] = productV2Float(req.Package.Width)
		set["packageheight"] = productV2Float(req.Package.Height)
	}
	if req.Description != nil {
		set["description"] = *req.Description
	}
	if req.ImageURI != nil {
		set["imageuri"] = *req.ImageURI
	}
	if req.ImageURIThumb != nil {
		set["imageurithumb"] = *req.ImageURIThumb
	}
	if req.Images != nil {
		set["images"] = *req.Images
	}
	if req.Videos != nil {
		set["videos"] = *req.Videos
	}
	set["updatedat"] = now
	set["updatedby"] = username
	return set
}
