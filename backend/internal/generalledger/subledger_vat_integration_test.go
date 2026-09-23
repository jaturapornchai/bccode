//go:build integration

package generalledger

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func vatItem(id string, taxType, documentType int, invoice, date string, year, month, claim int, base string, vat *Amount) SubledgerVat {
	v := SubledgerVat{ID: id, TaxType: taxType, DocumentType: documentType, TaxInvoiceNo: invoice, TaxInvoiceDate: date,
		TaxPeriodYear: year, TaxPeriodMonth: month, ClaimStatus: claim, PartnerName: "บริษัท สยามพาณิชย์ โฮลดิ้ง จำกัด",
		PartnerTaxID: "0105558012349", PartnerBranchNo: "00000", BaseAmount: Amount(base), Rate: "7", VatAmount: vat}
	if documentType != 1 {
		v.OriginalInvoiceNo, v.OriginalInvoiceDate = "IV-ORIGINAL", "2026-01-05"
	}
	return v
}

// ฐานภาษีมูลค่าเพิ่มต้องแก้ได้เสมอ — ตรวจค่าที่บันทึกใน PostgreSQL ทีละขั้น (สร้าง → แก้ร่าง → ปฏิเสธค่าผิด → ผ่านบัญชี → แก้หลังผ่านบัญชี)
func TestPostgresVatBaseEditableAlways(t *testing.T) {
	f := newPGIntegrityFixture(t)
	stored := func(id string) []map[string]any {
		t.Helper()
		var raw []byte
		if err := f.db.QueryRow(`SELECT COALESCE(payload->'details'->'vats','[]'::jsonb) FROM gl_records WHERE company='C' AND kind='journals' AND id=$1`, id).Scan(&raw); err != nil {
			t.Fatal(err)
		}
		var rows []map[string]any
		if err := json.Unmarshal(raw, &rows); err != nil {
			t.Fatal(err)
		}
		return rows
	}
	expect := func(id, base, vat string) {
		t.Helper()
		rows := stored(id)
		if len(rows) != 1 || rows[0]["base_amount"] != base || rows[0]["vat_amount"] != vat {
			t.Fatalf("PG vats = %v, want base=%s vat=%s", rows, base, vat)
		}
	}

	// 1) สร้าง: ขายเชื่อ 107,000 รวม VAT — ฐาน 100,000 ภาษีว่าง = ระบบคำนวณ 7% (คู่ค้าระบุหรือไม่ก็ได้)
	j := Journal{DocNo: "UV1", Date: "2026-01-10", BookCode: "UV", FiscalYear: "2026", Description: "ขายวัสดุก่อสร้าง ให้ลูกค้าโครงการก่อสร้าง A เป็นเงินเชื่อ", Kind: "manual", BranchCode: "B1",
		Lines: []Line{{AccountCode: "1000", Debit: "107000", Credit: "0"}, {AccountCode: "4000", Debit: "0", Credit: "107000"}},
		Details: &JournalDetails{
			Partners: []SubledgerPartner{{Code: "SIAM", Name: "บริษัท สยามพาณิชย์ โฮลดิ้ง จำกัด", TaxID: "0105558012349", IsCustomer: true, IsActive: true}},
			Vats: []SubledgerVat{func() SubledgerVat {
				v := vatItem("V1", 2, 1, "IV2601-001", "2026-01-10", 2026, 1, 0, "100000.00", nil)
				v.PartnerCode = "SIAM"
				return v
			}()},
		}}
	r := f.run(Command{Resource: "journals", Action: "create", Journal: &j})
	expect(r.ID, "100000", "7000")

	// 2) แก้ฐานในใบร่าง + พิมพ์ภาษีเองตามใบกำกับจริง — แถวเดิมเปลี่ยน ไม่ใช่เพิ่มแถว
	j = f.journal(r.ID)
	typed := Amount("6300.04")
	j.Details.Vats = []SubledgerVat{vatItem("V1", 2, 1, "IV2601-001", "2026-01-10", 2026, 1, 0, "90000.50", &typed)}
	f.run(Command{Resource: "journals", Action: "update", ID: j.ID, Version: j.Version, Journal: &j})
	expect(j.ID, "90000.5", "6300.04")

	// 3) ค่าที่ผิดตาม vat.sql ต้องถูกปฏิเสธ และค่าใน PG ต้องไม่เปลี่ยน
	j = f.journal(j.ID)
	negative := Amount("-1")
	for name, bad := range map[string]SubledgerVat{
		"negative vat":        vatItem("V1", 2, 1, "IV2601-001", "2026-01-10", 2026, 1, 0, "100", &negative),
		"sale without period": vatItem("V1", 2, 1, "IV2601-001", "2026-01-10", 0, 0, 0, "100", nil),
		"credit note no origin": func() SubledgerVat {
			v := vatItem("V1", 2, 3, "CN2601-001", "2026-01-10", 2026, 1, 0, "100", nil)
			v.OriginalInvoiceNo = ""
			return v
		}(),
		"purchase no status": vatItem("V1", 1, 1, "PI-1", "2026-01-10", 2026, 1, 0, "100", nil),
		"unknown partner": func() SubledgerVat {
			v := vatItem("V1", 2, 1, "IV2601-001", "2026-01-10", 2026, 1, 0, "100", nil)
			v.PartnerCode = "NOPE"
			return v
		}(),
	} {
		next := j
		next.Details = &JournalDetails{Vats: []SubledgerVat{bad}}
		if _, err := f.execute(f.scope, Command{Resource: "journals", Action: "update", ID: j.ID, Version: j.Version, Journal: &next}); err == nil {
			t.Fatalf("%s: expected rejection", name)
		}
	}
	expect(j.ID, "90000.5", "6300.04")

	// 4) ผ่านบัญชีแล้วยังแก้ฐานได้ (reconcile + เหตุผล) — ยอด GL ไม่เปลี่ยน, audit เก็บค่าก่อน/หลัง
	j = f.post(j)
	noReason := subledgerReconcile(f, j, JournalDetails{Vats: []SubledgerVat{vatItem("V1", 2, 1, "IV2601-001", "2026-01-10", 2026, 1, 0, "95000", nil)}})
	noReason.Reason = "  "
	if _, err := f.execute(f.scope, noReason); err == nil {
		t.Fatal("tax-base edit after posting without a reason must be rejected")
	}
	expect(j.ID, "90000.5", "6300.04")
	f.run(subledgerReconcile(f, j, JournalDetails{Vats: []SubledgerVat{vatItem("V1", 2, 1, "IV2601-001", "2026-01-10", 2026, 1, 0, "95000", nil)}}))
	expect(j.ID, "95000", "6650")
	var before, reason string
	if err := f.db.QueryRow(`SELECT payload->'before'->0->>'base_amount', payload->>'reason' FROM gl_subledger_audit WHERE company='C' AND journal_id=$1 AND action='vat_replace'`, j.ID).Scan(&before, &reason); err != nil || before != "90000.5" || reason != "reconcile evidence" {
		t.Fatalf("audit before base = %q reason = %q err=%v", before, reason, err)
	}
	var debit string
	if err := f.db.QueryRow(`SELECT SUM(debit)::text FROM gl_lines WHERE company='C' AND journal_id=$1`, j.ID).Scan(&debit); err != nil || !strings.HasPrefix(debit, "107000") {
		t.Fatalf("GL debit after tax-base edit = %q err=%v (must stay 107000)", debit, err)
	}
	if got := f.journal(j.ID).Status; got != "posted" {
		t.Fatalf("status after reconcile = %s", got)
	}
}

// VatRecordsForPeriod: เฉพาะใบที่ผ่านบัญชี, ตามงวดภาษี (ไม่ใช่วันที่ใบสำคัญ), ขาย = ทุกรายการ, ซื้อ = เฉพาะใช้สิทธิ (1)
func TestPostgresVatRecordsForPeriod(t *testing.T) {
	f := newPGIntegrityFixture(t)
	create := func(doc, date string, vats ...SubledgerVat) Journal {
		t.Helper()
		j := Journal{DocNo: doc, Date: date, BookCode: "JV", FiscalYear: "2026", Description: "บันทึกภาษีมูลค่าเพิ่ม " + doc, Kind: "manual", BranchCode: "B1",
			Lines:   []Line{{AccountCode: "1000", Debit: "1000", Credit: "0"}, {AccountCode: "4000", Debit: "0", Credit: "1000"}},
			Details: &JournalDetails{Vats: vats}}
		return f.journal(f.run(Command{Resource: "journals", Action: "create", Journal: &j}).ID)
	}
	// JV-A ผ่านบัญชี: ขาย ม.ค. 2 ใบ (ใบกำกับ + ใบลดหนี้) + ขาย ก.พ. 1 ใบ; รายการในใบเรียงวันที่ใบกำกับ
	a := f.post(create("JV-A", "2026-01-20",
		vatItem("S2", 2, 3, "CN-1", "2026-01-25", 2026, 1, 0, "100", nil),
		vatItem("S1", 2, 1, "IV-1", "2026-01-20", 2026, 1, 0, "1000", nil),
		vatItem("S3", 2, 1, "IV-2", "2026-02-01", 2026, 2, 0, "500", nil)))
	// JV-B ผ่านบัญชี (ใบสำคัญเดือน ม.ค.): ซื้อใช้สิทธิ ก.พ., ซื้อต้องห้าม ม.ค., ซื้อรอใช้สิทธิ (ไม่มีงวด), ซื้อใช้สิทธิ ม.ค.
	f.post(create("JV-B", "2026-01-15",
		vatItem("P1", 1, 1, "PI-1", "2026-01-15", 2026, 2, 1, "400", nil),
		vatItem("P2", 1, 1, "PI-2", "2026-01-15", 2026, 1, 2, "300", nil),
		vatItem("P3", 1, 1, "PI-3", "2026-01-15", 0, 0, 3, "200", nil),
		vatItem("P4", 1, 1, "PI-4", "2026-01-14", 2026, 1, 1, "100", nil)))
	// JV-C ร่าง, JV-D ผ่านแล้วกลับรายการ — ต้องไม่นับทั้งคู่
	create("JV-C", "2026-01-21", vatItem("S9", 2, 1, "IV-9", "2026-01-21", 2026, 1, 0, "900", nil))
	d := f.post(create("JV-D", "2026-01-22", vatItem("S8", 2, 1, "IV-8", "2026-01-22", 2026, 1, 0, "800", nil)))
	f.run(Command{Resource: "journals", Action: "reverse", ID: d.ID, Version: d.Version, DocNo: "REV-D", Date: "2026-01-23", Reason: "ออกใบกำกับผิด"})
	// บริษัทอื่นต้องไม่เห็น
	if _, err := f.db.Exec(`INSERT INTO gl_records(company,kind,id,code,version,payload) SELECT 'OTHER',kind,id,code,version,payload FROM gl_records WHERE company='C' AND id=$1`, a.ID); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	invoices := func(year, month, taxType int) []string {
		t.Helper()
		records, err := VatRecordsForPeriod(ctx, f.db, "C", year, month, taxType)
		if err != nil {
			t.Fatal(err)
		}
		out := []string{}
		for _, r := range records {
			if r.DocNo == "" || r.JournalID == "" || r.DocDate == "" || r.VatAmount == nil {
				t.Fatalf("record without voucher/vat: %+v", r)
			}
			out = append(out, r.DocNo+":"+r.TaxInvoiceNo+":"+string(*r.VatAmount))
		}
		return out
	}
	check := func(got []string, want ...string) {
		t.Helper()
		if strings.Join(got, ",") != strings.Join(want, ",") {
			t.Fatalf("got %v want %v", got, want)
		}
	}
	check(invoices(2026, 1, 2), "JV-A:IV-1:70", "JV-A:CN-1:7")
	check(invoices(2026, 2, 2), "JV-A:IV-2:35")
	check(invoices(2026, 1, 1), "JV-B:PI-4:7")
	check(invoices(2026, 2, 1), "JV-B:PI-1:28")
	check(invoices(2026, 3, 2))
	if _, err := VatRecordsForPeriod(ctx, f.db, "C", 2026, 1, 3); err == nil {
		t.Fatal("tax type 3 must be rejected")
	}
	if _, err := VatRecordsForPeriod(ctx, f.db, "C", 2026, 13, 2); err == nil {
		t.Fatal("month 13 must be rejected")
	}
}
