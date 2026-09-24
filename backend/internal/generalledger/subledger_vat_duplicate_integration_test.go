//go:build integration

package generalledger

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// ใบกำกับภาษีฉบับเดียวกัน = ผู้ออก (เลขผู้เสียภาษี + สาขา) + เลขที่ + วันที่ (ม.86/4) — ผู้ขายต่างรายใช้เลขเดียวกันได้
// เตือนเท่านั้น ไม่บล็อก; เทียบทุกงวดของบริษัท เฉพาะใบที่ยังมีผล (ร่าง/ผ่านบัญชี ไม่ถูกลบ ไม่ใช่ใบกลับรายการ/ต้นฉบับที่กลับแล้ว)
func TestPostgresVatDuplicateInvoiceWarnings(t *testing.T) {
	f := newPGIntegrityFixture(t)
	partners := []SubledgerPartner{
		{Code: "SUP-A1", Name: "บริษัท สยามปูนซีเมนต์ไทย จำกัด", TaxID: "0105561111115", IsSupplier: true, IsActive: true},
		{Code: "SUP-A2", Name: "บริษัท สยามปูนซีเมนต์ไทย จำกัด (รหัสซ้ำ)", TaxID: "0105561111115", IsSupplier: true, IsActive: true},
		{Code: "SUP-N1", Name: "ร้านวัสดุก่อสร้างทองดี", IsSupplier: true, IsActive: true},
		{Code: "SUP-N2", Name: "ร้านเหล็กรุ่งเรือง", IsSupplier: true, IsActive: true},
	}
	seq := 0
	purchase := func(partnerCode, taxID, branch, invoice, date string) SubledgerVat {
		seq++
		v := vatItem(fmt.Sprintf("P%02d", seq), 1, 1, invoice, date, 2026, 1, 1, "1000", nil)
		v.PartnerCode, v.PartnerTaxID, v.PartnerBranchNo = partnerCode, taxID, branch
		return v
	}
	sale := func(invoice, date string) SubledgerVat {
		seq++
		return vatItem(fmt.Sprintf("S%02d", seq), 2, 1, invoice, date, 2026, 1, 0, "1000", nil)
	}
	create := func(doc, branch string, withPartners bool, vats ...SubledgerVat) Journal {
		t.Helper()
		j := Journal{DocNo: doc, Date: "2026-01-15", BookCode: "PV", FiscalYear: "2026", Description: "ซื้อวัสดุก่อสร้าง " + doc, Kind: "manual", BranchCode: branch,
			Lines:   []Line{{AccountCode: "5000", Debit: "1070", Credit: "0"}, {AccountCode: "1000", Debit: "0", Credit: "1070"}},
			Details: &JournalDetails{Vats: vats}}
		if withPartners {
			j.Details.Partners = partners
		}
		return f.journal(f.run(Command{Resource: "journals", Action: "create", Journal: &j}).ID)
	}
	const taxX, taxY = "0105558012349", "0105561234560"
	// ผู้ขาย X สำนักงานใหญ่ ใบ IV-100 วันที่ 10 ม.ค.
	f.post(create("PV01", "B1", true, purchase("", taxX, "00000", "IV-100", "2026-01-10")))
	// ผู้ขาย Y ใช้เลขเดียวกัน วันเดียวกัน → คนละผู้ออก ไม่เตือน
	f.post(create("PV02", "B1", false, purchase("", taxY, "00000", "IV-100", "2026-01-10")))
	// ผู้ขาย X สาขา 00001 เลขเดียวกัน → คนละผู้ออก (สาขาต่างกัน) ไม่เตือน
	f.post(create("PV03", "B1", false, purchase("", taxX, "00001", "IV-100", "2026-01-10")))
	// ผู้ขาย X เลขเดียวกัน คนละวันที่ → ไม่เตือน (ผู้ออกเริ่มนับเลขใหม่ / ใบแทนตาม ป.86/2542 ข้อ 25 ได้เลขใหม่วันที่เดิม)
	f.post(create("PV04", "B1", false, purchase("", taxX, "00000", "IV-100", "2026-01-11")))
	// ผู้ขาย X ใบ IV-200 ถูกคีย์ 2 ใบผ่านบัญชี (ตัวพิมพ์เล็ก/ช่องว่าง/สาขาว่าง = 00000) + 1 ใบร่าง → เตือนทั้งสองแถว
	f.post(create("PV05", "B1", false, purchase("", taxX, "00000", "IV-200", "2026-01-12")))
	f.post(create("PV06", "B1", false, purchase("", taxX, "", "  iv-200 ", "2026-01-12")))
	create("PV07", "B1", false, purchase("", taxX, "00000", "IV  200", "2026-01-12")) // ช่องว่างกลางต่างจากขีด → คนละเลข
	create("PV08", "B1", false, purchase("", taxX, "00000", "iv-200", "2026-01-12"))  // ใบร่างก็นับ
	// รหัสคู่ค้าต่างกันแต่เลขผู้เสียภาษีเดียวกัน → ผู้ออกรายเดียวกัน เตือน
	f.post(create("PV09", "B1", false, purchase("SUP-A1", "0105561111115", "00000", "INV-9", "2026-01-13")))
	f.post(create("PV10", "B1", false, purchase("SUP-A2", "0105561111115", "00000", "INV-9", "2026-01-13")))
	// ไม่มีเลขผู้เสียภาษี รหัสคู่ค้าต่างกัน → คนละผู้ออก ไม่เตือน
	f.post(create("PV11", "B1", false, purchase("SUP-N1", "", "", "RC-1", "2026-01-14")))
	f.post(create("PV12", "B1", false, purchase("SUP-N2", "", "", "RC-1", "2026-01-14")))
	// ต้นฉบับกลับรายการแล้ว + ใบกลับรายการ แล้วบันทึกใหม่ → ใบใหม่ไม่เตือน
	orig := f.post(create("PV13", "B1", false, purchase("", taxY, "00000", "IV-300", "2026-01-14")))
	f.run(Command{Resource: "journals", Action: "reverse", ID: orig.ID, Version: orig.Version, DocNo: "REV13", Date: "2026-01-16", Reason: "บันทึกผิดใบ"})
	f.post(create("PV14", "B1", false, purchase("", taxY, "00000", "IV-300", "2026-01-14")))
	// ภาษีขาย: ผู้ออก = บริษัทเรา + สาขาของใบสำคัญ — สาขาเดียวกันเลขซ้ำ = เตือน, คนละสาขา = ไม่เตือน
	f.post(create("SV01", "B1", false, sale("SO-1", "2026-01-20")))
	f.post(create("SV02", "B2", false, sale("SO-1", "2026-01-20")))
	f.post(create("SV03", "B1", false, sale("so-1", "2026-01-20")))
	// ใบกำกับฉบับเดียวกันในงวดอื่น (ก.พ.) ยังนับเป็นซ้ำ — เทียบทุกงวดของบริษัท
	feb := purchase("", taxY, "00000", "IV-400", "2026-01-25")
	feb.TaxPeriodMonth = 2
	f.post(create("PV15", "B1", false, feb))
	f.post(create("PV16", "B1", false, purchase("", taxY, "00000", "IV-400", "2026-01-25")))

	ctx := context.Background()
	warnings := func(taxType int) map[string]string {
		t.Helper()
		records, err := VatRecordsForPeriod(ctx, f.db, "C", 2026, 1, taxType)
		if err != nil {
			t.Fatal(err)
		}
		out := map[string]string{}
		for _, r := range records {
			if r.DuplicateDocNos == nil {
				t.Fatalf("%s: DuplicateDocNos must be [] not nil", r.DocNo)
			}
			out[r.DocNo] = strings.Join(r.DuplicateDocNos, ",")
		}
		return out
	}
	want := map[string]string{"PV01": "", "PV02": "", "PV03": "", "PV04": "", "PV05": "PV06,PV08", "PV06": "PV05,PV08",
		"PV09": "PV10", "PV10": "PV09", "PV11": "", "PV12": "", "PV14": "", "PV16": "PV15"}
	got := warnings(1)
	for doc, dup := range want {
		if v, ok := got[doc]; !ok || v != dup {
			t.Errorf("purchase %s duplicates = %q (present %v), want %q", doc, v, ok, dup)
		}
	}
	if _, ok := got["PV13"]; ok {
		t.Error("reversed original must not be in the register")
	}
	sales := warnings(2)
	for doc, dup := range map[string]string{"SV01": "SV03", "SV02": "", "SV03": "SV01"} {
		if sales[doc] != dup {
			t.Errorf("sale %s duplicates = %q want %q", doc, sales[doc], dup)
		}
	}
}

// งวดใช้สิทธิภาษีซื้อเกินช่วง (ม.82/3 + ประกาศอธิบดีฯ VAT ฉบับที่ 4/76) ต้องถูกปฏิเสธทั้งตอนบันทึก และตอนแก้หลังผ่านบัญชี (reconcile)
// — ตรวจค่าใน PostgreSQL ทีละขั้นว่าไม่ถูกเขียน
func TestPostgresPurchaseClaimWindowOnSaveAndReconcile(t *testing.T) {
	f := newPGIntegrityFixture(t)
	claimOf := func(id string) string {
		t.Helper()
		var period string
		if err := f.db.QueryRow(`SELECT COALESCE(payload->'details'->'vats'->0->>'tax_period_year','')||'-'||COALESCE(payload->'details'->'vats'->0->>'tax_period_month','')||'/'||COALESCE(payload->'details'->'vats'->0->>'claim_status','')
FROM gl_records WHERE company='C' AND kind='journals' AND id=$1`, id).Scan(&period); err != nil {
			t.Fatal(err)
		}
		return period
	}
	expectCode := func(err error, code string) {
		t.Helper()
		user, ok := AsUserError(err)
		if !ok || user.Code != code || user.Field != "tax_period_month" {
			t.Fatalf("err = %#v, want %s on tax_period_month", err, code)
		}
	}
	journal := func(v SubledgerVat) Journal {
		return Journal{DocNo: "PV6901-0042", Date: "2026-01-20", BookCode: "PV", FiscalYear: "2026", Description: "ซื้อปูนซีเมนต์ปอร์ตแลนด์ 50 กก. จำนวน 500 ถุง เป็นเงินเชื่อ",
			Kind: "manual", BranchCode: "B1", Lines: []Line{{AccountCode: "5000", Debit: "107000", Credit: "0"}, {AccountCode: "1000", Debit: "0", Credit: "107000"}},
			Details: &JournalDetails{Vats: []SubledgerVat{v}}}
	}

	// 1) สร้าง: ใบกำกับ 5 ม.ค. 2026 ใช้สิทธิงวด ส.ค. 2026 (เดือนที่ 7) → ปฏิเสธ และไม่มีใบใน PG
	bad := journal(vatItem("P1", 1, 1, "IV6901-0042", "2026-01-05", 2026, 8, 1, "100000", nil))
	_, err := f.execute(f.scope, Command{Resource: "journals", Action: "create", Journal: &bad})
	expectCode(err, "vat_claim_window_exceeded")
	var n int
	if err := f.db.QueryRow(`SELECT COUNT(*) FROM gl_records WHERE company='C' AND kind='journals' AND code='PV6901-0042'`).Scan(&n); err != nil || n != 0 {
		t.Fatalf("rejected journal written to PG: n=%d err=%v", n, err)
	}
	// งวดก่อนเดือนที่ออกใบกำกับ (ธ.ค. 2025) → ปฏิเสธ
	early := journal(vatItem("P1", 1, 1, "IV6901-0042", "2026-01-05", 2025, 12, 1, "100000", nil))
	_, err = f.execute(f.scope, Command{Resource: "journals", Action: "create", Journal: &early})
	expectCode(err, "vat_claim_before_invoice_month")

	// 2) สร้างเป็น "รอใช้สิทธิ" (ไม่มีงวด) ได้ แล้วผ่านบัญชี
	pending := journal(vatItem("P1", 1, 1, "IV6901-0042", "2026-01-05", 0, 0, 3, "100000", nil))
	j := f.post(f.journal(f.run(Command{Resource: "journals", Action: "create", Journal: &pending}).ID))
	if got := claimOf(j.ID); got != "-/3" {
		t.Fatalf("PG after post = %s", got)
	}

	// 3) หลังผ่านบัญชี เปลี่ยนเป็นใช้สิทธิงวด ส.ค. (เกินช่วง) ผ่าน reconcile → ปฏิเสธ ค่าใน PG ไม่เปลี่ยน
	_, err = f.execute(f.scope, subledgerReconcile(f, j, JournalDetails{Vats: []SubledgerVat{vatItem("P1", 1, 1, "IV6901-0042", "2026-01-05", 2026, 8, 1, "100000", nil)}}))
	expectCode(err, "vat_claim_window_exceeded")
	if got := claimOf(j.ID); got != "-/3" {
		t.Fatalf("PG after rejected reconcile = %s", got)
	}

	// 4) ใช้สิทธิงวด ก.ค. (เดือนที่ 6 — ยังอยู่ในช่วง) ผ่าน reconcile ได้ และ PG เปลี่ยนในแถวเดิม
	f.run(subledgerReconcile(f, j, JournalDetails{Vats: []SubledgerVat{vatItem("P1", 1, 1, "IV6901-0042", "2026-01-05", 2026, 7, 1, "100000", nil)}}))
	if got := claimOf(j.ID); got != "2026-7/1" {
		t.Fatalf("PG after valid reconcile = %s", got)
	}
}
