package handlers

// validator ล้วน (ไม่แตะ Mongo) สำหรับ item/update-listing — กฎ E01 E02 E05 E07 E08 E20 (§6)
// รหัสฟิลด์: REQUIRED TOO_SHORT TOO_LONG INVALID_CHAR INVALID_VALUE RANGE_OVERLAP OUT_OF_RANGE PACKAGE_INCOMPLETE DUPLICATE

import (
	"fmt"
	"math"
	"math/big"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	productmodels "smlcloudplatform/internal/product/product/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func validateProductV2Update(req productV2UpdateListingRequest, modelPrices []float64, limits productV2Limits) []productV2FieldError {
	var errs []productV2FieldError
	if req.Listing != nil {
		l := req.Listing
		if l.Title != nil {
			errs = append(errs, validateListingTitle(*l.Title, limits)...)
		}
		if l.Description != nil {
			errs = append(errs, validateListingDescription(*l.Description, limits)...)
		}
		if l.Condition != nil {
			errs = append(errs, validateCondition(*l.Condition)...)
		}
		if l.Wholesale != nil {
			if len(*l.Wholesale) > limits.WholesaleCountMax {
				errs = append(errs, productV2FieldError{"listing.wholesale", "TOO_LONG", fmt.Sprintf("ราคาขายส่งใส่ได้ไม่เกิน %d ช่วง", limits.WholesaleCountMax)})
			}
			errs = append(errs, validateWholesale(*l.Wholesale, productV2NormalPrice(modelPrices))...)
		}
		if l.SizeChart != nil {
			errs = append(errs, validateURI("listing.sizechart.uri", l.SizeChart.URI, false, limits)...)
			errs = append(errs, validateURI("listing.sizechart.urithumb", l.SizeChart.URIThumb, false, limits)...)
		}
		errs = append(errs, validatePurchaseLimitAndPreorder(l.PurchaseLimit, l.Preorder)...)
	}
	if req.Description != nil {
		errs = append(errs, validateProductDescription(*req.Description, limits)...)
	}
	if req.Package != nil {
		errs = append(errs, validatePackage(*req.Package, limits)...)
	}
	if req.ImageURI != nil {
		errs = append(errs, validateURI("imageuri", *req.ImageURI, false, limits)...)
	}
	if req.ImageURIThumb != nil {
		errs = append(errs, validateURI("imageurithumb", *req.ImageURIThumb, false, limits)...)
	}
	if req.Images != nil {
		errs = append(errs, validateImages(*req.Images, limits)...)
	}
	if req.Videos != nil {
		errs = append(errs, validateVideos(*req.Videos, limits)...)
	}
	return errs
}

// validateURI — uri ต้องเป็น path ในระบบ (ขึ้นต้น "/") หรือ http(s) เท่านั้น และยาวไม่เกินกำหนด (กัน data:/javascript:)
func validateURI(field, uri string, required bool, limits productV2Limits) []productV2FieldError {
	if uri == "" {
		if required {
			return []productV2FieldError{{field, "REQUIRED", "กรุณาระบุที่อยู่ไฟล์"}}
		}
		return nil
	}
	if len(uri) > limits.URIMax {
		return []productV2FieldError{{field, "TOO_LONG", fmt.Sprintf("ที่อยู่ไฟล์ยาวเกิน %d ตัวอักษร", limits.URIMax)}}
	}
	lower := strings.ToLower(uri)
	if !strings.HasPrefix(uri, "/") && !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
		return []productV2FieldError{{field, "INVALID_VALUE", "ที่อยู่ไฟล์ต้องเป็นไฟล์ในระบบ (/goapi/...) หรือ http(s) เท่านั้น"}}
	}
	for _, r := range uri {
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			return []productV2FieldError{{field, "INVALID_CHAR", "ที่อยู่ไฟล์มีอักขระที่ใช้ไม่ได้"}}
		}
	}
	return nil
}

// E01 — ชื่อสำหรับลงขาย
func validateListingTitle(title string, limits productV2Limits) []productV2FieldError {
	trimmed := strings.TrimSpace(title)
	length := utf8.RuneCountInString(trimmed)
	switch {
	case length < limits.TitleMin:
		return []productV2FieldError{{"listing.title", "TOO_SHORT", fmt.Sprintf("ชื่อสำหรับลงขายต้องยาวอย่างน้อย %d ตัวอักษร", limits.TitleMin)}}
	case length > limits.TitleMax:
		return []productV2FieldError{{"listing.title", "TOO_LONG", fmt.Sprintf("ชื่อสำหรับลงขายยาว %d ตัวอักษร เกิน %d ตัวอักษร กรุณาย่อชื่อ", length, limits.TitleMax)}}
	}
	for _, r := range trimmed {
		if unicode.IsControl(r) {
			return []productV2FieldError{{"listing.title", "INVALID_CHAR", "ชื่อสำหรับลงขายมีอักขระควบคุมที่ใช้ไม่ได้ กรุณาลบออก"}}
		}
	}
	return nil
}

// E02 — รายละเอียดสำหรับลงขาย อยู่ใน descriptionmin..descriptionmax (ไม่มี limits ของร้าน → 0..2000)
func validateListingDescription(description string, limits productV2Limits) []productV2FieldError {
	length := utf8.RuneCountInString(strings.TrimSpace(description))
	switch {
	case length < limits.ListingDescriptionMin:
		return []productV2FieldError{{"listing.description", "TOO_SHORT", fmt.Sprintf("รายละเอียดสำหรับลงขายต้องยาวอย่างน้อย %d ตัวอักษร", limits.ListingDescriptionMin)}}
	case length > limits.ListingDescriptionMax:
		return []productV2FieldError{{"listing.description", "TOO_LONG", fmt.Sprintf("รายละเอียดสำหรับลงขายยาว %d ตัวอักษร เกิน %d ตัวอักษร", length, limits.ListingDescriptionMax)}}
	}
	return nil
}

// E02 — product.description ยังคง max 1500 ตาม product.go
func validateProductDescription(description string, limits productV2Limits) []productV2FieldError {
	if length := utf8.RuneCountInString(description); length > limits.ProductDescriptionMax {
		return []productV2FieldError{{"description", "TOO_LONG", fmt.Sprintf("รายละเอียดสินค้ายาว %d ตัวอักษร เกิน %d ตัวอักษร", length, limits.ProductDescriptionMax)}}
	}
	return nil
}

// E05 — น้ำหนัก/ขนาดกล่อง ต้องครบ 4 ช่อง หรือว่างทั้งหมด, > 0 และไม่เกินเพดานที่สมเหตุสมผล
func validatePackage(pkg productV2PackagePatch, limits productV2Limits) []productV2FieldError {
	values := []struct {
		name  string
		value *float64
		max   float64
	}{{"weight", pkg.Weight, limits.PackageWeightMax}, {"length", pkg.Length, limits.PackageDimensionMax}, {"width", pkg.Width, limits.PackageDimensionMax}, {"height", pkg.Height, limits.PackageDimensionMax}}
	present := 0
	for _, v := range values {
		if v.value != nil {
			present++
		}
	}
	if present == 0 {
		return nil
	}
	if present != len(values) {
		return []productV2FieldError{{"package", "PACKAGE_INCOMPLETE", productV2MsgPackageIncomplete}}
	}
	var errs []productV2FieldError
	for _, v := range values {
		switch {
		case math.IsNaN(*v.value) || math.IsInf(*v.value, 0) || *v.value <= 0:
			errs = append(errs, productV2FieldError{"package." + v.name, "INVALID_VALUE", "ค่าน้ำหนัก/ขนาดกล่องต้องมากกว่า 0"})
		case *v.value > v.max:
			errs = append(errs, productV2FieldError{"package." + v.name, "OUT_OF_RANGE", fmt.Sprintf("ค่าน้ำหนัก/ขนาดกล่องต้องไม่เกิน %.0f", v.max)})
		}
	}
	return errs
}

// productV2NormalPrice — ราคาปกติที่ใช้เทียบราคาขายส่ง = ราคาต่ำสุดในบรรดา model (0 = ไม่มีราคา ไม่เทียบ)
func productV2NormalPrice(modelPrices []float64) float64 {
	normal := 0.0
	for _, price := range modelPrices {
		if price > 0 && (normal == 0 || price < normal) {
			normal = price
		}
	}
	return normal
}

func decimal128ToRat(value primitive.Decimal128) (*big.Rat, bool) {
	rat, ok := new(big.Rat).SetString(value.String())
	return rat, ok
}

// E07 — ราคาขายส่ง: ช่วงไม่ทับซ้อน, mincount<maxcount, unitprice ≥ 0 และ ≤ ราคาปกติ
// (ยังไม่บังคับ wholesalepctmin เพราะยังไม่มี limits ของร้าน)
func validateWholesale(tiers []productmodels.ProductListingWholesale, normalPrice float64) []productV2FieldError {
	var errs []productV2FieldError
	normalRat := new(big.Rat).SetFloat64(normalPrice)
	ordered := make([]int, len(tiers))
	for i := range tiers {
		ordered[i] = i
	}
	sort.SliceStable(ordered, func(a, b int) bool { return tiers[ordered[a]].MinCount < tiers[ordered[b]].MinCount })

	prevMax := 0
	for pos, idx := range ordered {
		tier := tiers[idx]
		field := fmt.Sprintf("listing.wholesale[%d]", idx)
		if tier.MinCount < 1 || tier.MinCount >= tier.MaxCount {
			errs = append(errs, productV2FieldError{field, "INVALID_VALUE", "จำนวนขั้นต่ำต้องอย่างน้อย 1 และน้อยกว่าจำนวนสูงสุด"})
		}
		if pos > 0 && tier.MinCount <= prevMax {
			errs = append(errs, productV2FieldError{field, "RANGE_OVERLAP", "ช่วงจำนวนของราคาขายส่งทับซ้อนกับช่วงอื่น"})
		}
		if tier.MaxCount > prevMax {
			prevMax = tier.MaxCount
		}
		price, ok := decimal128ToRat(tier.UnitPrice.Decimal128)
		switch {
		case !ok:
			errs = append(errs, productV2FieldError{field + ".unitprice", "INVALID_VALUE", "ราคาต้องเป็นตัวเลขทศนิยม เช่น 199.00 หรือ \"199.00\""})
		case price.Sign() < 0:
			errs = append(errs, productV2FieldError{field + ".unitprice", "INVALID_VALUE", "ราคาขายส่งต้องไม่ติดลบ"})
		case normalPrice > 0 && price.Cmp(normalRat) > 0:
			errs = append(errs, productV2FieldError{field + ".unitprice", "OUT_OF_RANGE", fmt.Sprintf("ราคาขายส่งต้องไม่เกินราคาปกติ %.2f", normalPrice)})
		}
	}
	return errs
}

// E08 — จำนวนซื้อขั้นต่ำ/สูงสุด และวันเตรียมส่งของสินค้าสั่งจอง
// ช่วง daystoshipmin..max มาจาก limits ของหมวดที่จับคู่ ซึ่งยังไม่มีในเฟสนี้ → บังคับแค่ daystoship ≥ 1 เมื่อสั่งจอง
func validatePurchaseLimitAndPreorder(limit *productmodels.ProductListingPurchaseLimit, preorder *productmodels.ProductListingPreorder) []productV2FieldError {
	var errs []productV2FieldError
	if limit != nil {
		if limit.Min < 0 || limit.Max < 0 {
			errs = append(errs, productV2FieldError{"listing.purchaselimit", "INVALID_VALUE", "จำนวนซื้อต้องไม่ติดลบ"})
		} else if limit.Max > 0 && limit.Min > limit.Max {
			errs = append(errs, productV2FieldError{"listing.purchaselimit", "OUT_OF_RANGE", "จำนวนซื้อขั้นต่ำต้องไม่เกินจำนวนซื้อสูงสุด"})
		}
	}
	if preorder != nil && preorder.IsPreorder && preorder.DaysToShip < 1 {
		errs = append(errs, productV2FieldError{"listing.preorder.daystoship", "INVALID_VALUE", "สินค้าสั่งจองต้องระบุวันเตรียมส่งอย่างน้อย 1 วัน"})
	}
	return errs
}

// E20 — สภาพสินค้า
func validateCondition(condition string) []productV2FieldError {
	switch condition {
	case "", productmodels.ListingConditionNew, productmodels.ListingConditionUsed:
		return nil
	}
	return []productV2FieldError{{"listing.condition", "INVALID_VALUE", "สภาพสินค้าต้องเป็น NEW (ใหม่) หรือ USED (มือสอง)"}}
}

// E14 — รูปสินค้า: จำนวนไม่เกินที่กำหนด, xorder ไม่ซ้ำ, ทุกรูปต้องมี uri + urithumb (รูปย่อ) ที่ถูกต้อง
func validateImages(images []productmodels.ProductImage, limits productV2Limits) []productV2FieldError {
	var errs []productV2FieldError
	if len(images) > limits.ImageCountMax {
		errs = append(errs, productV2FieldError{"images", "TOO_LONG", fmt.Sprintf("รูปสินค้าใส่ได้ไม่เกิน %d รูป", limits.ImageCountMax)})
	}
	seen := map[int]struct{}{}
	for i, image := range images {
		if _, dup := seen[image.XOrder]; dup {
			errs = append(errs, productV2FieldError{fmt.Sprintf("images[%d].xorder", i), "DUPLICATE", "ลำดับรูปซ้ำกัน"})
		}
		seen[image.XOrder] = struct{}{}
		errs = append(errs, validateURI(fmt.Sprintf("images[%d].uri", i), image.URI, true, limits)...)
		if strings.TrimSpace(image.URIThumb) == "" {
			errs = append(errs, productV2FieldError{fmt.Sprintf("images[%d].urithumb", i), "REQUIRED", "ต้องมีรูปย่อคู่กับรูปนี้"})
		} else {
			errs = append(errs, validateURI(fmt.Sprintf("images[%d].urithumb", i), image.URIThumb, true, limits)...)
		}
	}
	return errs
}

// วิดีโอสินค้า: จำนวนไม่เกินที่กำหนด, xorder ไม่ซ้ำ, uri ต้องมีและถูกต้อง
func validateVideos(videos []productmodels.ProductVideo, limits productV2Limits) []productV2FieldError {
	var errs []productV2FieldError
	if len(videos) > limits.VideoCountMax {
		errs = append(errs, productV2FieldError{"videos", "TOO_LONG", fmt.Sprintf("วิดีโอสินค้าใส่ได้ไม่เกิน %d รายการ", limits.VideoCountMax)})
	}
	seen := map[int]struct{}{}
	for i, video := range videos {
		if _, dup := seen[video.XOrder]; dup {
			errs = append(errs, productV2FieldError{fmt.Sprintf("videos[%d].xorder", i), "DUPLICATE", "ลำดับวิดีโอซ้ำกัน"})
		}
		seen[video.XOrder] = struct{}{}
		errs = append(errs, validateURI(fmt.Sprintf("videos[%d].uri", i), video.URI, true, limits)...)
		errs = append(errs, validateURI(fmt.Sprintf("videos[%d].posteruri", i), video.PosterURI, false, limits)...)
		if video.DurationSec < 0 || video.SizeBytes < 0 {
			errs = append(errs, productV2FieldError{fmt.Sprintf("videos[%d]", i), "INVALID_VALUE", "ความยาว/ขนาดวิดีโอต้องไม่ติดลบ"})
		}
	}
	return errs
}
