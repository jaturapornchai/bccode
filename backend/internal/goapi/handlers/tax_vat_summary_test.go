package handlers

import (
	"testing"

	"smlcloudplatform/internal/generalledger"
	"smlcloudplatform/internal/goapi/language"
)

func vatOn(r generalledger.VatRecord, date string) generalledger.VatRecord {
	r.TaxInvoiceDate = date
	return r
}

// รายงานสรุปยอดภาษี (Champ 5539): ภาษีซื้อ + ภาษีขายเรียงรวมตามวันที่ใบกำกับ, ใบลดหนี้ติดลบ, คงเหลือสะสม = ซื้อ − ขาย,
// รวมรายวัน = ซื้อ − ขายของวันนั้น (ไม่ใช่ผลบวกยอดคงเหลือแบบ Champ), ยอดทั้งงวดตรงกับ sumPP30 ข้อ 5/7 และข้อ 8/9
func TestBuildVatSummary(t *testing.T) {
	sales := []generalledger.VatRecord{
		vatOn(vatRecord("UV1", "IV001", 1, "1000", "0", "0", "70.004"), "2026-09-02"),
		vatOn(vatRecord("UV2", "CN001", 3, "100", "0", "0", "7"), "2026-09-05"),
	}
	purchases := []generalledger.VatRecord{
		vatOn(vatRecord("SV1", "PI001", 1, "400", "0", "0", "28"), "2026-09-02"),
		vatOn(vatRecord("SV2", "PI002", 1, "200", "0", "0", "14"), "2026-09-03"),
	}
	rows, days, totals := buildVatSummary(purchases, sales, "", "th")
	type view struct{ doc, in, out, balance string }
	want := []view{{"UV1", "", "70.00", "-70.00"}, {"SV1", "28.00", "", "-42.00"}, {"SV2", "14.00", "", "-28.00"}, {"UV2", "", "-7.00", "-21.00"}}
	if len(rows) != len(want) {
		t.Fatalf("rows = %+v", rows)
	}
	for i, w := range want {
		got := view{rows[i].DocNo, rows[i].TaxIn, rows[i].TaxOut, rows[i].Balance}
		if got != w || rows[i].No != i+1 {
			t.Fatalf("row %d = %+v want %+v", i, rows[i], w)
		}
	}
	if rows[3].Description != language.Text("gl_vat_document_credit_note", "th")+" บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด" {
		t.Fatalf("credit note description = %q", rows[3].Description)
	}
	wantDays := []TaxVatSummaryDay{
		{Date: "2026-09-02", Count: 2, TaxIn: "28.00", TaxOut: "70.00", Net: "-42.00"},
		{Date: "2026-09-03", Count: 1, TaxIn: "14.00", TaxOut: "0.00", Net: "14.00"},
		{Date: "2026-09-05", Count: 1, TaxIn: "0.00", TaxOut: "-7.00", Net: "7.00"},
	}
	if len(days) != len(wantDays) {
		t.Fatalf("days = %+v", days)
	}
	for i := range wantDays {
		if days[i] != wantDays[i] {
			t.Fatalf("day %d = %+v want %+v", i, days[i], wantDays[i])
		}
	}
	pp30 := sumPP30(sales, purchases)
	if totals.TaxOut != moneyText(pp30.outputVat) || totals.TaxIn != moneyText(pp30.inputVat) || totals.Count != 4 ||
		totals.Net != "-21.00" || totals.TaxPayable != "21.00" || totals.TaxExcess != "0.00" {
		t.Fatalf("totals = %+v pp30 out %s in %s", totals, pp30.outputVat, pp30.inputVat)
	}

	// ภาษีซื้อมากกว่า → ข้อ 9 ชำระเกิน; เรียงตามเลขที่เอกสารไม่มีรวมรายวัน
	rows, days, totals = buildVatSummary(purchases, sales[1:], "docno", "th")
	if len(days) != 0 || rows[0].DocNo != "SV1" || rows[2].DocNo != "UV2" || totals.TaxPayable != "0.00" || totals.TaxExcess != "49.00" {
		t.Fatalf("docno sort rows=%+v days=%+v totals=%+v", rows, days, totals)
	}
	// เรียงตามเลขที่ใบกำกับ: CN001 < PI001 < PI002
	if rows, _, _ = buildVatSummary(purchases, sales[1:], "taxno", "th"); rows[0].TaxInvoiceNo != "CN001" || rows[2].TaxInvoiceNo != "PI002" {
		t.Fatalf("taxno sort rows=%+v", rows)
	}
	if rows, days, totals = buildVatSummary(nil, nil, "", "th"); len(rows) != 0 || len(days) != 0 || totals.TaxPayable != "0.00" || totals.Net != "0.00" {
		t.Fatalf("empty = %+v %+v %+v", rows, days, totals)
	}
}
