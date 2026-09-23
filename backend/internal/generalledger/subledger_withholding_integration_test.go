//go:build integration

package generalledger

import (
	"encoding/json"
	"strings"
	"testing"
)

// ฐานภาษีหัก ณ ที่จ่ายต้องแก้ได้เสมอ (ลุงจืด 2026-09-23) — ตรวจค่าที่บันทึกใน PostgreSQL ทีละขั้น
func TestPostgresWithholdingBaseEditableAlways(t *testing.T) {
	f := newPGIntegrityFixture(t)
	stored := func(id string) []map[string]any {
		t.Helper()
		var raw []byte
		if err := f.db.QueryRow(`SELECT COALESCE(payload->'details'->'withholdings','[]'::jsonb) FROM gl_records WHERE company='C' AND kind='journals' AND id=$1`, id).Scan(&raw); err != nil {
			t.Fatal(err)
		}
		var rows []map[string]any
		if err := json.Unmarshal(raw, &rows); err != nil {
			t.Fatal(err)
		}
		return rows
	}
	expect := func(id, base, tax string) {
		t.Helper()
		rows := stored(id)
		if len(rows) != 1 || rows[0]["base_amount"] != base || rows[0]["tax_amount"] != tax {
			t.Fatalf("PG withholdings = %v, want base=%s tax=%s", rows, base, tax)
		}
	}
	item := func(base, rate string, tax *Amount) SubledgerWithholding {
		return SubledgerWithholding{ID: "W1", Direction: 1, FormType: "PND53", PartnerCode: "TRANS", PaymentDate: "2026-01-10", IncomeType: "3_tres", Description: "ค่าขนส่งสินค้า", Condition: 1, Rate: Amount(rate), BaseAmount: Amount(base), TaxAmount: tax}
	}

	// 1) สร้าง: ยอดจ่าย 107,000 รวม VAT แต่ฐานภาษีหักคือก่อน VAT 100,000 — ภาษีว่าง = ระบบคำนวณ 3%
	j := Journal{DocNo: "PV1", Date: "2026-01-10", BookCode: "JV", FiscalYear: "2026", Description: "จ่ายค่าขนส่ง หักภาษี ณ ที่จ่าย 3%", Kind: "manual", BranchCode: "B1",
		Lines: []Line{{AccountCode: "1000", Debit: "107000", Credit: "0"}, {AccountCode: "4000", Debit: "0", Credit: "107000"}},
		Details: &JournalDetails{
			Partners:     []SubledgerPartner{{Code: "TRANS", Name: "บริษัท ขนส่งไทยเร็ว จำกัด", TaxID: "0105558012349", IsSupplier: true, IsActive: true}},
			Withholdings: []SubledgerWithholding{item("100000.00", "3", nil)},
		}}
	r := f.run(Command{Resource: "journals", Action: "create", Journal: &j})
	expect(r.ID, "100000", "3000")

	// 2) แก้ฐานในใบร่าง + พิมพ์ภาษีเอง (ปัดเศษตามเอกสารจริง) — ค่าเดิมในแถวเดิมต้องเปลี่ยน ไม่ใช่เพิ่มแถว
	j = f.journal(r.ID)
	tax := Amount("2700.02")
	j.Details.Withholdings = []SubledgerWithholding{item("90000.50", "3", &tax)}
	f.run(Command{Resource: "journals", Action: "update", ID: j.ID, Version: j.Version, Journal: &j})
	expect(j.ID, "90000.5", "2700.02")

	// 3) ค่าที่ผิดต้องถูกปฏิเสธ และค่าใน PG ต้องไม่เปลี่ยน
	j = f.journal(j.ID)
	over := Amount("90001")
	for name, bad := range map[string]SubledgerWithholding{
		"tax over base": item("90000.50", "3", &over),
		"rate over 100": item("90000.50", "101", nil),
		"negative base": item("-1", "3", nil),
		"payroll form":  func() SubledgerWithholding { w := item("100", "3", nil); w.FormType = "PND1"; return w }(),
		"no partner":    func() SubledgerWithholding { w := item("100", "3", nil); w.PartnerCode = "NOPE"; return w }(),
	} {
		next := j
		next.Details = &JournalDetails{Withholdings: []SubledgerWithholding{bad}}
		if _, err := f.execute(f.scope, Command{Resource: "journals", Action: "update", ID: j.ID, Version: j.Version, Journal: &next}); err == nil {
			t.Fatalf("%s: expected rejection", name)
		}
	}
	expect(j.ID, "90000.5", "2700.02")

	// 4) ผ่านบัญชีแล้วยังแก้ฐานได้ (reconcile) — ยอด GL ไม่เปลี่ยน, audit เก็บค่าก่อน/หลัง
	j = f.post(j)
	noReason := subledgerReconcile(f, j, JournalDetails{Withholdings: []SubledgerWithholding{item("95000", "3", nil)}})
	noReason.Reason = "  "
	if _, err := f.execute(f.scope, noReason); err == nil {
		t.Fatal("tax-base edit after posting without a reason must be rejected")
	}
	expect(j.ID, "90000.5", "2700.02")
	f.run(subledgerReconcile(f, j, JournalDetails{Withholdings: []SubledgerWithholding{item("95000", "3", nil)}}))
	expect(j.ID, "95000", "2850")
	var before, reason string
	if err := f.db.QueryRow(`SELECT payload->'before'->0->>'base_amount', payload->>'reason' FROM gl_subledger_audit WHERE company='C' AND journal_id=$1 AND action='withholding_replace'`, j.ID).Scan(&before, &reason); err != nil || before != "90000.5" || reason != "reconcile evidence" {
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
