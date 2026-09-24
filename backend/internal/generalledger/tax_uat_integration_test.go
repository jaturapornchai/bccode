//go:build integration

package generalledger

import (
	"errors"
	"testing"
)

// UAT V13 2026-09-24: ไม่ส่ง vat_rate ต้องถูกปฏิเสธทั้งตอนสร้างใบและตอนกระทบยอด และ PostgreSQL ต้องไม่มีอะไรถูกบันทึก
// (เดิม JSON round trip แปลงค่าว่างเป็น "0" → เก็บเป็นภาษี 0% เงียบ ๆ)
func TestPostgresVatRateRequiredOnCreateAndReconcile(t *testing.T) {
	f := newPGIntegrityFixture(t)
	wantRateError := func(stage string, err error) {
		t.Helper()
		user, ok := AsUserError(err)
		if !ok || user.Code != "vat_rate_invalid" || user.Field != "vat_rate" {
			t.Fatalf("%s: err = %#v, want vat_rate_invalid", stage, err)
		}
	}
	noRate := func() SubledgerVat {
		v := vatItem("V1", 2, 1, "IV2610-001", "2026-01-10", 2026, 1, 0, "1000.00", nil)
		v.Rate = ""
		return v
	}

	j := Journal{DocNo: "UV-NORATE", Date: "2026-01-10", BookCode: "UV", FiscalYear: "2026", Description: "ขายวัสดุก่อสร้าง เป็นเงินเชื่อ", Kind: "manual", BranchCode: "B1",
		Lines:   []Line{{AccountCode: "1000", Debit: "1070", Credit: "0"}, {AccountCode: "4000", Debit: "0", Credit: "1070"}},
		Details: &JournalDetails{Vats: []SubledgerVat{noRate()}}}
	_, err := f.execute(f.scope, Command{Resource: "journals", Action: "create", Journal: &j})
	wantRateError("create", err)
	var count int
	if err := f.db.QueryRow(`SELECT COUNT(*) FROM gl_records WHERE company='C' AND kind='journals' AND code='UV-NORATE'`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("rejected journal stored: count=%d err=%v", count, err)
	}

	posted := f.post(f.draft("UV-POSTED", "B1"))
	_, err = f.execute(f.scope, subledgerReconcile(f, posted, JournalDetails{Vats: []SubledgerVat{noRate()}}))
	wantRateError("reconcile", err)
	var vats string
	if err := f.db.QueryRow(`SELECT COALESCE(payload->'details'->'vats','[]'::jsonb)::text FROM gl_records WHERE company='C' AND kind='journals' AND id=$1`, posted.ID).Scan(&vats); err != nil || vats != "[]" {
		t.Fatalf("reconcile stored vats = %s err=%v", vats, err)
	}
}

// UAT S09 2026-09-24: กระทบยอดใบร่างต้องบอกเหตุผล ไม่ใช่ "ไม่พบรายการ" (404)
func TestPostgresReconcileDraftSaysPostFirst(t *testing.T) {
	f := newPGIntegrityFixture(t)
	draft := f.draft("JV-DRAFT", "B1")
	_, err := f.execute(f.scope, subledgerReconcile(f, draft, JournalDetails{Vats: []SubledgerVat{vatItem("V1", 2, 1, "IV2601-009", "2026-01-10", 2026, 1, 0, "100", nil)}}))
	if errors.Is(err, ErrNotFound) {
		t.Fatalf("draft reconcile still reports not found: %v", err)
	}
	if user, ok := AsUserError(err); !ok || user.Code != "reconcile_requires_posted" {
		t.Fatalf("draft reconcile err = %#v", err)
	}
}

// UAT S08/V14 2026-09-24: กลับรายการที่ไม่ผ่านต้องบอกสาเหตุจริงทีละข้อ (เดิมข้อความเดียวรวม 5 สาเหตุ)
func TestPostgresReverseErrorsNameTheCause(t *testing.T) {
	f := newPGIntegrityFixture(t)
	original := f.post(f.draft("JV-REV", "B1"))
	reverse := func(j Journal, doc, date, reason string) error {
		_, err := f.execute(f.scope, Command{Resource: "journals", Action: "reverse", ID: j.ID, Version: j.Version, DocNo: doc, Date: date, Reason: reason})
		return err
	}
	for _, tc := range []struct {
		name, doc, date, reason, code, field string
	}{
		{"no reason", "JV-REV-R", "2026-01-11", " ", "reverse_reason_required", "reason"},
		{"date before original", "JV-REV-R", "2026-01-09", "บันทึกผิดบัญชี", "reverse_date_before_original", "date"},
		{"bad doc no", "", "2026-01-11", "บันทึกผิดบัญชี", "reverse_doc_no_invalid", "docno"},
		{"bad date", "JV-REV-R", "11/01/2569", "บันทึกผิดบัญชี", "reverse_date_invalid", "date"},
	} {
		user, ok := AsUserError(reverse(original, tc.doc, tc.date, tc.reason))
		if !ok || user.Code != tc.code || user.Field != tc.field {
			t.Fatalf("%s: err = %#v, want %s/%s", tc.name, user, tc.code, tc.field)
		}
	}
	f.run(Command{Resource: "journals", Action: "reverse", ID: original.ID, Version: original.Version, DocNo: "JV-REV-R", Date: "2026-01-11", Reason: "บันทึกผิดบัญชี"})
	var reversalID string
	if err := f.db.QueryRow(`SELECT id FROM gl_records WHERE company='C' AND kind='journals' AND code='JV-REV-R'`).Scan(&reversalID); err != nil {
		t.Fatal(err)
	}
	if user, ok := AsUserError(reverse(f.journal(reversalID), "JV-REV-R2", "2026-01-12", "กลับซ้ำ")); !ok || user.Code != "reverse_kind_not_allowed" {
		t.Fatalf("reverse of reversal = %#v", user)
	}
	if user, ok := AsUserError(reverse(f.journal(original.ID), "JV-REV-R3", "2026-01-12", "กลับซ้ำ")); !ok || user.Code != "reverse_requires_posted" {
		t.Fatalf("reverse of reversed original = %#v", user)
	}
}
