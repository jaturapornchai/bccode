package handlers

// POST /goapi/product/v2/item/readiness — ตรวจว่าสินค้าพร้อมลงขายหรือยัง แยกเป็น 3 กลุ่ม (§5.3)
//   accounting — ชั้นบัญชี (อ่านอย่างเดียว): เปิดใช้งาน มีชื่อ มีหน่วยนับ
//   listing    — ชั้นลงขาย: กฎกลางเดียวกับ readiness ใน update-listing + ความยาวชื่อ (E01) + ครอบคลุม model (E09)
//   channel    — เฉพาะเมื่อส่ง channel+shopid คู่กัน: ร้าน/การเชื่อมต่อ/การจับคู่หมวด+แบรนด์ (อ่านจากข้อมูลที่พักไว้ ไม่ต่อช่องทางจริง)
//
// หมายเหตุดีไซน์: กลุ่ม channel ถูก omit จาก response เมื่อไม่ได้ส่ง channel/shopid (ไม่ตอบ ready แทนการไม่ได้ตรวจ)
// การตรวจ listing เมื่อสินค้ามี tiers เช็คครบทุกชุดผสมตาม E09 (ไม่ใช่แค่มี model ตัวเดียว)

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/models"
	productmodels "smlcloudplatform/internal/product/product/models"
	msmodels "smlcloudplatform/pkg/microservice/models"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// productV2ChannelShopDoc — เฉพาะฟิลด์ของ channel_shop ที่ readiness ใช้ (ไม่แตะ auth.credentialref)
type productV2ChannelShopDoc struct {
	IsActive bool `bson:"isactive"`
	Auth     *struct {
		TokenExpiresAt time.Time `bson:"tokenexpiresat"`
	} `bson:"auth"`
}

// productV2ChannelCategoryMapDoc — เฉพาะฟิลด์ของ channel_category_map ที่ readiness ใช้
type productV2ChannelCategoryMapDoc struct {
	CategoryID     int64 `bson:"categoryid"`
	BrandMandatory bool  `bson:"brandmandatory"`
}

func ProductV2ItemReadinessHandler(c echo.Context) error {
	productV2LimitBody(c)
	var req productV2ItemReadinessRequest
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
	req.Channel = strings.TrimSpace(req.Channel)
	req.ShopID = strings.TrimSpace(req.ShopID)
	if (req.Channel == "") != (req.ShopID == "") {
		return productV2Error(c, http.StatusUnprocessableEntity, "VALIDATION_FAILED", productV2MsgValidationFailed, []productV2FieldError{{"channel", "REQUIRED", "กรุณาระบุช่องทาง (channel) และรหัสร้าน (shopid) พร้อมกัน หรือเว้นว่างทั้งคู่"}})
	}

	ctx, cancel := context.WithTimeout(context.Background(), productV2Timeout)
	defer cancel()

	allowed, err := productV2CheckPermission(ctx, userInfo, productV2PermissionKey)
	if err != nil {
		logger.Error("ProductV2ItemReadiness: permission: %v", err)
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
		logger.Error("ProductV2ItemReadiness: load product: %v", err)
		return productV2Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", productV2MsgInternal, nil)
	}

	cursor, err := db.Collection(productV2BarcodesCollection).Find(ctx, bson.M{
		"holdingcode":  holdingCode,
		"businesscode": businessCode,
		"itemcode":     req.ItemCode,
		"deletedat":    nil,
	}, options.Find().SetProjection(bson.M{"barcode": 1, "listing": 1}).SetLimit(500))
	if err != nil {
		logger.Error("ProductV2ItemReadiness: load models: %v", err)
		return productV2Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", productV2MsgInternal, nil)
	}
	var modelDocs []productV2ModelDoc
	if err := cursor.All(ctx, &modelDocs); err != nil {
		logger.Error("ProductV2ItemReadiness: decode models: %v", err)
		return productV2Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", productV2MsgInternal, nil)
	}

	data := map[string]any{
		"accounting": productV2AccountingReadiness(doc),
		"listing":    productV2ListingReadinessFull(doc.ProductDoc, modelDocs),
	}

	if req.Channel != "" {
		channel, err := productV2ChannelReadinessGroup(ctx, db, holdingCode, businessCode, req.Channel, req.ShopID, doc.ProductDoc.CategoryCode, doc.ProductDoc.BrandCode)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return productV2Error(c, http.StatusNotFound, "SHOP_NOT_FOUND", productV2MsgShopNotFound, nil)
			}
			logger.Error("ProductV2ItemReadiness: channel: %v", err)
			return productV2Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", productV2MsgInternal, nil)
		}
		data["channel"] = channel
	}

	return c.JSON(http.StatusOK, map[string]any{"success": true, "data": data})
}

// productV2ReadinessGroup — ประกอบ {ready, missing[]} รูปแบบเดียวกันทั้ง 3 กลุ่ม (missing ไม่เป็น nil เพื่อให้ JSON เป็น [] เสมอ)
func productV2ReadinessGroup(missing []map[string]string) map[string]any {
	if missing == nil {
		missing = []map[string]string{}
	}
	return map[string]any{"ready": len(missing) == 0, "missing": missing}
}

// productV2ListingReadiness — ความพร้อมตามกฎกลาง (ชื่อลงขาย, กล่องครบ, มีรูป ≥ 1)
// ใช้ร่วมกันระหว่าง update-listing (ตอบกลับหลังบันทึก) และ item/readiness — ความพร้อมต่อช่องทางอยู่ที่ productV2ChannelReadinessGroup
func productV2ListingReadiness(doc productmodels.ProductDoc) map[string]any {
	missing := []map[string]string{}
	if doc.Listing == nil || strings.TrimSpace(doc.Listing.Title) == "" {
		missing = append(missing, map[string]string{"field": "listing.title", "message": "ยังไม่ได้กรอกชื่อสำหรับลงขาย"})
	}
	if doc.PackageWeight <= 0 {
		missing = append(missing, map[string]string{"field": "package.weight", "message": "ยังไม่ได้กรอกน้ำหนักสินค้า (กิโลกรัม)"})
	}
	if doc.PackageLength <= 0 || doc.PackageWidth <= 0 || doc.PackageHeight <= 0 {
		missing = append(missing, map[string]string{"field": "package", "message": "ยังไม่ได้กรอกขนาดกล่อง (เซนติเมตร) ให้ครบ"})
	}
	if doc.Images == nil || len(*doc.Images) == 0 {
		missing = append(missing, map[string]string{"field": "media.images", "message": "ต้องมีรูปสินค้าอย่างน้อย 1 รูป"})
	}
	return map[string]any{"ready": len(missing) == 0, "missing": missing}
}

// productV2ListingReadinessFull — กลุ่ม listing ของ item/readiness: กฎกลาง + ความยาวชื่อ (E01) + ครอบคลุม model (E09)
func productV2ListingReadinessFull(doc productmodels.ProductDoc, modelDocs []productV2ModelDoc) map[string]any {
	missing := productV2ListingReadiness(doc)["missing"].([]map[string]string)
	if doc.Listing != nil && strings.TrimSpace(doc.Listing.Title) != "" {
		for _, fieldErr := range validateListingTitle(doc.Listing.Title, productV2DefaultLimits) {
			missing = append(missing, map[string]string{"field": fieldErr.Field, "message": fieldErr.Message})
		}
	}
	tiers := []productmodels.ProductListingTier{}
	if doc.Listing != nil && doc.Listing.Tiers != nil {
		tiers = doc.Listing.Tiers
	}
	missing = append(missing, productV2ModelsReadiness(tiers, modelDocs)...)
	return productV2ReadinessGroup(missing)
}

// productV2AccountingReadiness — ชั้นบัญชีต้องออกเอกสารได้: เปิดใช้งาน มีชื่อ มีหน่วยนับ (ชั้นนี้แก้ผ่าน v2 ไม่ได้ แจ้งให้ไปแก้ที่หน้าสินค้า)
func productV2AccountingReadiness(doc productV2ProductDoc) map[string]any {
	missing := []map[string]string{}
	if !productV2IsActive(doc) {
		missing = append(missing, map[string]string{"field": "isactive", "message": "สินค้านี้ถูกปิดใช้งานอยู่ กรุณาเปิดใช้ในหน้าสินค้าก่อน"})
	}
	if !productV2HasName(doc.ProductDoc.Names) {
		missing = append(missing, map[string]string{"field": "names", "message": "สินค้ายังไม่มีชื่อ กรุณาตั้งชื่อในหน้าสินค้า"})
	}
	if strings.TrimSpace(doc.ProductDoc.UnitCode) == "" {
		missing = append(missing, map[string]string{"field": "unitcode", "message": "สินค้ายังไม่มีหน่วยนับ กรุณาตั้งหน่วยนับในหน้าสินค้า"})
	}
	return productV2ReadinessGroup(missing)
}

// productV2HasName — มีชื่อที่ไม่ใช่ช่องว่างอย่างน้อย 1 ภาษา
func productV2HasName(names *[]models.NameX) bool {
	if names == nil {
		return false
	}
	for _, name := range *names {
		if name.Name != nil && strings.TrimSpace(*name.Name) != "" {
			return true
		}
	}
	return false
}

// productV2ModelsReadiness — E09: ต้องมี model ที่เปิดขายอย่างน้อย 1; ถ้ามี tiers ทุกชุดผสมต้องมี model ที่ tierindex ตรง
func productV2ModelsReadiness(tiers []productmodels.ProductListingTier, modelDocs []productV2ModelDoc) []map[string]string {
	if len(modelDocs) == 0 {
		return []map[string]string{{"field": "models", "message": "สินค้านี้ยังไม่มีบาร์โค้ด กรุณาเพิ่มบาร์โค้ดในหน้าสินค้าก่อน"}}
	}
	missing := []map[string]string{}
	hasForSale := false
	covered := map[string]bool{}
	for _, m := range modelDocs {
		if m.Listing == nil {
			continue
		}
		if m.Listing.IsForSale {
			hasForSale = true
		}
		if productV2TierIndexInRange(m.Listing.TierIndex, tiers) {
			covered[productV2TierIndexKey(m.Listing.TierIndex)] = true
		}
	}
	if !hasForSale {
		missing = append(missing, map[string]string{"field": "models", "message": "ยังไม่มีตัวเลือกที่เปิดขาย กรุณาเปิดขายอย่างน้อย 1 ตัวเลือก"})
	}
	if len(tiers) > 0 {
		expected := 1
		for _, tier := range tiers {
			expected *= len(tier.Options)
		}
		if short := expected - len(covered); short > 0 {
			missing = append(missing, map[string]string{"field": "models", "message": fmt.Sprintf("ตัวเลือกยังไม่ครบ ขาดอีก %d ชุด กรุณาจัดการตัวเลือกให้ครบทุกชุดผสม", short)})
		}
	}
	return missing
}

// productV2TierIndexInRange — tierindex ยาวเท่าจำนวนชั้นและทุกค่าอยู่ในช่วง options ของชั้นนั้น
func productV2TierIndexInRange(tierIndex []int, tiers []productmodels.ProductListingTier) bool {
	if len(tierIndex) != len(tiers) {
		return false
	}
	for i, idx := range tierIndex {
		if idx < 0 || idx >= len(tiers[i].Options) {
			return false
		}
	}
	return true
}

func productV2TierIndexKey(tierIndex []int) string {
	parts := make([]string, len(tierIndex))
	for i, idx := range tierIndex {
		parts[i] = fmt.Sprintf("%d", idx)
	}
	return strings.Join(parts, ",")
}

// productV2ChannelReadinessGroup — กลุ่ม channel: โหลดร้าน/หมวด/แบรนด์ที่จับคู่จาก Mongo แล้วประเมินด้วย productV2ChannelReadiness
// ร้านไม่เจอ → คืน mongo.ErrNoDocuments ให้ handler ตอบ 404 SHOP_NOT_FOUND
func productV2ChannelReadinessGroup(ctx context.Context, db *mongo.Database, holdingCode, businessCode, channel, shopID, categoryCode, brandCode string) (map[string]any, error) {
	var shop productV2ChannelShopDoc
	err := db.Collection(productV2ChannelShopsCollection).FindOne(ctx, bson.M{
		"holdingcode":  holdingCode,
		"businesscode": businessCode,
		"channel":      channel,
		"shopid":       shopID,
		"isdeleted":    false,
	}, options.FindOne().SetProjection(bson.M{"isactive": 1, "auth.tokenexpiresat": 1})).Decode(&shop)
	if err != nil {
		return nil, err
	}

	var categoryMap *productV2ChannelCategoryMapDoc
	if strings.TrimSpace(categoryCode) != "" {
		var found productV2ChannelCategoryMapDoc
		err = db.Collection(productV2ChannelCategoryMapsCollection).FindOne(ctx, bson.M{
			"holdingcode":  holdingCode,
			"businesscode": businessCode,
			"channel":      channel,
			"categorycode": categoryCode,
			"isdeleted":    false,
		}, options.FindOne().SetProjection(bson.M{"categoryid": 1, "brandmandatory": 1})).Decode(&found)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		if err == nil {
			categoryMap = &found
		}
	}

	hasBrandMap := false
	if categoryMap != nil && categoryMap.BrandMandatory {
		count, err := db.Collection(productV2ChannelBrandMapsCollection).CountDocuments(ctx, bson.M{
			"holdingcode":  holdingCode,
			"businesscode": businessCode,
			"channel":      channel,
			"brandcode":    brandCode,
			"isdeleted":    false,
		}, options.Count().SetLimit(1))
		if err != nil {
			return nil, err
		}
		hasBrandMap = count > 0
	}

	return productV2ChannelReadiness(shop, categoryCode, categoryMap, hasBrandMap, time.Now().UTC()), nil
}

// productV2ChannelReadiness — ตรรกะล้วน (ทดสอบได้): ร้านเปิดใช้งาน การเชื่อมต่อไม่หมดอายุ จับคู่หมวดแล้ว หมวดที่บังคับแบรนด์ต้องจับคู่แบรนด์แล้ว
func productV2ChannelReadiness(shop productV2ChannelShopDoc, categoryCode string, categoryMap *productV2ChannelCategoryMapDoc, hasBrandMap bool, now time.Time) map[string]any {
	missing := []map[string]string{}
	if !shop.IsActive {
		missing = append(missing, map[string]string{"field": "shop.isactive", "message": "ร้านนี้ถูกปิดการเชื่อมต่ออยู่ กรุณาเปิดใช้ในหน้าตั้งค่าร้าน"})
	}
	if shop.Auth != nil && !shop.Auth.TokenExpiresAt.IsZero() && !now.Before(shop.Auth.TokenExpiresAt) {
		missing = append(missing, map[string]string{"field": "shop.auth", "message": "สิทธิ์เชื่อมต่อร้านหมดอายุแล้ว กรุณาเชื่อมร้านใหม่"})
	}
	if strings.TrimSpace(categoryCode) == "" {
		missing = append(missing, map[string]string{"field": "categoryid", "message": "สินค้านี้ยังไม่มีหมวดหมู่ กรุณาตั้งหมวดหมู่ในหน้าสินค้าก่อน"})
	} else if categoryMap == nil {
		missing = append(missing, map[string]string{"field": "categoryid", "message": fmt.Sprintf("ยังไม่ได้เลือกหมวดหมู่ของช่องทางสำหรับหมวด %s", categoryCode)})
	} else if categoryMap.BrandMandatory && !hasBrandMap {
		missing = append(missing, map[string]string{"field": "brandid", "message": "หมวดนี้บังคับระบุแบรนด์ กรุณาจับคู่แบรนด์ของช่องทางก่อน"})
	}
	return productV2ReadinessGroup(missing)
}
