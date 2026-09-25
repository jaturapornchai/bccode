//go:build integration

package generalledger

import "testing"

// ภาษีหัก ณ ที่จ่ายของใบที่กลับรายการหลังงวด: เครดิต ภ.ง.ด.50/51 ยังนับ (แยกนับเตือน) และยังออก 50 ทวิ ได้ —
// กติกาเดียวกับรายงานภาษีหัก/ภาษีซื้อขาย; กลับภายในงวด = ไม่นับ/ไม่พบ; ใบยอดยกมาไม่ใช่หลักฐานการหักใหม่
func TestPostgresWithholdingReversedAfterPeriod(t *testing.T) {
	f := newPGIntegrityFixture(t)
	partner := SubledgerPartner{Code: "C101", Name: "บริษัท หอมกรุ่น คอฟฟี่ แอนด์ เบเกอรี่ จำกัด", TaxID: "0105558012349", TaxBranch: "00000", IsCustomer: true, IsSupplier: true, IsActive: true}
	create := func(doc, date string, direction int, rate string) Journal {
		t.Helper()
		j := Journal{DocNo: doc, Date: date, BookCode: "JV", FiscalYear: "2026", Kind: "manual", BranchCode: "B1", Description: "ค่าบริการขนส่งวัสดุก่อสร้าง " + doc,
			Lines: []Line{{AccountCode: "1000", Debit: "1000", Credit: "0"}, {AccountCode: "4000", Debit: "0", Credit: "1000"}},
			Details: &JournalDetails{Partners: []SubledgerPartner{partner},
				Withholdings: []SubledgerWithholding{{ID: "W1", Direction: direction, FormType: "PND53", PartnerCode: "C101", PaymentDate: date,
					IncomeType: "3_tres", Description: "ค่าบริการขนส่ง", Condition: 1, Rate: Amount(rate), BaseAmount: "1000"}}}}
		return f.post(f.journal(f.run(Command{Resource: "journals", Action: "create", Journal: &j}).ID))
	}
	reverse := func(j Journal, doc, date string) {
		t.Helper()
		f.run(Command{Resource: "journals", Action: "reverse", ID: j.ID, Version: j.Version, DocNo: doc, Date: date, Reason: "บันทึกหนังสือรับรองผิดฉบับ"})
	}
	// ลูกค้าหักเรา: A 3% มี.ค. กลับ ส.ค. (หลังครึ่งปีแรก), B 5% เม.ย. กลับในเดือน, C 1% พ.ค. ปกติ, D 2% พ.ค. แต่เป็นใบยอดยกมา
	reverse(create("RV-A", "2026-03-10", 2, "3"), "REV-A", "2026-08-05")
	reverse(create("RV-B", "2026-04-10", 2, "5"), "REV-B", "2026-04-20")
	create("RV-C", "2026-05-10", 2, "1")
	d := create("RV-D", "2026-05-11", 2, "2")
	if _, err := f.db.Exec(`UPDATE gl_records SET payload = jsonb_set(payload, '{kind}', '"opening"') WHERE company='C' AND kind='journals' AND id=$1`, d.ID); err != nil {
		t.Fatal(err)
	}
	credit := func(from, to string) WithheldCredit {
		t.Helper()
		c, err := WithheldFromCompanyTotal(f.ctx, f.db, "C", from, to)
		if err != nil {
			t.Fatal(err)
		}
		return c
	}
	if c := credit("2026-01-01", "2026-06-30"); c.Total.StringFixed(2) != "40.00" || c.Count != 2 || c.ReversedLater != 1 {
		t.Fatalf("PND51 half-year credit = %s count %d reversed-later %d, want 40.00 (A 30 + C 10) 2 1", c.Total.StringFixed(2), c.Count, c.ReversedLater)
	}
	if c := credit("2026-01-01", "2026-12-31"); c.Total.StringFixed(2) != "10.00" || c.Count != 1 || c.ReversedLater != 0 {
		t.Fatalf("PND50 full-year credit = %s count %d reversed-later %d, want 10.00 (C only) 1 0", c.Total.StringFixed(2), c.Count, c.ReversedLater)
	}

	// เราหักผู้รับ: E พ.ค. กลับ มิ.ย. = ยังพิมพ์ 50 ทวิ ได้; F พ.ค. กลับในเดือน = ไม่พบ
	e := create("PV-E", "2026-05-15", 1, "3")
	reverse(e, "REV-E", "2026-06-02")
	fj := create("PV-F", "2026-05-16", 1, "3")
	reverse(fj, "REV-F", "2026-05-31")
	if item, _, err := RecordedWithholding(f.ctx, f.db, "C", e.ID, "W1"); err != nil || item.TaxAmount == nil || string(*item.TaxAmount) != "30" {
		t.Fatalf("certificate of voucher reversed next month = %+v err=%v", item, err)
	}
	if _, _, err := RecordedWithholding(f.ctx, f.db, "C", fj.ID, "W1"); err != ErrNotFound {
		t.Fatalf("certificate of voucher reversed in the same month must be not found: %v", err)
	}
}
