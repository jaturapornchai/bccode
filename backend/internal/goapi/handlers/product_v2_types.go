package handlers

// ชั้นลงขาย (Product Listing API v2) — สัญญา: docs/kms/architecture/product-listing-api-v2.md
// ไฟล์นี้รวม DTO / ค่าคงที่ / ข้อความภาษาไทย ที่ handler product_v2_*.go ใช้ร่วมกัน

import (
	"errors"
	"net/http"
	"time"

	productmodels "smlcloudplatform/internal/product/product/models"

	"github.com/labstack/echo/v4"
)

const (
	productV2ProductsCollection = "products"
	productV2BarcodesCollection = "productbarcodes"
	// คอลเลกชันชั้นช่องทาง (ยังไม่มีเอกสารจริงในเฟส 1 — shop/* อยู่เฟส 2) ชื่อตาม label ใน MongoModel + plural ตาม products/productbarcodes
	productV2ChannelShopsCollection        = "channel_shops"
	productV2ChannelCategoryMapsCollection = "channel_category_maps"
	productV2ChannelBrandMapsCollection    = "channel_brand_maps"
	productV2ChannelListingsCollection     = "channel_listings"
	productV2Timeout                       = 15 * time.Second
	productV2MaxBodyBytes                  = 1 << 20 // 1 MiB — กัน body ใหญ่ผิดปกติกินหน่วยความจำ

	// productV2PermissionKey — สิทธิ์แก้ชั้นลงขาย (§2, รอยืนยัน G9)
	// หมายเหตุ: role_permission ที่ใช้งานจริงตอนนี้เก็บ menu id (เช่น "product") ยังไม่มีใครกำหนด key นี้ให้
	// ดังนั้นในเฟสนี้ผู้ที่ผ่านคือ ADMIN/OWNER (wildcard "*") และผู้ดูแลระบบ; USER จะได้ 403 จนกว่า seed/UI จะกำหนดสิทธิ์นี้
	productV2PermissionKey = "/product/listing"
)

// ข้อความภาษาไทยตาม §7
const (
	productV2MsgUnauthorized      = "กรุณาเข้าสู่ระบบใหม่"
	productV2MsgForbidden         = "คุณไม่มีสิทธิ์แก้ข้อมูลส่วนนี้ ติดต่อผู้ดูแลเพื่อขอสิทธิ์ \"ลงขายออนไลน์\""
	productV2MsgItemNotFound      = "ไม่พบสินค้ารหัสนี้ในบริษัทที่เลือก"
	productV2MsgVersionConflict   = "มีคนแก้ข้อมูลนี้ก่อนหน้าคุณ กรุณาโหลดหน้าใหม่แล้วแก้อีกครั้ง"
	productV2MsgValidationFailed  = "ข้อมูลยังไม่ครบ ดูรายการที่ต้องแก้ด้านล่าง"
	productV2MsgPackageIncomplete = "กรอกน้ำหนักและขนาดกล่องให้ครบทั้ง 4 ช่อง หรือเว้นว่างทั้งหมด"
	productV2MsgAccountingPrefix  = "ฟิลด์ต่อไปนี้แก้ได้เฉพาะหน้าสินค้าโดยผู้มีสิทธิ์บัญชี: "
	productV2MsgInvalidJSON       = "รูปแบบข้อมูลไม่ถูกต้อง"
	productV2MsgDBUnavailable     = "ระบบฐานข้อมูลยังไม่พร้อม กรุณาลองใหม่"
	productV2MsgPermissionLoad    = "โหลดสิทธิ์ของผู้ใช้งานไม่สำเร็จ"
	productV2MsgInternal          = "บันทึกข้อมูลไม่สำเร็จ กรุณาลองใหม่"
	productV2MsgBodyTooLarge      = "ข้อมูลที่ส่งมามีขนาดใหญ่เกินไป (ไม่เกิน 1 MB)"
	productV2MsgTiersReadOnly     = "แก้ชั้นตัวเลือกได้ผ่านเมนูตัวเลือกสินค้าเท่านั้น"
	productV2MsgPackageViaObject  = "กรุณาส่งน้ำหนัก/ขนาดกล่องผ่าน package{} (weight, length, width, height)"
	productV2MsgShopNotFound      = "ไม่พบร้านค้าที่เชื่อมไว้ กรุณาเชื่อมร้านก่อน"

	// tier/* (§4.2)
	productV2MsgForbiddenBarcode = "การสร้างบาร์โค้ดใหม่ต้องมีสิทธิ์จัดการสินค้า ติดต่อผู้ดูแลเพื่อขอสิทธิ์"
	productV2MsgTierExists       = "สินค้านี้มีชั้นตัวเลือกอยู่แล้ว กรุณาใช้เมนูแก้ไขตัวเลือก"
	productV2MsgTierMissing      = "สินค้านี้ยังไม่มีชั้นตัวเลือก กรุณาสร้างตัวเลือกก่อน"
	productV2MsgPromotionLock    = "สินค้านี้กำลังติดโปรโมชันบนช่องทางขายออนไลน์ แก้ตัวเลือกได้หลังโปรโมชันจบ"
	productV2MsgBarcodeGenFailed = "ระบบสร้างเลขบาร์โค้ดใหม่ไม่สำเร็จ (เลขซ้ำ) กรุณากดบันทึกอีกครั้ง"
)

// productV2ProductPermissionKeys — สิทธิ์จัดการสินค้า/บาร์โค้ด (ชั้นบัญชี)
// มีสองรูปแบบเพราะข้อมูลจริงเก็บเป็นรหัสเมนู ("product") ส่วนสัญญา §2 เขียนเป็น path ("/product") — รอยืนยัน G9
var productV2ProductPermissionKeys = []string{"/product", "product"}

// productV2FieldError — รายการฟิลด์ที่ผิดพลาด (fields[] ใน response §1)
type productV2FieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// productV2Limits — ขีดจำกัดที่ validator ใช้เมื่อยังไม่มี limits ของร้าน (§10: ค่าประมาณของเราเอง)
type productV2Limits struct {
	TitleMin              int
	TitleMax              int
	ListingDescriptionMin int
	ListingDescriptionMax int
	ProductDescriptionMax int // product.description ยังคง max 1500 ตาม product.go
	ImageCountMax         int
	VideoCountMax         int
	WholesaleCountMax     int
	URIMax                int     // ความยาว uri ทุกชนิด (รูป/วิดีโอ/ตารางไซซ์)
	PackageWeightMax      float64 // กิโลกรัม
	PackageDimensionMax   float64 // เซนติเมตร
	TierMax               int     // จำนวนชั้นตัวเลือก (E03)
	TierCombinationMax    int     // จำนวนชุดผสมของทุกชั้น (E04)
	TierNameMax           int     // ความยาวชื่อชั้นตัวเลือก
	TierOptionNameMax     int     // ความยาวชื่อตัวเลือก
}

// productV2DefaultLimits — ค่าเริ่มต้นเมื่อยังไม่มี limits ของร้าน
// ตามสัญญา §6: E01 5..120, E14 ≤9; ค่าที่สัญญาไม่ได้กำหนด (E02 0..2000, วิดีโอ/ขายส่ง/uri/กล่อง) เป็นค่าประมาณของเราเอง (§10)
// ยังไม่บังคับ: wholesalepctmin (E07) และช่วง daystoship (E08) — ตรวจเมื่อมี limits ของหมวด/ร้านแล้ว
var productV2DefaultLimits = productV2Limits{
	TitleMin:              5,
	TitleMax:              120,
	ListingDescriptionMin: 0,
	ListingDescriptionMax: 2000,
	ProductDescriptionMax: 1500,
	ImageCountMax:         9,
	VideoCountMax:         1,
	WholesaleCountMax:     10,
	URIMax:                2048,
	PackageWeightMax:      1000,
	PackageDimensionMax:   1000,
	TierMax:               2,
	TierCombinationMax:    50,
	TierNameMax:           20,
	TierOptionNameMax:     40,
}

type productV2ItemGetRequest struct {
	ItemCode     string `json:"itemcode"`
	HoldingCode  string `json:"holdingcode"`
	BusinessCode string `json:"businesscode"`
}

// productV2ItemReadinessRequest — body ของ item/readiness (§5.3); channel/shopid ต้องส่งคู่กันหรือไม่ส่งเลย
type productV2ItemReadinessRequest struct {
	ItemCode     string `json:"itemcode"`
	HoldingCode  string `json:"holdingcode"`
	BusinessCode string `json:"businesscode"`
	Channel      string `json:"channel"`
	ShopID       string `json:"shopid"`
}

// productV2PackagePatch — package{} ใน payload; ส่งแค่บางช่อง = ไม่ครบ (E05)
type productV2PackagePatch struct {
	Weight *float64 `json:"weight"`
	Length *float64 `json:"length"`
	Width  *float64 `json:"width"`
	Height *float64 `json:"height"`
}

// productV2ListingPatch — listing{} ใน payload; ฟิลด์ที่ไม่ส่งมาจะคงค่าเดิม
// ไม่มี tiers — ชั้นตัวเลือกแก้ผ่าน tier/* เท่านั้น (§4.2, E03/E04 + sync tierindex); ส่งมาจะถูกปฏิเสธ 422
type productV2ListingPatch struct {
	Title         *string                                    `json:"title"`
	Description   *string                                    `json:"description"`
	Condition     *string                                    `json:"condition"`
	Preorder      *productmodels.ProductListingPreorder      `json:"preorder"`
	PurchaseLimit *productmodels.ProductListingPurchaseLimit `json:"purchaselimit"`
	Wholesale     *[]productmodels.ProductListingWholesale   `json:"wholesale"`
	SizeChart     *productmodels.ProductListingSizeChart     `json:"sizechart"`
}

// productV2UpdateListingRequest — body ของ item/update-listing (§5.2) หลังผ่านการกรองฟิลด์บัญชีแล้ว
type productV2UpdateListingRequest struct {
	ItemCode      string                        `json:"itemcode"`
	HoldingCode   string                        `json:"holdingcode"`
	BusinessCode  string                        `json:"businesscode"`
	Version       *int64                        `json:"__v"`
	Listing       *productV2ListingPatch        `json:"listing"`
	Package       *productV2PackagePatch        `json:"package"`
	Description   *string                       `json:"description"`
	ImageURI      *string                       `json:"imageuri"`
	ImageURIThumb *string                       `json:"imageurithumb"`
	Images        *[]productmodels.ProductImage `json:"images"`
	Videos        *[]productmodels.ProductVideo `json:"videos"`
}

// productV2ProductDoc — เอกสาร product ฝั่งอ่าน + เลขเวอร์ชัน (__v ไม่ถูกเพิ่มใน ProductDoc เพื่อไม่กระทบการเขียนฝั่งบัญชี)
type productV2ProductDoc struct {
	productmodels.ProductDoc `bson:",inline"`
	Version                  int64 `bson:"__v"`
	IsActive                 *bool `bson:"isactive"` // ชั้นบัญชี (Product struct ไม่มีฟิลด์นี้) อ่านอย่างเดียว
}

// productV2LimitBody — จำกัดขนาด body ก่อนอ่าน; error ที่ได้ตรวจด้วย productV2IsBodyTooLarge
func productV2LimitBody(c echo.Context) {
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, productV2MaxBodyBytes)
}

func productV2IsBodyTooLarge(err error) bool {
	var maxBytesErr *http.MaxBytesError
	return errors.As(err, &maxBytesErr)
}

// productV2Error — ตอบ error ตามรูปแบบ §1
func productV2Error(c echo.Context, status int, code, message string, fields []productV2FieldError) error {
	body := map[string]any{
		"success": false,
		"code":    code,
		"message": message,
	}
	if len(fields) > 0 {
		body["fields"] = fields
	}
	return c.JSON(status, body)
}

func productV2CompanyError(c echo.Context, cerr *companyContextError) error {
	message := cerr.Message
	switch cerr.Code {
	case "UNAUTHORIZED":
		message = productV2MsgUnauthorized
	case "FORBIDDEN":
		message = productV2MsgForbidden
	}
	return productV2Error(c, cerr.Status, cerr.Code, message, nil)
}

func productV2NotFound(c echo.Context) error {
	return productV2Error(c, http.StatusNotFound, "ITEM_NOT_FOUND", productV2MsgItemNotFound, nil)
}
