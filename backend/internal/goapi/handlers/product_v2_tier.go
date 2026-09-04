package handlers

// POST /goapi/product/v2/tier/init และ POST /goapi/product/v2/tier/update — ชั้นตัวเลือกสินค้า (§4.2, §5.4)
//
// หลักการ
//   - ทั้งสอง endpoint รับ "สถานะสุดท้ายทั้งชุด" (tiers + models) เหมือนกัน ผู้เรียกเป็นผู้ระบุว่าบาร์โค้ดใดอยู่ชุดผสมใด
//     จึงเปลี่ยนชื่อตัวเลือก (แก้คำผิด) ได้โดยบาร์โค้ดไม่หลุดตำแหน่ง
//   - init ใช้กับสินค้าที่ยังไม่มีชั้นตัวเลือก; update ใช้กับสินค้าที่มีอยู่แล้ว
//   - ตัวเลือกที่ขายได้ = เอกสาร productbarcode (ไม่มีคอลเลกชันใหม่) บาร์โค้ดที่หลุดชุดผสมจะถูกปิดขาย ไม่ถูกลบ
//   - เขียน product และ productbarcode ใน transaction เดียว; ถ้าเซิร์ฟเวอร์ไม่รองรับ transaction จะเขียนแบบเรียงลำดับพร้อมบันทึกคำเตือน

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sort"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	commonmodels "smlcloudplatform/internal/models"
	productmodels "smlcloudplatform/internal/product/product/models"
	barcodemodels "smlcloudplatform/internal/product/productbarcode/models"
	"smlcloudplatform/internal/utils"
	msmodels "smlcloudplatform/pkg/microservice/models"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// productV2TierWritableTop — key ระดับบนสุดที่ tier/* รับ
var productV2TierWritableTop = map[string]struct{}{
	"itemcode":     {},
	"holdingcode":  {},
	"businesscode": {},
	"__v":          {},
	"tiers":        {},
	"models":       {},
	"itemunitcode": {},
}

// productV2TierBarcodeAttempts — จำนวนครั้งที่สุ่มบาร์โค้ดใหม่ก่อนยอมแพ้ (เท่ากับฝั่งหน้าจอเดิม)
const productV2TierBarcodeAttempts = 5

// productV2TierBarcodePrefix — คำนำหน้าบาร์โค้ดที่ระบบสร้างเอง (ช่วงรหัสภายในร้าน ไม่ใช่รหัสสากล)
const productV2TierBarcodePrefix = "200"

// productV2ErrVersionConflict — sentinel ภายในสำหรับยกเลิก transaction เมื่อ __v ไม่ตรง
var productV2ErrVersionConflict = errors.New("product v2: version conflict")

func ProductV2TierInitHandler(c echo.Context) error {
	return productV2TierHandler(c, false)
}

func ProductV2TierUpdateHandler(c echo.Context) error {
	return productV2TierHandler(c, true)
}

func productV2TierHandler(c echo.Context, isUpdate bool) error {
	logPrefix := "ProductV2TierInit"
	if isUpdate {
		logPrefix = "ProductV2TierUpdate"
	}

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

	// E24 — ฟิลด์ชั้นบัญชีต้องถูกปฏิเสธพร้อมรายชื่อ
	accounting, unknown := scanProductV2TierPayload(raw)
	if len(accounting) > 0 {
		message, fields := productV2AccountingFieldErrors(accounting)
		return productV2Error(c, http.StatusUnprocessableEntity, "ACCOUNTING_FIELD_READONLY", message, fields)
	}
	if len(unknown) > 0 {
		return productV2Error(c, http.StatusUnprocessableEntity, "VALIDATION_FAILED", productV2MsgValidationFailed, productV2UnknownFieldErrors(unknown))
	}

	var req productV2TierRequest
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

	fields := validateProductV2Tiers(req.Tiers, productV2DefaultLimits)
	if len(fields) == 0 {
		fields = validateProductV2TierModels(req.Tiers, req.Models)
	}
	if len(fields) > 0 {
		code := "VALIDATION_FAILED"
		message := productV2MsgValidationFailed
		if len(fields) == 1 && fields[0].Code == "TOO_MANY_COMBINATIONS" {
			code, message = "TIER_TOO_MANY_COMBINATIONS", fields[0].Message
		}
		return productV2Error(c, http.StatusUnprocessableEntity, code, message, fields)
	}

	ctx, cancel := context.WithTimeout(context.Background(), productV2Timeout)
	defer cancel()

	allowed, err := productV2CheckPermission(ctx, userInfo, productV2PermissionKey)
	if err != nil {
		logger.Error("%s: permission: %v", logPrefix, err)
		return productV2Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", productV2MsgPermissionLoad, nil)
	}
	if !allowed {
		return productV2Error(c, http.StatusForbidden, "FORBIDDEN", productV2MsgForbidden, nil)
	}
	// การสร้างบาร์โค้ดใหม่คือการสร้างเอกสารชั้นบัญชี จึงต้องมีสิทธิ์สินค้าเพิ่ม (§4.3)
	if productV2TierNeedsGenerate(req.Models) {
		allowedProduct, err := productV2CheckAnyPermission(ctx, userInfo, productV2ProductPermissionKeys)
		if err != nil {
			logger.Error("%s: product permission: %v", logPrefix, err)
			return productV2Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", productV2MsgPermissionLoad, nil)
		}
		if !allowedProduct {
			return productV2Error(c, http.StatusForbidden, "FORBIDDEN", productV2MsgForbiddenBarcode, nil)
		}
	}

	client, db := GetAtlasConnection()
	if db == nil {
		return productV2Error(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", productV2MsgDBUnavailable, nil)
	}
	products := db.Collection(productV2ProductsCollection)
	barcodes := db.Collection(productV2BarcodesCollection)
	itemFilter := productV2ItemFilter(holdingCode, businessCode, req.ItemCode)

	var current productV2ProductDoc
	if err := products.FindOne(ctx, itemFilter).Decode(&current); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return productV2NotFound(c)
		}
		logger.Error("%s: load product: %v", logPrefix, err)
		return productV2Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", productV2MsgInternal, nil)
	}

	hasTiers := current.Listing != nil && len(current.Listing.Tiers) > 0
	if !isUpdate && hasTiers {
		return productV2Error(c, http.StatusConflict, "TIER_ALREADY_EXISTS", productV2MsgTierExists, nil)
	}
	if isUpdate && !hasTiers {
		return productV2Error(c, http.StatusConflict, "TIER_NOT_INITIALIZED", productV2MsgTierMissing, nil)
	}

	// E17 — ห้ามแก้โครงตัวเลือกขณะรายการบนช่องทางติดโปรโมชัน
	locked, err := productV2HasPromotionLock(ctx, db, holdingCode, businessCode, req.ItemCode)
	if err != nil {
		logger.Error("%s: promotion lock: %v", logPrefix, err)
		return productV2Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", productV2MsgInternal, nil)
	}
	if locked {
		return productV2Error(c, http.StatusConflict, "LISTING_LOCKED_BY_PROMOTION", productV2MsgPromotionLock, nil)
	}

	// โหลดบาร์โค้ดทั้งหมดของสินค้านี้ครั้งเดียว ใช้ตรวจ E10 และหาบาร์โค้ดที่หลุดชุดผสม
	existing, err := loadProductV2ItemBarcodes(ctx, db, holdingCode, businessCode, req.ItemCode)
	if err != nil {
		logger.Error("%s: load barcodes: %v", logPrefix, err)
		return productV2Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", productV2MsgInternal, nil)
	}
	if fields := validateProductV2TierBarcodesExist(req.Models, existing); len(fields) > 0 {
		return productV2Error(c, http.StatusUnprocessableEntity, "VALIDATION_FAILED", productV2MsgValidationFailed, fields)
	}

	// หน่วยนับของบาร์โค้ดที่จะสร้างใหม่ ต้องเป็นหน่วยของสินค้านี้ (หน่วยหลักหรือหน่วยแปลง)
	unitCode := strings.TrimSpace(req.ItemUnitCode)
	var unitDef productmodels.ProductUnitConversion
	if productV2TierNeedsGenerate(req.Models) {
		if unitCode == "" {
			unitCode = current.UnitCode
		}
		def, ok := current.Product.UnitDefinition(unitCode)
		if !ok {
			return productV2Error(c, http.StatusUnprocessableEntity, "VALIDATION_FAILED", productV2MsgValidationFailed,
				[]productV2FieldError{{"itemunitcode", "INVALID_VALUE", "หน่วยนับนี้ไม่ใช่หน่วยของสินค้านี้ กรุณาเลือกหน่วยที่ตั้งไว้ในหน้าสินค้า"}})
		}
		unitDef = def
	}

	generated, err := generateProductV2Barcodes(ctx, db, holdingCode, businessCode, req.Models)
	if err != nil {
		if errors.Is(err, errProductV2BarcodeExhausted) {
			return productV2Error(c, http.StatusConflict, "BARCODE_GENERATE_FAILED", productV2MsgBarcodeGenFailed, nil)
		}
		logger.Error("%s: generate barcode: %v", logPrefix, err)
		return productV2Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", productV2MsgInternal, nil)
	}

	now := time.Now().UTC()
	tiers := buildProductV2TierDocs(req.Tiers)
	plan := buildProductV2TierWritePlan(req, tiers, generated, existing, current, unitCode, unitDef, userInfo.Username, now)

	productFilter := bson.M{}
	for k, v := range itemFilter {
		productFilter[k] = v
	}
	for k, v := range productV2VersionFilter(*req.Version) {
		productFilter[k] = v
	}
	productUpdate := bson.M{
		"$set": bson.M{
			"listing.tiers": tiers,
			"updatedat":     now,
			"updatedby":     userInfo.Username,
		},
		"$inc": bson.M{"__v": 1},
	}

	apply := func(ctx context.Context) error {
		result, err := products.UpdateOne(ctx, productFilter, productUpdate)
		if err != nil {
			return err
		}
		if result.MatchedCount == 0 {
			return productV2ErrVersionConflict
		}
		if len(plan) == 0 {
			return nil
		}
		_, err = barcodes.BulkWrite(ctx, plan, options.BulkWrite().SetOrdered(false))
		return err
	}

	if err := productV2RunAtomic(ctx, client, logPrefix, apply); err != nil {
		if errors.Is(err, productV2ErrVersionConflict) {
			count, cerr := products.CountDocuments(ctx, itemFilter, options.Count().SetLimit(1))
			if cerr != nil {
				logger.Error("%s: count: %v", logPrefix, cerr)
				return productV2Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", productV2MsgInternal, nil)
			}
			if count == 0 {
				return productV2NotFound(c)
			}
			return productV2Error(c, http.StatusConflict, "VERSION_CONFLICT", productV2MsgVersionConflict, nil)
		}
		if mongo.IsDuplicateKeyError(err) {
			return productV2Error(c, http.StatusConflict, "BARCODE_GENERATE_FAILED", productV2MsgBarcodeGenFailed, nil)
		}
		logger.Error("%s: apply: %v", logPrefix, err)
		return productV2Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", productV2MsgInternal, nil)
	}

	models := make([]map[string]any, 0, len(req.Models))
	for i, model := range req.Models {
		barcode := strings.TrimSpace(model.Barcode)
		created := false
		if generatedBarcode, ok := generated[i]; ok {
			barcode, created = generatedBarcode, true
		}
		models = append(models, map[string]any{
			"tierindex": model.TierIndex,
			"barcode":   barcode,
			"created":   created,
		})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"itemcode": req.ItemCode,
			"__v":      *req.Version + 1,
			"tiers":    tiers,
			"models":   models,
		},
	})
}

// scanProductV2TierPayload — เหมือน scanProductV2Payload แต่ใช้ชุด key ของ tier/*
func scanProductV2TierPayload(raw map[string]any) (accounting []string, unknown []string) {
	for key := range raw {
		if _, ok := productV2TierWritableTop[key]; ok {
			continue
		}
		if _, ok := productAccountingFields[key]; ok {
			accounting = append(accounting, key)
			continue
		}
		unknown = append(unknown, key)
	}
	sort.Strings(accounting)
	sort.Strings(unknown)
	return accounting, unknown
}

func productV2TierNeedsGenerate(models []productV2TierModelReq) bool {
	for _, model := range models {
		if strings.TrimSpace(model.Barcode) == "" && model.Generate {
			return true
		}
	}
	return false
}

// productV2ItemBarcode — ฟิลด์ของบาร์โค้ดที่ tier/* ต้องใช้
type productV2ItemBarcode struct {
	Barcode   string                               `bson:"barcode"`
	GuidFixed string                               `bson:"guidfixed"`
	Listing   *barcodemodels.ProductBarcodeListing `bson:"listing"`
}

func loadProductV2ItemBarcodes(ctx context.Context, db *mongo.Database, holdingCode, businessCode, itemCode string) (map[string]productV2ItemBarcode, error) {
	cursor, err := db.Collection(productV2BarcodesCollection).Find(ctx, bson.M{
		"holdingcode":  holdingCode,
		"businesscode": businessCode,
		"itemcode":     itemCode,
		"deletedat":    nil,
	}, options.Find().SetProjection(bson.M{"barcode": 1, "guidfixed": 1, "listing": 1}))
	if err != nil {
		return nil, err
	}
	var docs []productV2ItemBarcode
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	result := make(map[string]productV2ItemBarcode, len(docs))
	for _, doc := range docs {
		result[strings.ToLower(strings.TrimSpace(doc.Barcode))] = doc
	}
	return result, nil
}

// validateProductV2TierBarcodesExist — E10: บาร์โค้ดที่ระบุต้องมีอยู่จริงและเป็นของสินค้านี้
func validateProductV2TierBarcodesExist(models []productV2TierModelReq, existing map[string]productV2ItemBarcode) []productV2FieldError {
	errs := []productV2FieldError{}
	for i, model := range models {
		barcode := strings.TrimSpace(model.Barcode)
		if barcode == "" {
			continue
		}
		if _, ok := existing[strings.ToLower(barcode)]; !ok {
			errs = append(errs, productV2FieldError{
				Field:   fmt.Sprintf("models[%d].barcode", i),
				Code:    "NOT_FOUND",
				Message: "ไม่พบบาร์โค้ดนี้ในสินค้าตัวนี้ กรุณาเลือกบาร์โค้ดของสินค้านี้ หรือให้ระบบสร้างให้",
			})
		}
	}
	return errs
}

// productV2HasPromotionLock — มีรายการบนช่องทางที่เผยแพร่อยู่และติดโปรโมชันหรือไม่ (E17)
func productV2HasPromotionLock(ctx context.Context, db *mongo.Database, holdingCode, businessCode, itemCode string) (bool, error) {
	count, err := db.Collection(productV2ChannelListingsCollection).CountDocuments(ctx, bson.M{
		"holdingcode":           holdingCode,
		"businesscode":          businessCode,
		"itemcode":              itemCode,
		"isdeleted":             false,
		"sync.state":            "published",
		"platform.haspromotion": true,
	}, options.Count().SetLimit(1))
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func buildProductV2TierDocs(tiers []productV2TierReq) []productmodels.ProductListingTier {
	docs := make([]productmodels.ProductListingTier, 0, len(tiers))
	for i, tier := range tiers {
		options := make([]productmodels.ProductListingTierOption, 0, len(tier.Options))
		for j, option := range tier.Options {
			options = append(options, productmodels.ProductListingTierOption{
				XOrder:        j,
				Name:          strings.TrimSpace(option.Name),
				ImageURI:      strings.TrimSpace(option.ImageURI),
				ImageURIThumb: strings.TrimSpace(option.ImageURIThumb),
			})
		}
		docs = append(docs, productmodels.ProductListingTier{
			XOrder:  i,
			Name:    strings.TrimSpace(tier.Name),
			Options: options,
		})
	}
	return docs
}

// buildProductV2TierWritePlan — คำสั่งเขียน productbarcode ทั้งหมด: ผูก tierindex, สร้างบาร์โค้ดใหม่, ปิดขายตัวที่หลุดชุดผสม
func buildProductV2TierWritePlan(
	req productV2TierRequest,
	tiers []productmodels.ProductListingTier,
	generated map[int]string,
	existing map[string]productV2ItemBarcode,
	current productV2ProductDoc,
	unitCode string,
	unitDef productmodels.ProductUnitConversion,
	username string,
	now time.Time,
) []mongo.WriteModel {
	plan := make([]mongo.WriteModel, 0, len(req.Models)+len(existing))
	used := map[string]bool{}

	for i, model := range req.Models {
		barcode := strings.TrimSpace(model.Barcode)
		if newBarcode, ok := generated[i]; ok {
			doc := buildProductV2NewBarcodeDoc(req, tiers, model, newBarcode, current, unitCode, unitDef, username, now)
			plan = append(plan, mongo.NewInsertOneModel().SetDocument(doc))
			used[strings.ToLower(newBarcode)] = true
			continue
		}
		if barcode == "" {
			continue
		}
		used[strings.ToLower(barcode)] = true
		plan = append(plan, mongo.NewUpdateOneModel().
			SetFilter(bson.M{
				"holdingcode":  current.HoldingCode,
				"businesscode": current.BusinessCode,
				"barcode":      barcode,
			}).
			SetUpdate(bson.M{"$set": bson.M{
				"listing.tierindex": model.TierIndex,
				"listing.isforsale": true,
				"updatedat":         now,
				"updatedby":         username,
			}}))
	}

	// บาร์โค้ดเดิมที่เคยผูกตัวเลือกไว้แต่ไม่อยู่ในชุดใหม่ → ปิดขายออนไลน์ (ไม่ลบ เพราะมีประวัติสต๊อก)
	for key, doc := range existing {
		if used[key] || doc.Listing == nil || len(doc.Listing.TierIndex) == 0 {
			continue
		}
		plan = append(plan, mongo.NewUpdateOneModel().
			SetFilter(bson.M{
				"holdingcode":  current.HoldingCode,
				"businesscode": current.BusinessCode,
				"barcode":      doc.Barcode,
			}).
			SetUpdate(bson.M{"$set": bson.M{
				"listing.tierindex": []int{},
				"listing.isforsale": false,
				"updatedat":         now,
				"updatedby":         username,
			}}))
	}
	return plan
}

// buildProductV2NewBarcodeDoc — เอกสาร productbarcode ที่ระบบสร้างให้ชุดผสมนี้
// ชั้นบัญชีใส่เท่าที่จำเป็นต่อการขาย (รหัสสินค้า ชื่อ หน่วยนับ อัตราส่วน) ไม่คัดลอกราคา — ต้องตั้งราคาที่หน้าสินค้า
func buildProductV2NewBarcodeDoc(
	req productV2TierRequest,
	tiers []productmodels.ProductListingTier,
	model productV2TierModelReq,
	barcode string,
	current productV2ProductDoc,
	unitCode string,
	unitDef productmodels.ProductUnitConversion,
	username string,
	now time.Time,
) bson.M {
	tierIndex := model.TierIndex
	if tierIndex == nil {
		tierIndex = []int{}
	}
	return bson.M{
		"holdingcode":   current.HoldingCode,
		"businesscode":  current.BusinessCode,
		"guidfixed":     utils.NewID(),
		"itemcode":      req.ItemCode,
		"itemguid":      current.GuidFixed,
		"barcode":       barcode,
		"names":         productV2BarcodeNames(current, tiers, tierIndex),
		"itemunitcode":  unitCode,
		"itemunitnames": unitDef.UnitNames,
		"standvalue":    productV2UnitValue(unitDef.StandValue),
		"dividevalue":   productV2UnitValue(unitDef.DivideValue),
		"condition":     false,
		"ismainbarcode": false,
		"listing": bson.M{
			"tierindex": tierIndex,
			"gtin":      "",
			"isforsale": true,
		},
		"isdeleted": false,
		"deletedat": nil,
		"createdat": now,
		"createdby": username,
		"__v":       int64(0),
	}
}

func productV2UnitValue(value int64) int64 {
	if value <= 0 {
		return 1
	}
	return value
}

// productV2BarcodeNames — ชื่อบาร์โค้ดใหม่ = ชื่อสินค้าแต่ละภาษา + ชื่อตัวเลือก
func productV2BarcodeNames(current productV2ProductDoc, tiers []productmodels.ProductListingTier, tierIndex []int) []commonmodels.NameX {
	label := productV2TierLabelFromDocs(tiers, tierIndex)
	names := []commonmodels.NameX{}
	for _, name := range productV2Names(current.Names) {
		if name.Code == nil || name.Name == nil {
			continue
		}
		combined := strings.TrimSpace(*name.Name)
		if label != "" {
			combined = strings.TrimSpace(combined + " " + label)
		}
		code := *name.Code
		names = append(names, commonmodels.NameX{Code: &code, Name: &combined})
	}
	if len(names) == 0 {
		code, value := "th", strings.TrimSpace(current.Code+" "+label)
		names = append(names, commonmodels.NameX{Code: &code, Name: &value})
	}
	return names
}

func productV2TierLabelFromDocs(tiers []productmodels.ProductListingTier, tierIndex []int) string {
	parts := make([]string, 0, len(tierIndex))
	for level, idx := range tierIndex {
		if level >= len(tiers) || idx < 0 || idx >= len(tiers[level].Options) {
			continue
		}
		parts = append(parts, tiers[level].Options[idx].Name)
	}
	return strings.Join(parts, " / ")
}

var errProductV2BarcodeExhausted = errors.New("product v2: barcode candidates exhausted")

// generateProductV2Barcodes — สร้างเลขบาร์โค้ดให้ชุดผสมที่ขอ generate โดยตรวจไม่ให้ซ้ำของเดิมในบริษัท
// คืน map[ตำแหน่งใน models]เลขบาร์โค้ด
func generateProductV2Barcodes(ctx context.Context, db *mongo.Database, holdingCode, businessCode string, models []productV2TierModelReq) (map[int]string, error) {
	result := map[int]string{}
	taken := map[string]bool{}
	for i, model := range models {
		if strings.TrimSpace(model.Barcode) != "" || !model.Generate {
			continue
		}
		barcode := ""
		for attempt := 0; attempt < productV2TierBarcodeAttempts; attempt++ {
			candidate, err := productV2RandomBarcode()
			if err != nil {
				return nil, err
			}
			if taken[candidate] {
				continue
			}
			count, err := db.Collection(productV2BarcodesCollection).CountDocuments(ctx, bson.M{
				"holdingcode":  holdingCode,
				"businesscode": businessCode,
				"barcode":      candidate,
			}, options.Count().SetLimit(1))
			if err != nil {
				return nil, err
			}
			if count == 0 {
				barcode = candidate
				break
			}
		}
		if barcode == "" {
			return nil, errProductV2BarcodeExhausted
		}
		taken[barcode] = true
		result[i] = barcode
	}
	return result, nil
}

// productV2RandomBarcode — เลข 13 หลัก: คำนำหน้าภายในร้าน + สุ่ม 9 หลัก + หลักตรวจสอบ
func productV2RandomBarcode() (string, error) {
	digits := make([]byte, 0, 13)
	digits = append(digits, []byte(productV2TierBarcodePrefix)...)
	for i := 0; i < 12-len(productV2TierBarcodePrefix); i++ {
		n, err := crand.Int(crand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		digits = append(digits, byte('0'+n.Int64()))
	}
	return string(digits) + productV2CheckDigit(string(digits)), nil
}

// productV2CheckDigit — หลักตรวจสอบมาตรฐาน (น้ำหนัก 1 และ 3 สลับกันจากซ้าย)
func productV2CheckDigit(body string) string {
	sum := 0
	for i, ch := range body {
		digit := int(ch - '0')
		if i%2 == 0 {
			sum += digit
		} else {
			sum += digit * 3
		}
	}
	return fmt.Sprintf("%d", (10-sum%10)%10)
}

// productV2RunAtomic — รัน apply ใน transaction; ถ้าเซิร์ฟเวอร์ไม่รองรับ (ไม่ใช่ replica set) จะรันตรงพร้อมบันทึกคำเตือน
func productV2RunAtomic(ctx context.Context, client *mongo.Client, logPrefix string, apply func(context.Context) error) error {
	if client == nil {
		return apply(ctx)
	}
	session, err := client.StartSession()
	if err != nil {
		logger.Warn("%s: start session ไม่สำเร็จ เขียนแบบไม่ใช้ transaction: %v", logPrefix, err)
		return apply(ctx)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (any, error) {
		return nil, apply(sc)
	})
	if err != nil && productV2IsTransactionUnsupported(err) {
		logger.Warn("%s: เซิร์ฟเวอร์ไม่รองรับ transaction เขียนแบบเรียงลำดับแทน: %v", logPrefix, err)
		return apply(ctx)
	}
	return err
}

// productV2IsTransactionUnsupported — MongoDB แบบ standalone จะปฏิเสธ transaction ด้วยข้อความนี้
func productV2IsTransactionUnsupported(err error) bool {
	if errors.Is(err, productV2ErrVersionConflict) {
		return false
	}
	message := err.Error()
	return strings.Contains(message, "Transaction numbers are only allowed") ||
		strings.Contains(message, "Transactions are not supported") ||
		strings.Contains(message, "IllegalOperation")
}
