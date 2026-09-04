package handlers

// test ตรรกะล้วนของ item/readiness (§5.3) — ไม่แตะ Mongo
// ครอบคลุม: กลุ่ม accounting / listing (กฎกลาง + E01 + E09) / channel (ร้าน หมวด แบรนด์)

import (
	"strings"
	"testing"
	"time"

	"smlcloudplatform/internal/models"
	productmodels "smlcloudplatform/internal/product/product/models"
	barcodemodels "smlcloudplatform/internal/product/productbarcode/models"
)

func readinessMissing(group map[string]any) []map[string]string {
	missing, _ := group["missing"].([]map[string]string)
	return missing
}

func readinessFields(group map[string]any) string {
	parts := []string{}
	for _, m := range readinessMissing(group) {
		parts = append(parts, m["field"])
	}
	return strings.Join(parts, ",")
}

func readyProductDoc() productmodels.ProductDoc {
	doc := productmodels.ProductDoc{}
	doc.Code = "P1-004"
	doc.UnitCode = "PCS"
	th := "th"
	name := "สินค้าทดสอบ"
	doc.Names = &[]models.NameX{{Code: &th, Name: &name}}
	doc.PackageWeight, doc.PackageLength, doc.PackageWidth, doc.PackageHeight = 0.5, 10, 10, 10
	doc.Images = &[]productmodels.ProductImage{{XOrder: 0, URI: "/goapi/s3/file/a.webp", URIThumb: "/goapi/s3/file/a_thumb.webp"}}
	doc.Listing = &productmodels.ProductListing{Title: "ชื่อสำหรับลงขายที่ยาวพอ"}
	return doc
}

func TestProductV2ListingReadinessCore(t *testing.T) {
	empty := productmodels.ProductDoc{}
	if got := readinessFields(productV2ListingReadiness(empty)); got != "listing.title,package.weight,package,media.images" {
		t.Fatalf("empty doc must miss everything: %s", got)
	}
	full := productV2ListingReadiness(readyProductDoc())
	if full["ready"] != true || len(readinessMissing(full)) != 0 {
		t.Fatalf("complete doc must be ready: %+v", full)
	}
}

func TestProductV2AccountingReadiness(t *testing.T) {
	doc := productV2ProductDoc{ProductDoc: readyProductDoc()}
	if group := productV2AccountingReadiness(doc); group["ready"] != true {
		t.Fatalf("complete accounting must be ready: %+v", group)
	}

	inactive := false
	doc.IsActive = &inactive
	if got := readinessFields(productV2AccountingReadiness(doc)); got != "isactive" {
		t.Fatalf("inactive: %s", got)
	}

	doc.IsActive = nil
	doc.UnitCode = " "
	blank := ""
	th := "th"
	doc.Names = &[]models.NameX{{Code: &th, Name: &blank}}
	if got := readinessFields(productV2AccountingReadiness(doc)); got != "names,unitcode" {
		t.Fatalf("blank names/unitcode: %s", got)
	}

	doc.Names = nil
	if got := readinessFields(productV2AccountingReadiness(doc)); !strings.Contains(got, "names") {
		t.Fatalf("nil names: %s", got)
	}
}

func TestProductV2ModelsReadiness(t *testing.T) {
	if got := productV2ModelsReadiness(nil, nil); len(got) != 1 || got[0]["field"] != "models" || !strings.Contains(got[0]["message"], "บาร์โค้ด") {
		t.Fatalf("no models: %+v", got)
	}

	// ไม่มี tiers: ต้องมี isforsale อย่างน้อย 1 (E09)
	notForSale := []productV2ModelDoc{{Barcode: "B1", Listing: &barcodemodels.ProductBarcodeListing{}}}
	if got := productV2ModelsReadiness(nil, notForSale); len(got) != 1 || !strings.Contains(got[0]["message"], "เปิดขาย") {
		t.Fatalf("none for sale: %+v", got)
	}
	forSale := []productV2ModelDoc{{Barcode: "B1", Listing: &barcodemodels.ProductBarcodeListing{IsForSale: true}}}
	if got := productV2ModelsReadiness(nil, forSale); len(got) != 0 {
		t.Fatalf("no tiers + for sale must pass: %+v", got)
	}

	// มี tiers 2x2 = 4 ชุดผสม ต้องมี model ครบทุกชุด
	tiers := []productmodels.ProductListingTier{
		{XOrder: 0, Name: "สี", Options: []productmodels.ProductListingTierOption{{XOrder: 0, Name: "ดำ"}, {XOrder: 1, Name: "ขาว"}}},
		{XOrder: 1, Name: "ขนาด", Options: []productmodels.ProductListingTierOption{{XOrder: 0, Name: "M"}, {XOrder: 1, Name: "L"}}},
	}
	partial := []productV2ModelDoc{
		{Barcode: "B1", Listing: &barcodemodels.ProductBarcodeListing{TierIndex: []int{0, 0}, IsForSale: true}},
		{Barcode: "B2", Listing: &barcodemodels.ProductBarcodeListing{TierIndex: []int{0, 1}}},
		{Barcode: "B3", Listing: &barcodemodels.ProductBarcodeListing{TierIndex: []int{1, 0}}},
		{Barcode: "B4", Listing: &barcodemodels.ProductBarcodeListing{TierIndex: []int{9, 9}}}, // นอกช่วง ไม่นับ
	}
	got := productV2ModelsReadiness(tiers, partial)
	if len(got) != 1 || !strings.Contains(got[0]["message"], "ขาดอีก 1 ชุด") {
		t.Fatalf("partial tiers: %+v", got)
	}

	complete := append(partial[:3:3],
		productV2ModelDoc{Barcode: "B4", Listing: &barcodemodels.ProductBarcodeListing{TierIndex: []int{1, 1}}})
	if got := productV2ModelsReadiness(tiers, complete); len(got) != 0 {
		t.Fatalf("complete tiers must pass: %+v", got)
	}
}

func TestProductV2ListingReadinessFull(t *testing.T) {
	// ชื่อสั้นเกิน (E01) ต้องโดนจับแม้กรอกแล้ว
	doc := readyProductDoc()
	doc.Listing.Title = "สั้น"
	if got := readinessFields(productV2ListingReadinessFull(doc, nil)); !strings.Contains(got, "listing.title") {
		t.Fatalf("short title must be caught: %s", got)
	}
	// ready เมื่อทุกอย่างครบ (models ว่าง → ขาด models เท่านั้น เพราะยังไม่มีบาร์โค้ด)
	doc.Listing.Title = "ชื่อสำหรับลงขายที่ยาวพอ"
	if got := readinessFields(productV2ListingReadinessFull(doc, nil)); got != "models" {
		t.Fatalf("only models missing expected: %s", got)
	}
}

func TestProductV2ChannelReadiness(t *testing.T) {
	now := time.Now().UTC()
	activeShop := productV2ChannelShopDoc{IsActive: true}
	categoryMap := &productV2ChannelCategoryMapDoc{CategoryID: 123}

	if group := productV2ChannelReadiness(activeShop, "APPAREL", categoryMap, false, now); group["ready"] != true {
		t.Fatalf("all set must be ready: %+v", group)
	}
	if got := readinessFields(productV2ChannelReadiness(activeShop, "", nil, false, now)); got != "categoryid" {
		t.Fatalf("no categorycode: %s", got)
	}
	if got := readinessFields(productV2ChannelReadiness(activeShop, "APPAREL", nil, false, now)); got != "categoryid" {
		t.Fatalf("no category map: %s", got)
	}

	mandatory := &productV2ChannelCategoryMapDoc{CategoryID: 123, BrandMandatory: true}
	if got := readinessFields(productV2ChannelReadiness(activeShop, "APPAREL", mandatory, false, now)); got != "brandid" {
		t.Fatalf("brand mandatory without map: %s", got)
	}
	if group := productV2ChannelReadiness(activeShop, "APPAREL", mandatory, true, now); group["ready"] != true {
		t.Fatalf("brand mapped must be ready: %+v", group)
	}

	inactiveShop := productV2ChannelShopDoc{IsActive: false}
	if got := readinessFields(productV2ChannelReadiness(inactiveShop, "APPAREL", categoryMap, false, now)); got != "shop.isactive" {
		t.Fatalf("inactive shop: %s", got)
	}

	expiredShop := productV2ChannelShopDoc{IsActive: true}
	expiredShop.Auth = &struct {
		TokenExpiresAt time.Time `bson:"tokenexpiresat"`
	}{TokenExpiresAt: now.Add(-time.Hour)}
	if got := readinessFields(productV2ChannelReadiness(expiredShop, "APPAREL", categoryMap, false, now)); got != "shop.auth" {
		t.Fatalf("expired token: %s", got)
	}
}
