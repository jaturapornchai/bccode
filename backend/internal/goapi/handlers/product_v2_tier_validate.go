package handlers

// ตัวตรวจสอบชั้นตัวเลือกสินค้า (tier/*) — สัญญา §5.4 และกฎ E03 E04 E09 E10 E11 E12
// ทุกฟังก์ชันในไฟล์นี้เป็น pure function (ไม่แตะฐานข้อมูล) เพื่อให้ตรวจให้ครบก่อนเปิด transaction

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// productV2TierOptionReq — ตัวเลือกหนึ่งค่าในชั้น (เช่น "แดง")
type productV2TierOptionReq struct {
	Name          string `json:"name"`
	ImageURI      string `json:"imageuri"`
	ImageURIThumb string `json:"imageurithumb"`
}

// productV2TierReq — ชั้นตัวเลือกหนึ่งชั้น (เช่น "สี"); xorder ระบบกำหนดตามลำดับที่ส่งมา
type productV2TierReq struct {
	Name    string                   `json:"name"`
	Options []productV2TierOptionReq `json:"options"`
}

// productV2TierModelReq — ตัวเลือกที่ขายได้หนึ่งชุดผสม (1 ชุด = 1 บาร์โค้ด)
type productV2TierModelReq struct {
	TierIndex []int  `json:"tierindex"`
	Barcode   string `json:"barcode"`
	Generate  bool   `json:"generate"`
}

// productV2TierRequest — body ของ tier/init และ tier/update (รูปแบบเดียวกัน คือสถานะสุดท้ายทั้งชุด)
type productV2TierRequest struct {
	ItemCode     string                  `json:"itemcode"`
	HoldingCode  string                  `json:"holdingcode"`
	BusinessCode string                  `json:"businesscode"`
	Version      *int64                  `json:"__v"`
	Tiers        []productV2TierReq      `json:"tiers"`
	Models       []productV2TierModelReq `json:"models"`
	ItemUnitCode string                  `json:"itemunitcode"`
}

// validateProductV2Tiers — E03 (จำนวนชั้น), E04 (จำนวนชุดผสม/ชื่อซ้ำ/ความยาว), E12 (รูปเฉพาะชั้นแรกและต้องครบ)
func validateProductV2Tiers(tiers []productV2TierReq, limits productV2Limits) []productV2FieldError {
	errs := []productV2FieldError{}
	if len(tiers) == 0 {
		return append(errs, productV2FieldError{"tiers", "REQUIRED", "กรุณาระบุชั้นตัวเลือกอย่างน้อย 1 ชั้น"})
	}
	if len(tiers) > limits.TierMax {
		errs = append(errs, productV2FieldError{"tiers", "TOO_MANY", fmt.Sprintf("ใส่ชั้นตัวเลือกได้ไม่เกิน %d ชั้น", limits.TierMax)})
	}

	combinations := 1
	tierNames := map[string]bool{}
	for i, tier := range tiers {
		field := fmt.Sprintf("tiers[%d]", i)
		name := strings.TrimSpace(tier.Name)
		switch {
		case name == "":
			errs = append(errs, productV2FieldError{field + ".name", "REQUIRED", "กรุณาตั้งชื่อชั้นตัวเลือก เช่น สี หรือ ขนาด"})
		case utf8.RuneCountInString(name) > limits.TierNameMax:
			errs = append(errs, productV2FieldError{field + ".name", "TOO_LONG", fmt.Sprintf("ชื่อชั้นตัวเลือกยาวเกิน %d ตัวอักษร กรุณาย่อให้สั้นลง", limits.TierNameMax)})
		case tierNames[strings.ToLower(name)]:
			errs = append(errs, productV2FieldError{field + ".name", "DUPLICATE", "ชื่อชั้นตัวเลือกซ้ำกัน กรุณาตั้งชื่อให้ต่างกัน"})
		default:
			tierNames[strings.ToLower(name)] = true
		}

		if len(tier.Options) == 0 {
			errs = append(errs, productV2FieldError{field + ".options", "REQUIRED", "กรุณาใส่ตัวเลือกในชั้นนี้อย่างน้อย 1 ตัวเลือก"})
			continue
		}
		combinations *= len(tier.Options)
		errs = append(errs, validateProductV2TierOptions(i, tier.Options, limits)...)
	}

	if combinations > limits.TierCombinationMax {
		errs = append(errs, productV2FieldError{"tiers", "TOO_MANY_COMBINATIONS", fmt.Sprintf("ตัวเลือกรวมกันได้ %d ชุด เกิน %d ชุดที่รองรับ กรุณาลดตัวเลือก", combinations, limits.TierCombinationMax)})
	}
	return errs
}

// validateProductV2TierOptions — ชื่อตัวเลือกในชั้นเดียวกันห้ามซ้ำ และรูปใส่ได้เฉพาะชั้นแรก (E04, E12)
func validateProductV2TierOptions(tierIdx int, options []productV2TierOptionReq, limits productV2Limits) []productV2FieldError {
	errs := []productV2FieldError{}
	seen := map[string]bool{}
	withImage := 0
	for j, option := range options {
		field := fmt.Sprintf("tiers[%d].options[%d]", tierIdx, j)
		name := strings.TrimSpace(option.Name)
		switch {
		case name == "":
			errs = append(errs, productV2FieldError{field + ".name", "REQUIRED", "กรุณาตั้งชื่อตัวเลือก"})
		case utf8.RuneCountInString(name) > limits.TierOptionNameMax:
			errs = append(errs, productV2FieldError{field + ".name", "TOO_LONG", fmt.Sprintf("ชื่อตัวเลือกยาวเกิน %d ตัวอักษร กรุณาย่อให้สั้นลง", limits.TierOptionNameMax)})
		case seen[strings.ToLower(name)]:
			errs = append(errs, productV2FieldError{field + ".name", "DUPLICATE", "ชื่อตัวเลือกซ้ำกันในชั้นเดียวกัน กรุณาตั้งชื่อให้ต่างกัน"})
		default:
			seen[strings.ToLower(name)] = true
		}

		uri := strings.TrimSpace(option.ImageURI)
		thumb := strings.TrimSpace(option.ImageURIThumb)
		if uri == "" && thumb == "" {
			continue
		}
		if tierIdx > 0 {
			errs = append(errs, productV2FieldError{field + ".imageuri", "NOT_ALLOWED", "ใส่รูปได้เฉพาะตัวเลือกชั้นแรกเท่านั้น"})
			continue
		}
		withImage++
		errs = append(errs, validateURI(field+".imageuri", uri, true, limits)...)
		if thumb == "" {
			errs = append(errs, productV2FieldError{field + ".imageurithumb", "REQUIRED", "ต้องมีรูปย่อคู่กับรูปตัวเลือกนี้"})
		} else {
			errs = append(errs, validateURI(field+".imageurithumb", thumb, true, limits)...)
		}
	}
	// E12 — รูปของชั้นแรกต้องใส่ครบทุกตัวเลือก หรือไม่ใส่เลย
	if tierIdx == 0 && withImage > 0 && withImage != len(options) {
		errs = append(errs, productV2FieldError{fmt.Sprintf("tiers[%d].options", tierIdx), "INCOMPLETE", "ถ้าใส่รูปตัวเลือก ต้องใส่ให้ครบทุกตัวเลือกในชั้นแรก หรือไม่ใส่เลย"})
	}
	return errs
}

// validateProductV2TierModels — E09 (ครบทุกชุดผสม), E11 (tierindex ถูกต้องและไม่ซ้ำ), และต้องระบุบาร์โค้ดหรือให้ระบบสร้าง
func validateProductV2TierModels(tiers []productV2TierReq, models []productV2TierModelReq) []productV2FieldError {
	errs := []productV2FieldError{}
	expected := 1
	for _, tier := range tiers {
		expected *= len(tier.Options)
	}
	if len(models) == 0 {
		return append(errs, productV2FieldError{"models", "REQUIRED", "กรุณาระบุบาร์โค้ดของทุกชุดตัวเลือก"})
	}

	seen := map[string]bool{}
	seenBarcode := map[string]bool{}
	for i, model := range models {
		field := fmt.Sprintf("models[%d]", i)
		if len(model.TierIndex) != len(tiers) {
			errs = append(errs, productV2FieldError{field + ".tierindex", "INVALID_VALUE", fmt.Sprintf("tierindex ต้องมี %d ค่า (เท่ากับจำนวนชั้นตัวเลือก)", len(tiers))})
			continue
		}
		inRange := true
		for level, idx := range model.TierIndex {
			if idx < 0 || idx >= len(tiers[level].Options) {
				errs = append(errs, productV2FieldError{fmt.Sprintf("%s.tierindex[%d]", field, level), "OUT_OF_RANGE", "ไม่พบตัวเลือกที่ตำแหน่งนี้ กรุณาเลือกใหม่"})
				inRange = false
			}
		}
		if !inRange {
			continue
		}
		key := productV2TierIndexKey(model.TierIndex)
		if seen[key] {
			errs = append(errs, productV2FieldError{field + ".tierindex", "DUPLICATE", "ชุดตัวเลือกนี้ถูกระบุซ้ำ กรุณาระบุชุดละหนึ่งบาร์โค้ด"})
			continue
		}
		seen[key] = true

		barcode := strings.TrimSpace(model.Barcode)
		if barcode == "" && !model.Generate {
			errs = append(errs, productV2FieldError{field + ".barcode", "REQUIRED", "กรุณาเลือกบาร์โค้ด หรือให้ระบบสร้างให้ (generate)"})
			continue
		}
		if barcode == "" {
			continue
		}
		if seenBarcode[strings.ToLower(barcode)] {
			errs = append(errs, productV2FieldError{field + ".barcode", "DUPLICATE", "บาร์โค้ดนี้ถูกใช้กับตัวเลือกอื่นแล้ว กรุณาใช้บาร์โค้ดละหนึ่งตัวเลือก"})
			continue
		}
		seenBarcode[strings.ToLower(barcode)] = true
	}

	if len(seen) != expected {
		errs = append(errs, productV2FieldError{"models", "INCOMPLETE", fmt.Sprintf("ต้องระบุบาร์โค้ดให้ครบ %d ชุดตัวเลือก (ตอนนี้ %d ชุด)", expected, len(seen))})
	}
	return errs
}
