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

// แถวภาษีหักเดิมที่เปลี่ยนคู่ค้า/ทิศทาง ต้องไม่พก snapshot ของคู่ค้าเดิม (เดิม ภ.ง.ด./50 ทวิ ออกในนามคู่ค้าเดิม — adversarial review 2026-09-24)
func TestPostgresWithholdingPartnerChangeRefreshesSnapshot(t *testing.T) {
	f := newPGIntegrityFixture(t)
	snapshot := func(id string) SubledgerWithholding {
		t.Helper()
		j := f.journal(id)
		if j.Details == nil || len(j.Details.Withholdings) != 1 {
			t.Fatalf("stored withholdings = %+v", j.Details)
		}
		return j.Details.Withholdings[0]
	}
	j := Journal{DocNo: "PV6901-010", Date: "2026-01-10", BookCode: "JV", FiscalYear: "2026", Kind: "manual", BranchCode: "B1", Description: "จ่ายค่าขนส่ง หักภาษี ณ ที่จ่าย 1%",
		Lines: []Line{{AccountCode: "5000", Debit: "10000", Credit: "0"}, {AccountCode: "1000", Debit: "0", Credit: "10000"}},
		Details: &JournalDetails{
			Partners: []SubledgerPartner{
				{Code: "S001", Name: "บริษัท ขนส่งไทยเร็ว จำกัด", TaxID: "0105558012349", IsSupplier: true, IsActive: true},
				{Code: "S002", Name: "ห้างหุ้นส่วนจำกัด รุ่งเรืองขนส่ง", TaxID: "0103561234568", IsSupplier: true, IsActive: true},
			},
			Withholdings: []SubledgerWithholding{{ID: "W1", Direction: 1, FormType: "PND53", PartnerCode: "S001", PaymentDate: "2026-01-10", IncomeType: "3_tres", Condition: 1, Rate: "1", BaseAmount: "10000"}},
		}}
	r := f.run(Command{Resource: "journals", Action: "create", Journal: &j})
	if w := snapshot(r.ID); w.PayeeTaxID != "0105558012349" || w.PayeeName != "บริษัท ขนส่งไทยเร็ว จำกัด" {
		t.Fatalf("step 1 payee snapshot = %+v", w)
	}

	// 1) ใบร่าง: เปลี่ยนคู่ค้าเป็น S002 โดยจอส่ง payee_* ของ S001 กลับมาเหมือนเดิม → ต้องได้ข้อมูล S002
	j = f.journal(r.ID)
	j.Details.Withholdings[0].PartnerCode = "S002"
	f.run(Command{Resource: "journals", Action: "update", ID: j.ID, Version: j.Version, Journal: &j})
	if w := snapshot(j.ID); w.PayeeTaxID != "0103561234568" || w.PayeeName != "ห้างหุ้นส่วนจำกัด รุ่งเรืองขนส่ง" {
		t.Fatalf("step 1 partner change kept old payee: %+v", w)
	}

	// 2) ผู้ใช้พิมพ์ชื่อผู้รับเงินเองพร้อมเปลี่ยนคู่ค้า → ค่าที่พิมพ์ใหม่ต้องอยู่ ไม่ถูกทะเบียนทับ
	j = f.journal(j.ID)
	j.Details.Withholdings[0].PartnerCode = "S001"
	j.Details.Withholdings[0].PayeeName = "บริษัท ขนส่งไทยเร็ว จำกัด (สาขาบางนา)"
	j.Details.Withholdings[0].PayeeTaxID = "0105558012349"
	j.Details.Withholdings[0].PayeeAddress = ""
	f.run(Command{Resource: "journals", Action: "update", ID: j.ID, Version: j.Version, Journal: &j})
	if w := snapshot(j.ID); w.PayeeName != "บริษัท ขนส่งไทยเร็ว จำกัด (สาขาบางนา)" || w.PayeeTaxID != "0105558012349" {
		t.Fatalf("step 2 typed payee overwritten: %+v", w)
	}

	// 3) ผ่านบัญชีแล้ว reconcile เปลี่ยนคู่ค้ากลับเป็น S002 (ส่ง snapshot เดิมมาด้วย) → ข้อมูล S002
	j = f.post(f.journal(j.ID))
	row := snapshot(j.ID)
	row.PartnerCode = "S002"
	f.run(subledgerReconcile(f, j, JournalDetails{Withholdings: []SubledgerWithholding{row}}))
	if w := snapshot(j.ID); w.PayeeTaxID != "0103561234568" || w.PayeeName != "ห้างหุ้นส่วนจำกัด รุ่งเรืองขนส่ง" {
		t.Fatalf("step 3 reconcile partner change kept old payee: %+v", w)
	}

	// 4) เปลี่ยนทิศทาง 1→2: คู่ค้าเป็นผู้จ่าย ฝั่งผู้รับ (บริษัท) ต้องไม่ใช่ข้อมูลคู่ค้า
	j = f.journal(j.ID)
	row = snapshot(j.ID)
	row.Direction = 2
	f.run(subledgerReconcile(f, j, JournalDetails{Withholdings: []SubledgerWithholding{row}}))
	w := snapshot(j.ID)
	if w.PayerTaxID != "0103561234568" || w.PayerName != "ห้างหุ้นส่วนจำกัด รุ่งเรืองขนส่ง" {
		t.Fatalf("step 4 payer (partner) snapshot = %+v", w)
	}
	if !w.Payee().blank() {
		t.Fatalf("step 4 payee (company side) kept partner data: %+v", w.Payee())
	}
}
