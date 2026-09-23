package handlers

import (
	"context"
	"database/sql"

	"smlcloudplatform/internal/generalledger"
)

// fillVatForm - ภ.พ.30 ข้อ 1-3, 5-7 จากรายการภาษีขาย (งวดภาษีนี้) และภาษีซื้อที่ใช้สิทธิในงวดนี้
// ฐานภาษีคือยอดที่ผู้ใช้บันทึกในรายการภาษีของใบสำคัญ (ไม่คำนวณย้อนจากภาษี 7%)
func fillVatForm(ctx context.Context, db *sql.DB, company string, year, month int, f *formFiller) ([]TaxFormNote, error) {
	if err := generalledger.EnsureSchema(ctx, db); err != nil {
		return nil, err
	}
	sales, err := generalledger.VatRecordsForPeriod(ctx, db, company, year, month, 2)
	if err != nil {
		return nil, err
	}
	purchases, err := generalledger.VatRecordsForPeriod(ctx, db, company, year, month, 1)
	if err != nil {
		return nil, err
	}
	// ยอดเดียวกับรายงานภาษีซื้อ/ขาย (ปัดต่อรายการ ใบลดหนี้เป็นลบ) — ข้อ 1 = ฐาน + อัตรา 0 + ยกเว้น
	t := sumPP30(sales, purchases)
	f.set("sales_amount", moneyText(t.salesTaxable.Add(t.salesZeroRated).Add(t.salesExempt)))
	f.set("sales_zero_rate", moneyText(t.salesZeroRated))
	f.set("sales_exempt", moneyText(t.salesExempt))
	f.set("output_tax", moneyText(t.outputVat))
	f.set("purchase_amount", moneyText(t.purchaseTaxable))
	f.set("input_tax", moneyText(t.inputVat))
	notes := []TaxFormNote{}
	if len(sales)+len(purchases) == 0 {
		notes = append(notes, TaxFormNote{Key: "tax_form_note_no_vat"})
	} else {
		notes = append(notes, TaxFormNote{Key: "tax_form_note_vat_records", Count: len(sales) + len(purchases)})
	}
	return append(notes, TaxFormNote{Key: "tax_form_note_vat_forward"}), nil
}
