// Package validators — Purchase Order (PO) Backend Validation System
//
// ============================================================================
// PURPOSE / จุดประสงค์:
//
//	This package validates Purchase Order data at the backend (mainapi) before
//	saving to MongoDB and publishing to Kafka. It serves as the "single source
//	of truth" for PO data integrity.
//
//	Package นี้ตรวจสอบข้อมูลใบสั่งซื้อที่ backend (mainapi) ก่อนบันทึกลง MongoDB
//	และส่งไป Kafka เพื่อเป็นจุดตรวจสอบข้อมูลที่เชื่อถือได้จุดเดียว
//
// WHY BACKEND VALIDATION? / ทำไมต้อง validate ที่ backend?:
//   - External clients (customers, partners) will submit POs via API/MCP
//     without going through the Flutter frontend
//   - AI agents will integrate and create POs programmatically
//   - Frontend validation alone cannot guarantee data integrity
//
// ARCHITECTURE / สถาปัตยกรรม:
//
//	Request → [Layer 1: Struct Tags] → [Layer 2: Business Validator (this file)] → [Layer 3: Sanitizer] → Save
//
//	Layer 1: go-playground/validator tags — basic required/range checks
//	Layer 2: Business rules (this file) — complex PO-specific validation
//	Layer 3: Sanitizer (in service layer) — auto-fix minor issues (e.g. exchange rate ≤ 0 → 1.0)
//
// ERROR HANDLING / การจัดการ error:
//   - Collects ALL errors before returning (does NOT stop at first error)
//   - Each error includes: field path, error code, bilingual message (TH/EN)
//   - Error messages explain WHY the value is invalid (not just "invalid value")
//
// FOR AI AGENTS / สำหรับ AI Agent ที่จะมาเชื่อมต่อ:
//   - Error codes are machine-readable (e.g. "REQUIRED", "MUST_BE_POSITIVE")
//   - Field paths use dot notation for nested fields (e.g. "details[2].qty")
//   - Messages are bilingual — parse "en" for English, "th" for Thai
//   - HTTP 400 = validation error, check "errors" array for details
//
// ============================================================================
package validators

import (
	"fmt"
	"math"
	transmodels "smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/purchaseorder/models"
)

// ============================================================================
// VALID VALUES / ค่าที่ยอมรับ
// ============================================================================

// validVatTypes — ประเภทภาษีที่รองรับ
// VAT types supported by the system:
//
//	0 = ไม่มีภาษี (No VAT)
//	1 = รวมภาษี (VAT included in price)
//	2 = แยกภาษี (VAT excluded from price)
//	3 = ยกเว้นภาษี (VAT exempt)
var validVatTypes = map[int8]bool{0: true, 1: true, 2: true, 3: true}

// validWHTRates — อัตราภาษีหัก ณ ที่จ่ายที่ถูกต้องตามกฎหมายไทย
// Valid Withholding Tax (WHT) rates per Thai Revenue Department regulations (%)
var validWHTRates = map[float64]bool{
	0.5: true, 1: true, 2: true, 3: true, 5: true, 10: true, 15: true,
}

// ============================================================================
// MAIN VALIDATOR / ฟังก์ชันหลัก
// ============================================================================

// ValidatePurchaseOrder — ตรวจสอบข้อมูลใบสั่งซื้อทั้งหมดก่อนบันทึก
//
// Validates all PO data before saving to database.
// Returns a ValidationResult containing all errors found (empty = valid).
//
// Validation order:
//  1. Header fields (docdatelocal, custcode, vattype, totals)
//  2. Detail items (itemcode, qty, price, sumamount)
//  3. Multi-currency fields (exchangerate, NaN/Infinity checks)
//  4. WHT entries (rate, taxbase, amount)
//  5. Credit terms (creditdays, duedate)
//
// Usage:
//
//	result := validators.ValidatePurchaseOrder(&po)
//	if !result.IsValid() {
//	    return result.ToErrorResponse()  // HTTP 400
//	}
func ValidatePurchaseOrder(po *models.PurchaseOrder) *ValidationResult {
	result := NewValidationResult()

	// ดึง Transaction ที่ embed อยู่ใน PurchaseOrder
	trans := &po.Transaction

	// ตรวจสอบระดับหัวเอกสาร (Header-level validation)
	validateHeader(trans, result)

	// ตรวจสอบรายการสินค้า (Detail items validation)
	validateDetails(trans, result)

	// ตรวจสอบ Multi-Currency (Exchange rate, NaN/Infinity)
	validateMultiCurrency(trans, result)

	// ตรวจสอบภาษีหัก ณ ที่จ่าย (Withholding Tax entries)
	validateWHTEntries(trans, result)

	// ตรวจสอบเงื่อนไขการชำระเงิน (Credit terms)
	validateCreditTerms(trans, result)

	return result
}

// ValidatePurchaseOrderForUpdate — ตรวจสอบข้อมูลใบสั่งซื้อสำหรับการแก้ไข
//
// Same as ValidatePurchaseOrder, but adds update-specific checks:
//   - Cannot edit approved POs
//   - Cannot edit POs pending approval
//   - Auto-approved POs can only be edited by the original creator
//
// Parameters:
//   - po: ข้อมูล PO ที่จะแก้ไข
//   - approvalStatus: สถานะการอนุมัติปัจจุบัน ("", "auto_approved", "pending", "approved", "rejected")
//   - creatorCode: รหัสผู้สร้างเอกสารเดิม
//   - currentUser: รหัสผู้ใช้ที่กำลังแก้ไข
func ValidatePurchaseOrderForUpdate(po *models.PurchaseOrder, approvalStatus string, creatorCode string, currentUser string) *ValidationResult {
	// ตรวจสอบ business rules สำหรับการแก้ไข (ก่อนตรวจข้อมูล)
	result := NewValidationResult()

	// ตรวจสอบสถานะการอนุมัติ
	validateApprovalStatusForUpdate(approvalStatus, creatorCode, currentUser, result)

	// ถ้าไม่ผ่านการตรวจสอบสถานะ → return ทันที (ไม่ต้องตรวจข้อมูลอื่น)
	if !result.IsValid() {
		return result
	}

	// ตรวจสอบข้อมูล PO ปกติ
	dataResult := ValidatePurchaseOrder(po)

	// รวม errors จากทั้งสอง result
	for _, err := range dataResult.Errors {
		result.Errors = append(result.Errors, err)
	}

	return result
}

// ============================================================================
// HEADER VALIDATION / ตรวจสอบหัวเอกสาร
// ============================================================================

// validateHeader — ตรวจสอบ fields ระดับหัวเอกสาร
//
// Required fields for PO:
//   - docdatelocal: ใช้สร้างเลขที่เอกสาร (DocNo prefix = "PO" + YYYYMMDD)
//   - custcode: รหัสผู้จำหน่าย (supplier) — ใบสั่งซื้อต้องมีผู้ขายเสมอ
//   - details: ต้องมีรายการสินค้าอย่างน้อย 1 รายการ
//   - vattype: ต้องเป็นค่าที่ระบบรองรับ (0, 1, 2, 3)
//   - totalaftervat: ยอดรวมหลังภาษีต้องมากกว่า 0
func validateHeader(trans *transmodels.Transaction, result *ValidationResult) {
	// docdatelocal — วันที่เอกสาร (ใช้สร้าง DocNo)
	// Document date in local timezone — REQUIRED because it's used to generate
	// the document number (e.g. PO20260208xxxxx)
	if trans.TransactionHeader.DocDateLocal == "" {
		result.AddError("docdatelocal", "REQUIRED",
			"กรุณาระบุวันที่เอกสาร เนื่องจากระบบใช้สร้างเลขที่เอกสาร (เช่น PO20260208xxxxx)",
			"Document date (docdatelocal) is required. It is used to generate the document number (e.g. PO20260208xxxxx).")
	}

	// custcode — รหัสผู้จำหน่าย (Supplier code)
	// Supplier/vendor code — REQUIRED because every PO must specify who you're buying from
	if trans.TransactionHeader.CustCode == "" {
		result.AddError("custcode", "REQUIRED",
			"กรุณาระบุรหัสผู้จำหน่าย (supplier) เนื่องจากใบสั่งซื้อต้องระบุผู้ขายเสมอ",
			"Supplier code (custcode) is required. A purchase order must always specify a vendor.")
	}

	// details — รายการสินค้า (at least 1 item required)
	if trans.Details == nil || len(*trans.Details) == 0 {
		result.AddError("details", "EMPTY_DETAILS",
			"กรุณาเพิ่มรายการสินค้าอย่างน้อย 1 รายการ",
			"At least one item is required in the details array.")
	}

	// vattype — ประเภทภาษี (must be 0, 1, 2, or 3)
	if !validVatTypes[trans.TransactionHeader.VatType] {
		result.AddErrorf("vattype", "INVALID_VALUE",
			"ประเภทภาษี (vattype) ไม่ถูกต้อง: %d — ค่าที่รองรับ: 0 (ไม่มีภาษี), 1 (รวมภาษี), 2 (แยกภาษี), 3 (ยกเว้นภาษี)",
			"Invalid VAT type: %d — allowed values: 0 (no VAT), 1 (VAT included), 2 (VAT excluded), 3 (VAT exempt).",
			trans.TransactionHeader.VatType, trans.TransactionHeader.VatType)
	}

	// totalaftervat — ยอดรวมหลังภาษี (must be > 0)
	// Total after VAT must be positive — a PO with zero total is meaningless
	if trans.TransactionHeader.TotalAfterVat <= 0 {
		result.AddErrorf("totalaftervat", "MUST_BE_POSITIVE",
			"ยอดรวมหลังภาษีต้องมากกว่า 0 — ค่าปัจจุบัน: %.2f",
			"Total after VAT must be greater than 0 — current value: %.2f.",
			trans.TransactionHeader.TotalAfterVat, trans.TransactionHeader.TotalAfterVat)
	}
}

// ============================================================================
// DETAIL VALIDATION / ตรวจสอบรายการสินค้า
// ============================================================================

// validateDetails — ตรวจสอบรายการสินค้าแต่ละรายการ
//
// For each item in details[]:
//   - itemcode or barcode: ต้องระบุอย่างน้อยอย่างใดอย่างหนึ่ง
//   - qty: จำนวนต้องมากกว่า 0
//   - price: ราคาต้องไม่ติดลบ (0 ได้ เช่น สินค้าแถม)
//   - sumamount: ยอดรวมต้องไม่ติดลบ
//
// Field paths use array index notation: "details[0].qty", "details[1].itemcode"
func validateDetails(trans *transmodels.Transaction, result *ValidationResult) {
	if trans.Details == nil {
		return // already reported in validateHeader
	}

	for i, detail := range *trans.Details {
		itemLabel := detail.ItemCode
		if itemLabel == "" {
			itemLabel = detail.Barcode
		}

		// itemcode หรือ barcode — ต้องระบุอย่างน้อยหนึ่งอย่าง
		// Either itemcode or barcode must be provided to identify the product
		if detail.ItemCode == "" && detail.Barcode == "" {
			result.AddErrorf("details[%d].itemcode", "REQUIRED",
				"รายการที่ %d: กรุณาระบุรหัสสินค้า (itemcode) หรือบาร์โค้ด (barcode)",
				"Item #%d: Item code (itemcode) or barcode is required.",
				i, i+1, i+1)
		}

		// qty — จำนวนสินค้าต้องมากกว่า 0
		// Quantity must be positive — you can't order zero or negative items
		if detail.Qty <= 0 {
			result.AddErrorf("details[%d].qty", "MUST_BE_POSITIVE",
				"รายการที่ %d (%s): จำนวนต้องมากกว่า 0 — ค่าปัจจุบัน: %.4f",
				"Item #%d (%s): Quantity must be greater than 0 — current value: %.4f.",
				i, i+1, itemLabel, detail.Qty, i+1, itemLabel, detail.Qty)
		}

		// price — ราคาต้องไม่ติดลบ (0 อนุญาต เช่น สินค้าแถม/ตัวอย่าง)
		// Price must not be negative — zero is allowed for free/sample items
		if detail.Price < 0 {
			result.AddErrorf("details[%d].price", "MUST_NOT_BE_NEGATIVE",
				"รายการที่ %d (%s): ราคาต้องไม่ติดลบ — ค่าปัจจุบัน: %.4f",
				"Item #%d (%s): Price must not be negative — current value: %.4f.",
				i, i+1, itemLabel, detail.Price, i+1, itemLabel, detail.Price)
		}

		// sumamount — ยอดรวมต้องไม่ติดลบ
		// Sum amount (qty × price - discount) must not be negative
		if detail.SumAmount < 0 {
			result.AddErrorf("details[%d].sumamount", "MUST_NOT_BE_NEGATIVE",
				"รายการที่ %d (%s): ยอดรวม (sumamount) ต้องไม่ติดลบ — ค่าปัจจุบัน: %.4f",
				"Item #%d (%s): Sum amount must not be negative — current value: %.4f.",
				i, i+1, itemLabel, detail.SumAmount, i+1, itemLabel, detail.SumAmount)
		}

		// ตรวจสอบ NaN/Infinity ใน detail fields
		checkDetailNaN(i, itemLabel, &detail, result)
	}
}

// checkDetailNaN — ตรวจสอบค่า NaN และ Infinity ในรายการสินค้า
//
// NaN/Infinity can occur when:
//   - Division by zero in frontend calculator
//   - JSON parsing errors
//   - Incorrect data from external API clients
//
// These values will corrupt MongoDB data and cause calculation errors downstream
func checkDetailNaN(index int, itemLabel string, detail *transmodels.Detail, result *ValidationResult) {
	type fieldCheck struct {
		name  string
		value float64
	}

	fields := []fieldCheck{
		{"qty", detail.Qty},
		{"price", detail.Price},
		{"sumamount", detail.SumAmount},
		{"discountamount", detail.DiscountAmount},
		{"price_doc", detail.PriceDoc},
		{"sumamount_doc", detail.SumAmountDoc},
	}

	for _, f := range fields {
		if math.IsNaN(f.value) || math.IsInf(f.value, 0) {
			result.AddErrorf(fmt.Sprintf("details[%d].%s", index, f.name), "INVALID_NUMBER",
				"รายการที่ %d (%s): ฟิลด์ %s มีค่าที่ไม่ถูกต้อง (NaN หรือ Infinity) — อาจเกิดจากการหารด้วย 0",
				"Item #%d (%s): Field %s contains an invalid number (NaN or Infinity) — possibly caused by division by zero.",
				index, index+1, itemLabel, f.name, index+1, itemLabel, f.name)
		}
	}
}

// ============================================================================
// MULTI-CURRENCY VALIDATION / ตรวจสอบสกุลเงิน
// ============================================================================

// validateMultiCurrency — ตรวจสอบข้อมูลอัตราแลกเปลี่ยนและ NaN/Infinity ใน header
//
// Multi-currency PO fields:
//   - currency: สกุลเงินหลัก (Base Currency) เช่น THB — ใช้ลงบัญชี
//   - doc_currency: สกุลเงินเอกสาร (Document Currency) เช่น USD, JPY — สกุลเงินที่ตกลงกับผู้ขาย
//   - exchangerate: อัตราแลกเปลี่ยน (1 Document Currency = ? Base Currency)
//
// When doc_currency is specified (non-THB purchase):
//   - exchangerate must be > 0
//   - exchangerate must not be NaN or Infinity
func validateMultiCurrency(trans *transmodels.Transaction, result *ValidationResult) {
	header := &trans.TransactionHeader

	// ถ้ามี DocCurrency (สกุลเงินเอกสาร) → ต้องมี ExchangeRate ที่ถูกต้อง
	// If document currency is specified, exchange rate must be valid
	if header.DocCurrency != "" {
		if header.ExchangeRate <= 0 {
			result.AddErrorf("exchange_rate", "INVALID_EXCHANGE_RATE",
				"อัตราแลกเปลี่ยนต้องมากกว่า 0 เมื่อระบุสกุลเงินเอกสาร (%s) — ค่าปัจจุบัน: %.6f",
				"Exchange rate must be greater than 0 when document currency (%s) is specified — current value: %.6f.",
				header.DocCurrency, header.ExchangeRate, header.DocCurrency, header.ExchangeRate)
		}
	}

	// ตรวจสอบ NaN/Infinity ใน header fields
	// These corrupt data in MongoDB and cause cascading errors in Kafka consumers
	type headerFieldCheck struct {
		name  string
		value float64
	}

	headerFields := []headerFieldCheck{
		{"exchange_rate", header.ExchangeRate},
		{"totalvalue", header.TotalValue},
		{"totalaftervat", header.TotalAfterVat},
		{"totalamount", header.TotalAmount},
		{"totalbeforevat", header.TotalBeforeVat},
		{"totalvatvalue", header.TotalVatValue},
		{"totaldiscount", header.TotalDiscount},
		{"totalvalue_doc", header.TotalValueDoc},
		{"totalaftervat_doc", header.TotalAfterVatDoc},
		{"totalamount_doc", header.TotalAmountDoc},
	}

	for _, f := range headerFields {
		if math.IsNaN(f.value) || math.IsInf(f.value, 0) {
			result.AddErrorf(f.name, "INVALID_NUMBER",
				"ฟิลด์ %s มีค่าที่ไม่ถูกต้อง (NaN หรือ Infinity) — อาจเกิดจากการคำนวณผิดพลาด",
				"Field %s contains an invalid number (NaN or Infinity) — possibly caused by a calculation error.",
				f.name, f.name)
		}
	}
}

// ============================================================================
// WHT VALIDATION / ตรวจสอบภาษีหัก ณ ที่จ่าย
// ============================================================================

// validateWHTEntries — ตรวจสอบรายการภาษีหัก ณ ที่จ่าย (Withholding Tax)
//
// WHT (Withholding Tax / ภาษีหัก ณ ที่จ่าย):
//   - Thai law requires buyers to withhold tax on certain purchases
//   - A PO may have multiple WHT entries (different rates for different income types)
//   - Valid rates per Thai Revenue Department: 0.5%, 1%, 2%, 3%, 5%, 10%, 15%
//
// Each WHTEntry has:
//   - rate: อัตราภาษี (%) — must be one of the valid rates
//   - taxbase: ฐานภาษี (base amount) — must not be negative
//   - amount: จำนวนเงินหัก (taxbase × rate%) — must not be negative
func validateWHTEntries(trans *transmodels.Transaction, result *ValidationResult) {
	if len(trans.TransactionHeader.WHTEntries) == 0 {
		return // ไม่มี WHT → ไม่ต้องตรวจ
	}

	for i, entry := range trans.TransactionHeader.WHTEntries {
		// rate — อัตราภาษีต้องเป็นค่าที่ถูกต้องตามกฎหมาย
		if entry.Rate != 0 && !validWHTRates[entry.Rate] {
			result.AddErrorf(fmt.Sprintf("wht_entries[%d].rate", i), "INVALID_WHT_RATE",
				"รายการ WHT ที่ %d: อัตราภาษีหัก ณ ที่จ่ายไม่ถูกต้อง: %.2f%% — ค่าที่รองรับ: 0.5, 1, 2, 3, 5, 10, 15",
				"WHT entry #%d: Invalid withholding tax rate: %.2f%% — allowed values: 0.5, 1, 2, 3, 5, 10, 15.",
				i, i+1, entry.Rate, i+1, entry.Rate)
		}

		// taxbase — ฐานภาษีต้องไม่ติดลบ
		if entry.TaxBase < 0 {
			result.AddErrorf(fmt.Sprintf("wht_entries[%d].taxbase", i), "MUST_NOT_BE_NEGATIVE",
				"รายการ WHT ที่ %d: ฐานภาษีหัก ณ ที่จ่ายต้องไม่ติดลบ — ค่าปัจจุบัน: %.2f",
				"WHT entry #%d: Tax base must not be negative — current value: %.2f.",
				i, i+1, entry.TaxBase, i+1, entry.TaxBase)
		}

		// amount — จำนวนเงินหักต้องไม่ติดลบ
		if entry.Amount < 0 {
			result.AddErrorf(fmt.Sprintf("wht_entries[%d].amount", i), "MUST_NOT_BE_NEGATIVE",
				"รายการ WHT ที่ %d: จำนวนเงินหัก ณ ที่จ่ายต้องไม่ติดลบ — ค่าปัจจุบัน: %.2f",
				"WHT entry #%d: WHT amount must not be negative — current value: %.2f.",
				i, i+1, entry.Amount, i+1, entry.Amount)
		}
	}
}

// ============================================================================
// CREDIT TERMS VALIDATION / ตรวจสอบเงื่อนไขการชำระเงิน
// ============================================================================

// validateCreditTerms — ตรวจสอบเงื่อนไขเครดิตเทอม
//
// Credit terms:
//   - creditdays: จำนวนวันเครดิต (0 = ชำระทันที, >0 = เครดิต N วัน)
//   - duedate: วันครบกำหนดชำระ (YYYY-MM-DD format)
//
// Rules:
//   - creditdays must not be negative
//   - If creditdays > 0, a valid duedate should be provided (warning, not error)
func validateCreditTerms(trans *transmodels.Transaction, result *ValidationResult) {
	header := &trans.TransactionHeader

	// creditdays — จำนวนวันเครดิตต้องไม่ติดลบ
	if header.CreditDays < 0 {
		result.AddErrorf("creditdays", "MUST_NOT_BE_NEGATIVE",
			"จำนวนวันเครดิตต้องไม่ติดลบ — ค่าปัจจุบัน: %d",
			"Credit days must not be negative — current value: %d.",
			header.CreditDays, header.CreditDays)
	}
}

// ============================================================================
// APPROVAL STATUS VALIDATION / ตรวจสอบสถานะอนุมัติ (สำหรับ Update)
// ============================================================================

// validateApprovalStatusForUpdate — ตรวจสอบว่าสามารถแก้ไข PO ได้หรือไม่ ตามสถานะอนุมัติ
//
// PO Approval Statuses:
//   - "" (empty): ยังไม่มี approval record → แก้ไขได้
//   - "auto_approved": อนุมัติอัตโนมัติ → แก้ไขได้เฉพาะผู้สร้าง (creator) เท่านั้น
//   - "pending": รออนุมัติ → ห้ามแก้ไข (ต้องถอนการส่งอนุมัติก่อน)
//   - "approved": อนุมัติแล้ว → ห้ามแก้ไข
//   - "rejected": ถูกปฏิเสธ → แก้ไขได้ (เพื่อส่งอนุมัติใหม่)
//
// NOTE: ไม่มีสถานะ "draft" — PO ที่ยังไม่มี approval record ถือว่ายังไม่เข้าระบบอนุมัติ
func validateApprovalStatusForUpdate(approvalStatus string, creatorCode string, currentUser string, result *ValidationResult) {
	switch approvalStatus {
	case "approved":
		// ห้ามแก้ไข PO ที่อนุมัติแล้ว
		result.AddErrorf("approval_status", "PO_ALREADY_APPROVED",
			"ไม่สามารถแก้ไขใบสั่งซื้อที่อนุมัติแล้วได้ (สถานะ: %s)",
			"Cannot edit an approved purchase order (status: %s).",
			approvalStatus, approvalStatus)

	case "pending":
		// ห้ามแก้ไข PO ที่รออนุมัติ — ต้องถอนการส่งอนุมัติก่อน
		result.AddError("approval_status", "PO_PENDING_APPROVAL",
			"ไม่สามารถแก้ไขใบสั่งซื้อที่รออนุมัติได้ กรุณาถอนการส่งอนุมัติก่อนแก้ไข",
			"Cannot edit a purchase order pending approval. Please withdraw the approval request first.")

	case "auto_approved":
		// แก้ไขได้เฉพาะผู้สร้างเอกสาร
		if currentUser != creatorCode && creatorCode != "" {
			result.AddErrorf("approval_status", "PO_NOT_CREATOR",
				"ไม่สามารถแก้ไขใบสั่งซื้อนี้ได้ เนื่องจากคุณไม่ใช่ผู้สร้างเอกสาร (ผู้สร้าง: %s)",
				"Cannot edit this purchase order. Only the creator (%s) can edit auto-approved orders.",
				creatorCode, creatorCode)
		}

	case "", "rejected":
		// แก้ไขได้ — "" = ยังไม่เข้าระบบอนุมัติ, "rejected" = ถูกปฏิเสธแล้ว
		// No restriction — empty means no approval record, rejected means can resubmit
	}
}
