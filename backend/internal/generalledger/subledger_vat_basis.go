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

// ลำดับแถวของทะเบียนภาษีซื้อ/ขาย (Champ 5536/5537 "เรียงตาม" 4 ตัวเลือก: GLRepInputTaxCHMDlg.cpp:119-122,421-451,
// GLRepOutputTaxDlg.cpp:103-106,386-426 — ORDER BY ของทั้งสองรายงานเหมือนกัน); ว่าง = วันที่ใบกำกับแบบเดิมของระบบ
const (
	// VatSortInvoiceDate - วันที่ใบกำกับ แล้วเลขที่ใบสำคัญ (ลำดับเดิมของทะเบียน/ภ.พ.30 คงไว้ไม่เปลี่ยน)
	VatSortInvoiceDate = "invoice_date"
	// VatSortDocumentDate - Champ DocDate, TaxDate, DocNo, TaxNo
	VatSortDocumentDate = "document_date"
	// VatSortInvoiceNo - Champ TaxNo, TaxDate, DocDate, DocNo
	VatSortInvoiceNo = "invoice_no"
	// VatSortDocumentNo - Champ DocNo, DocDate, TaxDate, TaxNo
	VatSortDocumentNo = "document_no"
)

// vatSortOrders - ORDER BY คงที่ต่อคีย์ (whitelist) ของ SELECT สุดท้ายใน VatRegisterRecords — ห้ามต่อข้อความจากผู้ใช้เข้า SQL;
// ทุกแบบปิดท้ายด้วยลำดับรายการในใบ (p.ord) ให้ผลคงที่ทุกครั้ง
var vatSortOrders = map[string]string{
	"":                  `p.item->>'tax_invoice_date', p.doc_no, p.ord`,
	VatSortInvoiceDate:  `p.item->>'tax_invoice_date', p.doc_no, p.ord`,
	VatSortDocumentDate: `p.doc_date, p.item->>'tax_invoice_date', p.doc_no, p.item->>'tax_invoice_no', p.ord`,
	VatSortInvoiceNo:    `p.item->>'tax_invoice_no', p.item->>'tax_invoice_date', p.doc_date, p.doc_no, p.ord`,
	VatSortDocumentNo:   `p.doc_no, p.doc_date, p.item->>'tax_invoice_date', p.item->>'tax_invoice_no', p.ord`,
}

// ValidVatSort - ว่างถือเป็นค่าเริ่มต้น (VatSortInvoiceDate)
func ValidVatSort(sort string) bool {
	_, ok := vatSortOrders[sort]
	return ok
}

// ตัวเลือก "แสดงยอดขอคืนภาษี" ของ Champ 5536 (GLRepInputTaxCHMDlg.cpp:136-140,370-381: ขอคืน = ReturnTax 0, ไม่ขอคืน = ReturnTax 1,
// ทั้งหมด = ไม่กรอง) เทียบ claim_status ของ BC: ขอคืน = 1 ใช้สิทธิ; ไม่ขอคืน = 2 ต้องห้าม / 3 รอใช้สิทธิ / 4 ไม่ใช้สิทธิ — ภาษีซื้อเท่านั้น
const (
	// VatClaimClaimed - ใช้สิทธิ (claim_status 1) = ชุดเดียวกับ ภ.พ.30; ค่าเริ่มต้น
	VatClaimClaimed = "claimed"
	// VatClaimNotClaimed - claim_status 2/3/4 (ไม่อยู่ใน ภ.พ.30 ของงวด)
	VatClaimNotClaimed = "not_claimed"
	// VatClaimAll - ทุกสถานะ
	VatClaimAll = "all"
)

// ValidVatClaim - ว่างถือเป็นค่าเริ่มต้น (VatClaimClaimed)
func ValidVatClaim(claim string) bool {
	switch claim {
	case "", VatClaimClaimed, VatClaimNotClaimed, VatClaimAll:
		return true
	}
	return false
}

// VatRegisterQuery - ตัวเลือกของทะเบียนภาษีซื้อ/ขาย; ค่าว่างทุกช่อง = ชุดแถวและลำดับเดียวกับ ภ.พ.30 (VatRecordsForPeriod)
type VatRegisterQuery struct {
	Basis string // VatBasis*
	Sort  string // VatSort*
	Claim string // VatClaim* — ใช้กับภาษีซื้อเท่านั้น; ภาษีขายต้องว่าง
}
