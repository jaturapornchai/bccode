package handlers

import (
	"strings"
	"testing"
	"time"

	commonmodels "smlcloudplatform/internal/models"
	productmodels "smlcloudplatform/internal/product/product/models"
	barcodemodels "smlcloudplatform/internal/product/productbarcode/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func tierOption(name string) productV2TierOptionReq {
	return productV2TierOptionReq{Name: name}
}

func tierReq(name string, options ...string) productV2TierReq {
	tier := productV2TierReq{Name: name}
	for _, option := range options {
		tier.Options = append(tier.Options, tierOption(option))
	}
	return tier
}

// TestValidateProductV2Tiers_E03E04E12 — จำนวนชั้น ชุดผสม ชื่อซ้ำ และกฎรูปของตัวเลือก
func TestValidateProductV2Tiers_E03E04E12(t *testing.T) {
	limits := productV2DefaultLimits
	manyOptions := make([]string, 0, 26)
	for i := 0; i < 26; i++ {
		manyOptions = append(manyOptions, string(rune('A'+i)))
	}

	cases := []struct {
		name  string
		tiers []productV2TierReq
		want  string
	}{
		{"ไม่ส่งชั้น", nil, "tiers:REQUIRED"},
		{"เกิน 2 ชั้น", []productV2TierReq{tierReq("สี", "ดำ"), tierReq("ขนาด", "M"), tierReq("วัสดุ", "ผ้า")}, "tiers:TOO_MANY"},
		{"ชั้นไม่มีชื่อ", []productV2TierReq{tierReq("  ", "ดำ")}, "tiers[0].name:REQUIRED"},
		{"ชื่อชั้นซ้ำ", []productV2TierReq{tierReq("สี", "ดำ"), tierReq("สี", "M")}, "tiers[1].name:DUPLICATE"},
		{"ชั้นไม่มีตัวเลือก", []productV2TierReq{{Name: "สี"}}, "tiers[0].options:REQUIRED"},
		{"ตัวเลือกซ้ำ", []productV2TierReq{tierReq("สี", "ดำ", " ดำ ")}, "tiers[0].options[1].name:DUPLICATE"},
		{"ชื่อชั้นยาวเกิน", []productV2TierReq{tierReq(strings.Repeat("ก", limits.TierNameMax+1), "ดำ")}, "tiers[0].name:TOO_LONG"},
		{"ชื่อตัวเลือกยาวเกิน", []productV2TierReq{tierReq("สี", strings.Repeat("ก", limits.TierOptionNameMax+1))}, "tiers[0].options[0].name:TOO_LONG"},
		{"ชุดผสมเกิน 50", []productV2TierReq{tierReq("สี", manyOptions...), tierReq("ขนาด", "S", "M", "L")}, "tiers:TOO_MANY_COMBINATIONS"},
		{"ผ่าน", []productV2TierReq{tierReq("สี", "ดำ", "ขาว"), tierReq("ขนาด", "M", "L")}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := codesOf(validateProductV2Tiers(tc.tiers, limits))
			if tc.want == "" {
				if got != "" {
					t.Fatalf("ต้องไม่มีข้อผิดพลาด แต่ได้ %q", got)
				}
				return
			}
			if !strings.Contains(got, tc.want) {
				t.Fatalf("ต้องมี %q แต่ได้ %q", tc.want, got)
			}
		})
	}
}

// TestValidateProductV2Tiers_OptionImages — E12: รูปเฉพาะชั้นแรก ครบทุกตัวเลือก และต้องมีรูปย่อ
func TestValidateProductV2Tiers_OptionImages(t *testing.T) {
	limits := productV2DefaultLimits

	partial := []productV2TierReq{{Name: "สี", Options: []productV2TierOptionReq{
		{Name: "ดำ", ImageURI: "/goapi/s3/file/a.webp", ImageURIThumb: "/goapi/s3/file/a_thumb.webp"},
		{Name: "ขาว"},
	}}}
	if got := codesOf(validateProductV2Tiers(partial, limits)); !strings.Contains(got, "tiers[0].options:INCOMPLETE") {
		t.Fatalf("รูปไม่ครบต้องแจ้ง INCOMPLETE แต่ได้ %q", got)
	}

	noThumb := []productV2TierReq{{Name: "สี", Options: []productV2TierOptionReq{
		{Name: "ดำ", ImageURI: "/goapi/s3/file/a.webp"},
	}}}
	if got := codesOf(validateProductV2Tiers(noThumb, limits)); !strings.Contains(got, "tiers[0].options[0].imageurithumb:REQUIRED") {
		t.Fatalf("รูปไม่มี thumb ต้องแจ้ง REQUIRED แต่ได้ %q", got)
	}

	secondTier := []productV2TierReq{
		tierReq("สี", "ดำ"),
		{Name: "ขนาด", Options: []productV2TierOptionReq{{Name: "M", ImageURI: "/goapi/s3/file/m.webp", ImageURIThumb: "/goapi/s3/file/m_thumb.webp"}}},
	}
	if got := codesOf(validateProductV2Tiers(secondTier, limits)); !strings.Contains(got, "tiers[1].options[0].imageuri:NOT_ALLOWED") {
		t.Fatalf("รูปในชั้นที่สองต้องถูกปฏิเสธ แต่ได้ %q", got)
	}

	full := []productV2TierReq{{Name: "สี", Options: []productV2TierOptionReq{
		{Name: "ดำ", ImageURI: "/goapi/s3/file/a.webp", ImageURIThumb: "/goapi/s3/file/a_thumb.webp"},
		{Name: "ขาว", ImageURI: "/goapi/s3/file/b.webp", ImageURIThumb: "/goapi/s3/file/b_thumb.webp"},
	}}}
	if got := codesOf(validateProductV2Tiers(full, limits)); got != "" {
		t.Fatalf("รูปครบต้องผ่าน แต่ได้ %q", got)
	}
}

// TestValidateProductV2TierModels_E09E11 — ชุดผสมต้องครบ ไม่ซ้ำ อยู่ในช่วง และต้องมีบาร์โค้ดหรือขอให้สร้าง
func TestValidateProductV2TierModels_E09E11(t *testing.T) {
	tiers := []productV2TierReq{tierReq("สี", "ดำ", "ขาว"), tierReq("ขนาด", "M", "L")}
	full := []productV2TierModelReq{
		{TierIndex: []int{0, 0}, Barcode: "A1"},
		{TierIndex: []int{0, 1}, Barcode: "A2"},
		{TierIndex: []int{1, 0}, Barcode: "A3"},
		{TierIndex: []int{1, 1}, Generate: true},
	}

	if got := codesOf(validateProductV2TierModels(tiers, full)); got != "" {
		t.Fatalf("ชุดครบต้องผ่าน แต่ได้ %q", got)
	}
	if got := codesOf(validateProductV2TierModels(tiers, full[:3])); !strings.Contains(got, "models:INCOMPLETE") {
		t.Fatalf("ชุดไม่ครบต้องแจ้ง INCOMPLETE แต่ได้ %q", got)
	}
	if got := codesOf(validateProductV2TierModels(tiers, nil)); !strings.Contains(got, "models:REQUIRED") {
		t.Fatalf("ไม่ส่ง models ต้องแจ้ง REQUIRED แต่ได้ %q", got)
	}

	shortIndex := []productV2TierModelReq{{TierIndex: []int{0}, Barcode: "A1"}}
	if got := codesOf(validateProductV2TierModels(tiers, shortIndex)); !strings.Contains(got, "models[0].tierindex:INVALID_VALUE") {
		t.Fatalf("tierindex สั้นต้องแจ้ง INVALID_VALUE แต่ได้ %q", got)
	}

	outOfRange := []productV2TierModelReq{{TierIndex: []int{0, 5}, Barcode: "A1"}}
	if got := codesOf(validateProductV2TierModels(tiers, outOfRange)); !strings.Contains(got, "models[0].tierindex[1]:OUT_OF_RANGE") {
		t.Fatalf("tierindex เกินช่วงต้องแจ้ง OUT_OF_RANGE แต่ได้ %q", got)
	}

	duplicated := append([]productV2TierModelReq{}, full...)
	duplicated[3] = productV2TierModelReq{TierIndex: []int{0, 0}, Barcode: "A9"}
	if got := codesOf(validateProductV2TierModels(tiers, duplicated)); !strings.Contains(got, "models[3].tierindex:DUPLICATE") {
		t.Fatalf("tierindex ซ้ำต้องแจ้ง DUPLICATE แต่ได้ %q", got)
	}

	sameBarcode := append([]productV2TierModelReq{}, full...)
	sameBarcode[3] = productV2TierModelReq{TierIndex: []int{1, 1}, Barcode: "a1"}
	if got := codesOf(validateProductV2TierModels(tiers, sameBarcode)); !strings.Contains(got, "models[3].barcode:DUPLICATE") {
		t.Fatalf("บาร์โค้ดซ้ำต้องแจ้ง DUPLICATE แต่ได้ %q", got)
	}

	noBarcode := append([]productV2TierModelReq{}, full...)
	noBarcode[3] = productV2TierModelReq{TierIndex: []int{1, 1}}
	if got := codesOf(validateProductV2TierModels(tiers, noBarcode)); !strings.Contains(got, "models[3].barcode:REQUIRED") {
		t.Fatalf("ไม่มีบาร์โค้ดและไม่ขอสร้างต้องแจ้ง REQUIRED แต่ได้ %q", got)
	}
}

// TestValidateProductV2TierBarcodesExist_E10 — บาร์โค้ดต้องเป็นของสินค้าตัวนี้
func TestValidateProductV2TierBarcodesExist_E10(t *testing.T) {
	existing := map[string]productV2ItemBarcode{"a1": {Barcode: "A1"}}
	models := []productV2TierModelReq{{Barcode: "A1"}, {Barcode: "ZZ"}, {Generate: true}}
	got := codesOf(validateProductV2TierBarcodesExist(models, existing))
	if strings.Contains(got, "models[0]") || strings.Contains(got, "models[2]") {
		t.Fatalf("บาร์โค้ดที่มีอยู่/ที่ให้ระบบสร้างต้องไม่ผิด แต่ได้ %q", got)
	}
	if !strings.Contains(got, "models[1].barcode:NOT_FOUND") {
		t.Fatalf("บาร์โค้ดต่างสินค้าต้องแจ้ง NOT_FOUND แต่ได้ %q", got)
	}
}

func TestScanProductV2TierPayload(t *testing.T) {
	accounting, unknown := scanProductV2TierPayload(map[string]any{
		"itemcode":     "P1",
		"__v":          1,
		"tiers":        []any{},
		"models":       []any{},
		"itemunitcode": "PCS",
		"unitcode":     "BOX",
		"names":        []any{},
		"foo":          1,
	})
	if strings.Join(accounting, ",") != "names,unitcode" {
		t.Fatalf("ฟิลด์ชั้นบัญชีต้องเป็น names,unitcode แต่ได้ %v", accounting)
	}
	if strings.Join(unknown, ",") != "foo" {
		t.Fatalf("ฟิลด์ที่ไม่รู้จักต้องเป็น foo แต่ได้ %v", unknown)
	}
}

func TestBuildProductV2TierDocs(t *testing.T) {
	docs := buildProductV2TierDocs([]productV2TierReq{
		{Name: " สี ", Options: []productV2TierOptionReq{{Name: " ดำ "}, {Name: "ขาว"}}},
		{Name: "ขนาด", Options: []productV2TierOptionReq{{Name: "M"}}},
	})
	if len(docs) != 2 || docs[0].XOrder != 0 || docs[1].XOrder != 1 {
		t.Fatalf("xorder ของชั้นต้องเรียง 0,1 แต่ได้ %+v", docs)
	}
	if docs[0].Name != "สี" || docs[0].Options[0].Name != "ดำ" {
		t.Fatalf("ต้องตัดช่องว่างหัวท้าย แต่ได้ %+v", docs[0])
	}
	if docs[0].Options[1].XOrder != 1 {
		t.Fatalf("xorder ของตัวเลือกต้องเรียงตามลำดับ แต่ได้ %+v", docs[0].Options)
	}
}

// TestBuildProductV2TierWritePlan — ผูกบาร์โค้ดเดิม, แทรกบาร์โค้ดใหม่, ปิดขายตัวที่หลุดชุดผสม
func TestBuildProductV2TierWritePlan(t *testing.T) {
	tiersReq := []productV2TierReq{tierReq("สี", "ดำ", "ขาว")}
	tiers := buildProductV2TierDocs(tiersReq)
	req := productV2TierRequest{
		ItemCode: "P1",
		Tiers:    tiersReq,
		Models: []productV2TierModelReq{
			{TierIndex: []int{0}, Barcode: "A1"},
			{TierIndex: []int{1}, Generate: true},
		},
	}
	existing := map[string]productV2ItemBarcode{
		"a1": {Barcode: "A1"},
		"a9": {Barcode: "A9", Listing: &barcodemodels.ProductBarcodeListing{TierIndex: []int{1}, IsForSale: true}},
		"a8": {Barcode: "A8"}, // บาร์โค้ดที่ไม่เคยผูกตัวเลือก ต้องไม่ถูกแตะ
	}
	current := newTierProductDoc("demo", "C01", "P1", "เสื้อยืด")
	unit := productmodels.ProductUnitConversion{UnitCode: "PCS", DivideValue: 1, StandValue: 1}
	now := time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)

	plan := buildProductV2TierWritePlan(req, tiers, map[int]string{1: "2001234567895"}, existing, current, "PCS", unit, "tester", now)
	if len(plan) != 3 {
		t.Fatalf("ต้องมี 3 คำสั่ง (อัปเดต 1 + แทรก 1 + ปิดขาย 1) แต่ได้ %d", len(plan))
	}

	inserts, updates := 0, 0
	sawDisable := false
	for _, model := range plan {
		switch m := model.(type) {
		case *mongo.InsertOneModel:
			inserts++
			doc, ok := m.Document.(bson.M)
			if !ok {
				t.Fatalf("เอกสารที่แทรกต้องเป็น bson.M แต่ได้ %T", m.Document)
			}
			if doc["barcode"] != "2001234567895" || doc["itemcode"] != "P1" || doc["ismainbarcode"] != false {
				t.Fatalf("เอกสารบาร์โค้ดใหม่ไม่ถูกต้อง: %v", doc)
			}
			if _, hasPrices := doc["prices"]; hasPrices {
				t.Fatal("บาร์โค้ดใหม่ต้องไม่คัดลอกราคา (ตั้งราคาที่หน้าสินค้า)")
			}
			listing, ok := doc["listing"].(bson.M)
			if !ok || listing["isforsale"] != true {
				t.Fatalf("บาร์โค้ดใหม่ต้องเปิดขาย: %v", doc["listing"])
			}
		case *mongo.UpdateOneModel:
			updates++
			update, ok := m.Update.(bson.M)
			if !ok {
				t.Fatalf("คำสั่งอัปเดตต้องเป็น bson.M แต่ได้ %T", m.Update)
			}
			set, ok := update["$set"].(bson.M)
			if !ok {
				t.Fatalf("ต้องเขียนด้วย $set เท่านั้น แต่ได้ %v", update)
			}
			if set["listing.isforsale"] == false {
				sawDisable = true
			}
			filter, ok := m.Filter.(bson.M)
			if !ok || filter["holdingcode"] != "demo" || filter["businesscode"] != "C01" {
				t.Fatalf("ตัวกรองต้องจำกัดบริษัท แต่ได้ %v", m.Filter)
			}
			if filter["barcode"] == "A8" {
				t.Fatal("บาร์โค้ดที่ไม่เคยผูกตัวเลือกต้องไม่ถูกแก้")
			}
		}
	}
	if inserts != 1 || updates != 2 || !sawDisable {
		t.Fatalf("แผนการเขียนไม่ถูกต้อง inserts=%d updates=%d disable=%v", inserts, updates, sawDisable)
	}
}

func TestProductV2BarcodeNames(t *testing.T) {
	current := newTierProductDoc("demo", "C01", "P1", "เสื้อยืด")
	tiers := buildProductV2TierDocs([]productV2TierReq{tierReq("สี", "ดำ"), tierReq("ขนาด", "M")})
	names := productV2BarcodeNames(current, tiers, []int{0, 0})
	if len(names) != 1 || names[0].Name == nil || *names[0].Name != "เสื้อยืด ดำ / M" {
		t.Fatalf("ชื่อบาร์โค้ดใหม่ไม่ถูกต้อง: %+v", names)
	}
}

// TestProductV2CheckDigit — เทียบกับเลข 13 หลักมาตรฐานที่ทราบค่าหลักตรวจสอบ
func TestProductV2CheckDigit(t *testing.T) {
	cases := map[string]string{
		"400638133393": "1",
		"978030640615": "7",
		"590123412345": "7",
	}
	for body, want := range cases {
		if got := productV2CheckDigit(body); got != want {
			t.Fatalf("หลักตรวจสอบของ %s ต้องเป็น %s แต่ได้ %s", body, want, got)
		}
	}
}

func TestProductV2RandomBarcode(t *testing.T) {
	for i := 0; i < 50; i++ {
		barcode, err := productV2RandomBarcode()
		if err != nil {
			t.Fatal(err)
		}
		if len(barcode) != 13 {
			t.Fatalf("บาร์โค้ดต้องมี 13 หลัก แต่ได้ %q", barcode)
		}
		if !strings.HasPrefix(barcode, productV2TierBarcodePrefix) {
			t.Fatalf("บาร์โค้ดต้องขึ้นต้นด้วย %s แต่ได้ %q", productV2TierBarcodePrefix, barcode)
		}
		for _, ch := range barcode {
			if ch < '0' || ch > '9' {
				t.Fatalf("บาร์โค้ดต้องเป็นตัวเลขล้วน แต่ได้ %q", barcode)
			}
		}
		if productV2CheckDigit(barcode[:12]) != barcode[12:] {
			t.Fatalf("หลักตรวจสอบไม่ถูกต้อง: %q", barcode)
		}
	}
}

func TestProductV2IsTransactionUnsupported(t *testing.T) {
	if productV2IsTransactionUnsupported(productV2ErrVersionConflict) {
		t.Fatal("__v ไม่ตรง ต้องไม่ถูกมองว่าเซิร์ฟเวอร์ไม่รองรับ transaction")
	}
	if !productV2IsTransactionUnsupported(mongo.CommandError{Message: "Transaction numbers are only allowed on a replica set member or mongos"}) {
		t.Fatal("ข้อความของ MongoDB แบบ standalone ต้องถูกตรวจจับได้")
	}
}

func newTierProductDoc(holdingCode, businessCode, code, name string) productV2ProductDoc {
	doc := productV2ProductDoc{}
	doc.HoldingCode = holdingCode
	doc.BusinessCode = businessCode
	doc.Code = code
	doc.GuidFixed = "guid-" + code
	doc.UnitCode = "PCS"
	langCode, langName := "th", name
	names := []commonmodels.NameX{{Code: &langCode, Name: &langName}}
	doc.Names = &names
	return doc
}
