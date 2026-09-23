package generalledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/shopspring/decimal"
)

// SubledgerVat - ภาษีมูลค่าเพิ่มประกอบใบสำคัญ: 1 แถว = 1 รายการใบกำกับภาษี (ไม่ใช่บรรทัดเดบิต/เครดิต)
// ชื่อฟิลด์ตาม mydocs/datamodels/gl/vat.sql; ฐานภาษีเป็นค่าที่ผู้ใช้กรอกและแก้ได้เสมอ (แนวเดียวกับ Champ BaseOfTax)
// ไม่อนุมานจากบรรทัดบัญชี — ภ.พ.30 และรายงานภาษีซื้อ/ขายใช้ค่าที่บันทึกไว้ตรง ๆ
// เก็บยอดบวกเสมอ: ใบลดหนี้ (document_type 3) รายงานเป็นผู้กลับเครื่องหมาย
type SubledgerVat struct {
	ID                  string  `json:"id"`
	TaxType             int     `json:"tax_type"`      // 1=ภาษีซื้อ, 2=ภาษีขาย
	DocumentType        int     `json:"document_type"` // 1=ใบกำกับภาษี, 2=ใบเพิ่มหนี้, 3=ใบลดหนี้
	TaxInvoiceNo        string  `json:"tax_invoice_no"`
	TaxInvoiceDate      string  `json:"tax_invoice_date"`
	OriginalInvoiceNo   string  `json:"original_invoice_no,omitempty"`
	OriginalInvoiceDate string  `json:"original_invoice_date,omitempty"`
	TaxPeriodYear       int     `json:"tax_period_year,omitempty"` // ค.ศ.; 0 คู่กับเดือน 0 = ยังไม่กำหนดงวดใช้สิทธิ
	TaxPeriodMonth      int     `json:"tax_period_month,omitempty"`
	PartnerCode         string  `json:"partner_code,omitempty"` // ไม่บังคับ — ถ้าระบุต้องเป็นคู่ค้าที่เปิดใช้งาน
	PartnerTaxID        string  `json:"partner_tax_id,omitempty"`
	PartnerBranchNo     string  `json:"partner_branch_no,omitempty"` // 00000 = สำนักงานใหญ่
	PartnerName         string  `json:"partner_name"`
	BaseAmount          Amount  `json:"base_amount"`
	ZeroRateAmount      Amount  `json:"zero_rate_amount"`
	ExemptAmount        Amount  `json:"exempt_amount"`
	Rate                Amount  `json:"vat_rate"`
	VatAmount           *Amount `json:"vat_amount,omitempty"`   // null = ให้ระบบคำนวณ ฐาน × อัตรา
	ClaimStatus         int     `json:"claim_status,omitempty"` // เฉพาะซื้อ: 1=ใช้สิทธิ 2=ต้องห้าม 3=รอใช้สิทธิ 4=ไม่ใช้สิทธิ
	ClaimReason         string  `json:"claim_reason,omitempty"`
	Remark              string  `json:"remark,omitempty"`
}

// NUMERIC(16,2) ใน vat.sql — ยอดต้องน้อยกว่า 10^14
var vatAmountLimit = decimal.New(1, 14)

// vat - ตรวจรายการภาษีมูลค่าเพิ่มตาม CHECK ของ vat.sql; ภาษีว่าง = ฐาน × อัตรา ÷ 100 ปัดครึ่งขึ้นตามทศนิยมของปีบัญชี
// (ผู้ใช้พิมพ์ยอดภาษีเองได้ตามใบกำกับจริง — ใช้ค่าที่พิมพ์)
func (m *subledgerMutation) vat(v *SubledgerVat) error {
	v.TaxInvoiceNo = strings.TrimSpace(v.TaxInvoiceNo)
	v.OriginalInvoiceNo = strings.TrimSpace(v.OriginalInvoiceNo)
	v.PartnerTaxID = strings.TrimSpace(v.PartnerTaxID)
	v.PartnerBranchNo = strings.TrimSpace(v.PartnerBranchNo)
	v.PartnerName = strings.TrimSpace(v.PartnerName)
	v.ClaimReason = strings.TrimSpace(v.ClaimReason)
	v.Remark = strings.TrimSpace(v.Remark)
	if v.DocumentType == 0 {
		v.DocumentType = 1 // DEFAULT 1 ตาม vat.sql
	}
	if !subledgerID(v.ID) || (v.TaxType != 1 && v.TaxType != 2) || v.DocumentType < 1 || v.DocumentType > 3 ||
		v.TaxInvoiceNo == "" || utf8.RuneCountInString(v.TaxInvoiceNo) > 50 || !validDate(v.TaxInvoiceDate) ||
		utf8.RuneCountInString(v.OriginalInvoiceNo) > 50 || (v.OriginalInvoiceDate != "" && !validDate(v.OriginalInvoiceDate)) ||
		v.PartnerName == "" || utf8.RuneCountInString(v.PartnerName) > 255 ||
		utf8.RuneCountInString(v.ClaimReason) > 500 || utf8.RuneCountInString(v.Remark) > 500 {
		return fmt.Errorf("ข้อมูลภาษีมูลค่าเพิ่มไม่ครบหรือไม่ถูกต้อง")
	}
	if v.PartnerTaxID != "" && !subledgerTaxID.MatchString(v.PartnerTaxID) {
		return fmt.Errorf("เลขประจำตัวผู้เสียภาษีของคู่ค้าต้องเป็นตัวเลข 13 หลัก")
	}
	if v.PartnerBranchNo != "" && !subledgerTaxBranch.MatchString(v.PartnerBranchNo) {
		return fmt.Errorf("เลขสาขาของคู่ค้าต้องเป็นตัวเลข 5 หลัก (00000 = สำนักงานใหญ่)")
	}
	if v.DocumentType != 1 && (v.OriginalInvoiceNo == "" || v.OriginalInvoiceDate == "") {
		return fmt.Errorf("ใบเพิ่มหนี้/ใบลดหนี้ต้องระบุเลขที่และวันที่ใบกำกับภาษีเดิม")
	}
	hasPeriod := v.TaxPeriodYear != 0 || v.TaxPeriodMonth != 0
	if hasPeriod && (v.TaxPeriodYear < 1900 || v.TaxPeriodYear > 9999 || v.TaxPeriodMonth < 1 || v.TaxPeriodMonth > 12) {
		return fmt.Errorf("งวดภาษีต้องระบุทั้งปี (ค.ศ.) และเดือน 1–12")
	}
	switch {
	case v.TaxType == 1 && (v.ClaimStatus < 1 || v.ClaimStatus > 4):
		return fmt.Errorf("ภาษีซื้อต้องระบุสถานะการใช้สิทธิ")
	case v.TaxType == 2 && v.ClaimStatus != 0:
		return fmt.Errorf("ภาษีขายไม่มีสถานะการใช้สิทธิ")
	case v.TaxType == 2 && !hasPeriod:
		return fmt.Errorf("ภาษีขายต้องระบุงวดภาษี")
	case v.ClaimStatus == 1 && !hasPeriod:
		return fmt.Errorf("ภาษีซื้อที่ใช้สิทธิต้องระบุงวดภาษี")
	}
	rate := v.Rate.Decimal()
	if err := v.Rate.ValidateScale(2); err != nil || v.Rate == "" || rate.IsNegative() || rate.GreaterThan(hundred) {
		return fmt.Errorf("อัตราภาษีมูลค่าเพิ่มต้องอยู่ระหว่าง 0–100 และมีทศนิยมไม่เกิน 2 ตำแหน่ง")
	}
	if v.VatAmount == nil {
		vat := amountFromDecimal(v.BaseAmount.Decimal().Mul(rate).Div(hundred).Round(int32(m.scale)))
		v.VatAmount = &vat
	}
	for _, a := range []Amount{v.BaseAmount, v.ZeroRateAmount, v.ExemptAmount, *v.VatAmount} {
		d := a.Decimal()
		if err := a.ValidateScale(m.scale); err != nil || d.IsNegative() || d.GreaterThanOrEqual(vatAmountLimit) {
			return fmt.Errorf("ยอดเงินภาษีมูลค่าเพิ่มต้องไม่ติดลบ ไม่เกิน 14 หลัก และทศนิยมตามปีบัญชี")
		}
	}
	if v.PartnerCode != "" {
		var p SubledgerPartner
		if !validCode(v.PartnerCode) || m.loadJSON("gl_subledger_partners", "code", v.PartnerCode, &p) != nil || !p.IsActive {
			return fmt.Errorf("ไม่พบคู่ค้าของรายการภาษีมูลค่าเพิ่ม หรือคู่ค้าปิดใช้งาน")
		}
	}
	return nil
}

// VatRecord - รายการภาษีมูลค่าเพิ่มพร้อมใบสำคัญต้นทาง สำหรับรายงานภาษีซื้อ/ขาย และ ภ.พ.30
type VatRecord struct {
	JournalID string `json:"journalid"`
	DocNo     string `json:"docno"`
	DocDate   string `json:"docdate"`
	SubledgerVat
}

// VatRecordsForPeriod - VAT lines of POSTED vouchers of the company whose tax period is year/month.
// taxType 2 = sales (output VAT); taxType 1 = purchases with claim_status 1 (claimed in this period).
// เรียงตามวันที่ใบกำกับ เลขที่ใบสำคัญ และลำดับรายการในใบ; ใบร่าง/ลบ/กลับรายการแล้วไม่นับ
func VatRecordsForPeriod(ctx context.Context, db *sql.DB, company string, year, month, taxType int) ([]VatRecord, error) {
	if taxType != 1 && taxType != 2 {
		return nil, fmt.Errorf("invalid vat tax type %d", taxType)
	}
	if year < 1900 || year > 9999 || month < 1 || month > 12 {
		return nil, fmt.Errorf("invalid vat tax period %d-%d", year, month)
	}
	// jsonb containment เทียบตัวเลขตรง ๆ — แถวที่ไม่มีงวดหรือชนิดไม่ตรงจะไม่ถูกเลือก โดยไม่ต้อง cast ข้อความ
	filter := map[string]int{"tax_type": taxType, "tax_period_year": year, "tax_period_month": month}
	if taxType == 1 {
		filter["claim_status"] = 1
	}
	rawFilter, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
SELECT r.id, r.code, COALESCE(r.payload->>'date',''), v.item
FROM gl_records r
CROSS JOIN LATERAL jsonb_array_elements(CASE WHEN jsonb_typeof(r.payload->'details'->'vats') = 'array'
  THEN r.payload->'details'->'vats' ELSE '[]'::jsonb END) WITH ORDINALITY AS v(item, ord)
WHERE r.company = $1 AND r.kind = 'journals'
  AND r.payload->>'status' = 'posted'
  AND NOT COALESCE((r.payload->>'isdeleted')::boolean, false)
  AND v.item @> $2::jsonb
ORDER BY v.item->>'tax_invoice_date', r.code, v.ord`, company, string(rawFilter))
	if err != nil {
		return nil, fmt.Errorf("read vat records: %w", err)
	}
	defer rows.Close()
	records := []VatRecord{}
	for rows.Next() {
		var rec VatRecord
		var item []byte
		if err := rows.Scan(&rec.JournalID, &rec.DocNo, &rec.DocDate, &item); err != nil {
			return nil, fmt.Errorf("scan vat record: %w", err)
		}
		if err := json.Unmarshal(item, &rec.SubledgerVat); err != nil {
			return nil, fmt.Errorf("parse vat record of %s: %w", rec.DocNo, err)
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read vat records: %w", err)
	}
	return records, nil
}
