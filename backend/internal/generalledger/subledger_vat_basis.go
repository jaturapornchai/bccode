package generalledger

// ฐานวันที่ของทะเบียนภาษีซื้อ/ขาย (Champ 5536 "พิมพ์ตาม": GLRepInputTaxCHMDlg.cpp:108-110,266-280 เลือกคอลัมน์วันที่ที่ใช้กรองช่วงวันที่
// TaxDate = วันที่ใบกำกับ, DocDate = วันที่เอกสาร, TaxDate2 = วันที่ยื่นภาษี; 5537 เลือกได้เฉพาะ TaxDate/DocDate ตาม GLRepOutputTaxDlg.cpp:186-221)
const (
	// VatBasisPeriod - งวดภาษีที่บันทึกในแถว (tax_period_year/month = เดือนที่ยื่น ภ.พ.30 / งวดที่ใช้สิทธิภาษีซื้อ) เทียบ TaxDate2 ของ Champ; ค่าเริ่มต้น
	VatBasisPeriod = "period"
	// VatBasisInvoiceDate - เดือนของ tax_invoice_date (วันที่ใบกำกับ) เทียบ TaxDate ของ Champ
	VatBasisInvoiceDate = "invoice_date"
	// VatBasisDocumentDate - เดือนของวันที่ใบสำคัญ (วันที่เอกสาร) เทียบ DocDate ของ Champ
	VatBasisDocumentDate = "document_date"
)

// ValidVatBasis - ว่างถือเป็นค่าเริ่มต้น (VatBasisPeriod)
func ValidVatBasis(basis string) bool {
	switch basis {
	case "", VatBasisPeriod, VatBasisInvoiceDate, VatBasisDocumentDate:
		return true
	}
	return false
}
