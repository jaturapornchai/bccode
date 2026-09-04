package handlers

// ตัวตรวจฟิลด์ต้องห้าม (E24) สำหรับ item/update-listing — สัญญา §3
// payload ถูกถอดเป็น map ก่อน เพื่อให้เห็นทุก key ที่ส่งมา ไม่ให้ฟิลด์บัญชีหลุดผ่านแบบเงียบ ๆ

import (
	"reflect"
	"sort"
	"strings"

	productmodels "smlcloudplatform/internal/product/product/models"
)

// productV2WritableTop — key ระดับบนสุดที่ item/update-listing รับ (§3 ชั้นลงขาย + key ควบคุม)
var productV2WritableTop = map[string]struct{}{
	"itemcode":      {},
	"holdingcode":   {},
	"businesscode":  {},
	"__v":           {},
	"listing":       {},
	"package":       {},
	"description":   {},
	"imageuri":      {},
	"imageurithumb": {},
	"images":        {},
	"videos":        {},
}

// productV2PackageTopKeys — packageweight/... ต้องส่งผ่าน package{} เท่านั้น (ตรวจ E05 เป็นชุด)
var productV2PackageTopKeys = map[string]struct{}{
	"packageweight": {},
	"packagelength": {},
	"packagewidth":  {},
	"packageheight": {},
}

// productV2ExtraAccountingKeys — ฟิลด์ชั้นบัญชีที่ reflect ไม่เห็น: audit (json tag "-") และ isactive (§3 มีในสัญญา/แบบจำลอง แต่ ProductDoc ไม่มี)
var productV2ExtraAccountingKeys = []string{
	"_id", "id", "createdat", "createdby", "createdbyname",
	"updatedat", "updatedby", "updatedbyname",
	"deletedat", "deletedby", "deletedbyname", "isdeleted",
	"isactive",
}

// productV2AccountingLabels — ชื่อไทยสำหรับข้อความ error (ฟิลด์ที่ไม่มีในนี้ใช้ชื่อ key ตรง ๆ)
var productV2AccountingLabels = map[string]string{
	"code":             "รหัสสินค้า",
	"names":            "ชื่อสินค้า",
	"guidfixed":        "รหัสอ้างอิงภายใน",
	"unitcode":         "หน่วยนับ",
	"unitguid":         "หน่วยนับ",
	"unitnames":        "ชื่อหน่วยนับ",
	"unitconversions":  "อัตราแปลงหน่วย",
	"standvalue":       "ค่าตั้งต้นหน่วย",
	"dividevalue":      "ตัวหารหน่วย",
	"condition":        "เงื่อนไขหน่วย",
	"itemtype":         "ประเภทสินค้า",
	"materialtype":     "ประเภทวัตถุดิบ",
	"taxtype":          "ประเภทภาษี",
	"vattype":          "ประเภทภาษีมูลค่าเพิ่ม",
	"groupcode":        "กลุ่มสินค้า",
	"groupnames":       "ชื่อกลุ่มสินค้า",
	"categorycode":     "หมวดสินค้า",
	"categorynames":    "ชื่อหมวดสินค้า",
	"classcode":        "ชั้นสินค้า",
	"brandcode":        "ยี่ห้อ",
	"brandnames":       "ชื่อยี่ห้อ",
	"bom":              "สูตรการผลิต",
	"isusesubbarcodes": "การใช้บาร์โค้ดย่อย",
	"refbarcodes":      "บาร์โค้ดอ้างอิง",
	"orderpoint":       "จุดสั่งซื้อ",
	"minpoint":         "ยอดต่ำสุด",
	"maxpoint":         "ยอดสูงสุด",
	"isactive":         "สถานะใช้งาน",
	"isdeleted":        "สถานะลบ",
}

// productAccountingFields — ทุก key ของเอกสาร product ที่ v2 ห้ามแก้ = key ทั้งหมดใน ProductDoc (json+bson tag) − ชั้นลงขาย
var productAccountingFields = buildProductAccountingFields()

func buildProductAccountingFields() map[string]struct{} {
	keys := map[string]struct{}{}
	collectStructKeys(reflect.TypeOf(productmodels.ProductDoc{}), keys)
	for _, key := range productV2ExtraAccountingKeys {
		keys[key] = struct{}{}
	}
	for key := range productV2WritableTop {
		delete(keys, key)
	}
	for key := range productV2PackageTopKeys {
		delete(keys, key)
	}
	return keys
}

// collectStructKeys — เก็บชื่อ json/bson tag ระดับบนสุด (ลงลึกเฉพาะ struct ที่ inline)
func collectStructKeys(t reflect.Type, keys map[string]struct{}) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		bsonTag := field.Tag.Get("bson")
		if strings.Contains(bsonTag, "inline") && field.Type.Kind() == reflect.Struct {
			collectStructKeys(field.Type, keys)
			continue
		}
		for _, tag := range []string{field.Tag.Get("json"), bsonTag} {
			name := strings.Split(tag, ",")[0]
			if name != "" && name != "-" {
				keys[name] = struct{}{}
			}
		}
	}
}

// scanProductV2Payload — คืน key ชั้นบัญชี (→ E24) และ key ที่ไม่รู้จัก (→ VALIDATION_FAILED) เรียงตามตัวอักษร
func scanProductV2Payload(raw map[string]any) (accounting []string, unknown []string) {
	for key := range raw {
		if _, ok := productV2WritableTop[key]; ok {
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

// productV2AccountingFieldErrors — สร้าง message + fields[] สำหรับ 422 ACCOUNTING_FIELD_READONLY
func productV2AccountingFieldErrors(keys []string) (string, []productV2FieldError) {
	labels := make([]string, 0, len(keys))
	fields := make([]productV2FieldError, 0, len(keys))
	for _, key := range keys {
		label, ok := productV2AccountingLabels[key]
		if !ok {
			label = key
		}
		labels = append(labels, label)
		fields = append(fields, productV2FieldError{Field: key, Code: "READONLY", Message: label})
	}
	return productV2MsgAccountingPrefix + strings.Join(labels, ", "), fields
}
