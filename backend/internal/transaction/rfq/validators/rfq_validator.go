package validators

import (
	"fmt"
	"math"
	transmodels "smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/rfq/models"
)

var validVatTypes = map[int8]bool{0: true, 1: true, 2: true, 3: true}

// ValidateRFQ — ตรวจสอบข้อมูลใบสืบราคาก่อนบันทึก
func ValidateRFQ(rfq *models.RFQ) *ValidationResult {
	result := NewValidationResult()
	trans := &rfq.Transaction

	// ตรวจสอบหัวเอกสาร RFQ
	validateRFQHeader(rfq, trans, result)

	// ตรวจสอบรายการสินค้า
	validateDetails(trans, result)

	// ตรวจสอบ Multi-Currency
	validateMultiCurrency(trans, result)

	return result
}

// validateRFQHeader — ตรวจสอบ fields เฉพาะ RFQ
func validateRFQHeader(rfq *models.RFQ, trans *transmodels.Transaction, result *ValidationResult) {
	// docdatelocal — วันที่เอกสาร (ใช้สร้าง DocNo)
	if trans.TransactionHeader.DocDateLocal == "" {
		result.AddError("docdatelocal", "REQUIRED",
			"กรุณาระบุวันที่เอกสาร เนื่องจากระบบใช้สร้างเลขที่เอกสาร (เช่น RFQ20260314xxxxx)",
			"Document date (docdatelocal) is required. It is used to generate the document number (e.g. RFQ20260314xxxxx).")
	}

	// details — รายการสินค้าอย่างน้อย 1 รายการ
	if trans.Details == nil || len(*trans.Details) == 0 {
		result.AddError("details", "EMPTY_DETAILS",
			"กรุณาเพิ่มรายการสินค้าอย่างน้อย 1 รายการ",
			"At least one item is required in the details array.")
	}

	// vendorentries — ต้องมี vendor อย่างน้อย 1 ราย
	if len(rfq.VendorEntries) == 0 {
		result.AddError("vendorentries", "EMPTY_VENDOR_ENTRIES",
			"กรุณาเพิ่มผู้เสนอราคาอย่างน้อย 1 ราย",
			"At least one vendor entry is required.")
	}

	// vattype
	if !validVatTypes[trans.TransactionHeader.VatType] {
		result.AddErrorf("vattype", "INVALID_VALUE",
			"ประเภทภาษี (vattype) ไม่ถูกต้อง: %d — ค่าที่รองรับ: 0, 1, 2, 3",
			"Invalid VAT type: %d — allowed values: 0, 1, 2, 3.",
			trans.TransactionHeader.VatType, trans.TransactionHeader.VatType)
	}
}

// validateDetails — ตรวจสอบรายการสินค้า
func validateDetails(trans *transmodels.Transaction, result *ValidationResult) {
	if trans.Details == nil {
		return
	}
	for i, detail := range *trans.Details {
		itemLabel := detail.ItemCode
		if itemLabel == "" {
			itemLabel = detail.Barcode
		}

		if detail.ItemCode == "" && detail.Barcode == "" {
			result.AddErrorf("details[%d].itemcode", "REQUIRED",
				"รายการที่ %d: กรุณาระบุรหัสสินค้า (itemcode) หรือบาร์โค้ด (barcode)",
				"Item #%d: Item code (itemcode) or barcode is required.",
				i, i+1, i+1)
		}

		if detail.Qty <= 0 {
			result.AddErrorf("details[%d].qty", "MUST_BE_POSITIVE",
				"รายการที่ %d (%s): จำนวนต้องมากกว่า 0 — ค่าปัจจุบัน: %.4f",
				"Item #%d (%s): Quantity must be greater than 0 — current value: %.4f.",
				i, i+1, itemLabel, detail.Qty, i+1, itemLabel, detail.Qty)
		}

		checkDetailNaN(i, itemLabel, &detail, result)
	}
}

func checkDetailNaN(index int, itemLabel string, detail *transmodels.Detail, result *ValidationResult) {
	type fieldCheck struct {
		name  string
		value float64
	}
	fields := []fieldCheck{
		{"qty", detail.Qty},
		{"price", detail.Price},
		{"sumamount", detail.SumAmount},
	}
	for _, f := range fields {
		if math.IsNaN(f.value) || math.IsInf(f.value, 0) {
			result.AddErrorf(fmt.Sprintf("details[%d].%s", index, f.name), "INVALID_NUMBER",
				"รายการที่ %d (%s): ฟิลด์ %s มีค่าที่ไม่ถูกต้อง (NaN หรือ Infinity)",
				"Item #%d (%s): Field %s contains an invalid number (NaN or Infinity).",
				index, index+1, itemLabel, f.name, index+1, itemLabel, f.name)
		}
	}
}

func validateMultiCurrency(trans *transmodels.Transaction, result *ValidationResult) {
	header := &trans.TransactionHeader
	if header.DocCurrency != "" {
		if header.ExchangeRate <= 0 {
			result.AddErrorf("exchangerate", "INVALID_EXCHANGE_RATE",
				"อัตราแลกเปลี่ยนต้องมากกว่า 0 เมื่อระบุสกุลเงินเอกสาร (%s) — ค่าปัจจุบัน: %.6f",
				"Exchange rate must be greater than 0 when document currency (%s) is specified — current value: %.6f.",
				header.DocCurrency, header.ExchangeRate, header.DocCurrency, header.ExchangeRate)
		}
	}

	type headerFieldCheck struct {
		name  string
		value float64
	}
	headerFields := []headerFieldCheck{
		{"exchangerate", header.ExchangeRate},
		{"totalvalue", header.TotalValue},
		{"totalaftervat", header.TotalAfterVat},
		{"totalamount", header.TotalAmount},
	}
	for _, f := range headerFields {
		if math.IsNaN(f.value) || math.IsInf(f.value, 0) {
			result.AddErrorf(f.name, "INVALID_NUMBER",
				"ฟิลด์ %s มีค่าที่ไม่ถูกต้อง (NaN หรือ Infinity)",
				"Field %s contains an invalid number (NaN or Infinity).",
				f.name, f.name)
		}
	}
}
