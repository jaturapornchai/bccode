//go:build integration

package generalledger

import (
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf8"
)

// save-audit 2026-09-24 — round-trip บน PostgreSQL จริง ตรวจทีละขั้น:
// คู่ค้า (ชื่อไทยยาว เลขภาษีมีขีด สาขา "0") → ใบสำคัญที่มี VAT + ภาษีหัก → ผ่านบัญชี → แก้ทะเบียนคู่ค้า (เพิ่มบทบาท + แก้เลขภาษี)
// → รายละเอียดของใบเดิมไม่เปลี่ยน → ใบร่างที่ฝัง snapshot เก่ายังบันทึก/ผ่านรายการได้ → กระทบยอดล้างภาษีหักแถวสุดท้าย → audit ครบ
func TestPostgresPartnerSnapshotAndTaxRowRoundTrip(t *testing.T) {
	f := newPGIntegrityFixture(t)
	const code = "CUST-TH-001"
	if len(thaiPartnerName) <= 255 || utf8.RuneCountInString(thaiPartnerName) > 255 {
		t.Fatal("fixture name must exceed 255 bytes but not 255 runes")
	}
	master := func() (SubledgerPartner, int64) {
		t.Helper()
		var p SubledgerPartner
		var version int64
		var raw []byte
		if err := f.db.QueryRow(`SELECT payload, version FROM gl_subledger_partners WHERE company='C' AND code=$1`, code).Scan(&raw, &version); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(raw, &p); err != nil {
			t.Fatal(err)
		}
		return p, version
	}
	details := func(id string) string {
		t.Helper()
		var raw string
		if err := f.db.QueryRow(`SELECT COALESCE(payload->'details','null'::jsonb)::text FROM gl_records WHERE company='C' AND kind='journals' AND id=$1`, id).Scan(&raw); err != nil {
			t.Fatal(err)
		}
		return raw
	}
	detailValue := func(id, path string) string {
		t.Helper()
		var v string
		if err := f.db.QueryRow(`SELECT COALESCE(payload #>> $2::text[], '') FROM gl_records WHERE company='C' AND kind='journals' AND id=$1`, id, "{"+path+"}").Scan(&v); err != nil {
			t.Fatal(err)
		}
		return v
	}
	expectCode := func(name string, cmd Command, code string) {
		t.Helper()
		_, err := f.execute(f.scope, cmd)
		user, ok := AsUserError(err)
		if !ok || user.Code != code {
			t.Fatalf("%s: err = %v, want code %s", name, err, code)
		}
	}
	partner := SubledgerPartner{Code: code, Name: thaiPartnerName, TaxID: "0-1055-58012-34-9", TaxBranch: "0", Address: "88/12 ถนนลาดหลุมแก้ว ปทุมธานี 12140", IsCustomer: true, IsActive: true}
	lines := []Line{{AccountCode: "1000", Debit: "10700", Credit: "0"}, {AccountCode: "4000", Debit: "0", Credit: "10700"}}

	// 1) สร้างใบขาย: คู่ค้า + ลูกหนี้ + VAT ขาย + ลูกค้าหักภาษี ณ ที่จ่ายเรา 3% (ทิศทาง 2 — คู่ค้าเป็นผู้จ่าย)
	sale := Journal{DocNo: "SV6901-001", Date: "2026-01-10", BookCode: "JV", FiscalYear: "2026", Kind: "manual", BranchCode: "B1",
		Description: "ขายวัสดุก่อสร้างพร้อมค่าบริการขนส่ง ลูกค้าหักภาษี ณ ที่จ่าย 3%", Lines: lines,
		Details: &JournalDetails{
			Partners:    []SubledgerPartner{partner},
			Documents:   []SubledgerDocument{{ID: "AR-SV6901-001", Ledger: "ar", PartnerCode: code, DocumentNo: "IV6901-001", Date: "2026-01-10", BranchCode: "B1", Kind: 1, Side: 1, Amount: "10700", ControlAccountCode: "1000"}},
			Allocations: []SubledgerAllocation{{ID: "AL-SV6901-001", Ledger: "ar", DocumentID: "AR-SV6901-001", LineNumber: 1, Amount: "10700"}},
			Vats: []SubledgerVat{{ID: "V1", TaxType: 2, DocumentType: 1, TaxInvoiceNo: "IV6901-001", TaxInvoiceDate: "2026-01-10", TaxPeriodYear: 2026, TaxPeriodMonth: 1,
				PartnerCode: code, PartnerTaxID: "0-1055-58012-34-9", PartnerBranchNo: "0", PartnerName: thaiPartnerName,
				BaseAmount: "10000", ZeroRateAmount: "0", ExemptAmount: "0", Rate: "7"}},
			Withholdings: []SubledgerWithholding{{ID: "W1", Direction: 2, FormType: "PND53", PartnerCode: code, PaymentDate: "2026-01-10",
				IncomeType: "3_tres", Description: "ค่าบริการขนส่งวัสดุก่อสร้าง", Condition: 1, Rate: "3", BaseAmount: "10000",
				BookNo: "001", CertificateNo: "0012/2569", PayeeName: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", PayeeTaxID: "0-1055-61012-34-6", PayeeBranchNo: "0",
				Remark: "ลูกค้าส่งหนังสือรับรองทางไปรษณีย์"}},
		}}
	r := f.run(Command{Resource: "journals", Action: "create", Journal: &sale})
	p, version := master()
	if version != 1 || p.TaxID != "0105558012349" || p.TaxBranch != "00000" || p.Name != thaiPartnerName {
		t.Fatalf("step 1 PG partner = %+v v%d", p, version)
	}
	for path, want := range map[string]string{
		"details,vats,0,partner_tax_id":          "0105558012349",
		"details,vats,0,partner_branch_no":       "00000",
		"details,withholdings,0,payer_tax_id":    "0105558012349", // ฝั่งคู่ค้าว่าง → เก็บทะเบียน ณ วันบันทึก
		"details,withholdings,0,payer_branch_no": "00000",
		"details,withholdings,0,payer_name":      thaiPartnerName,
		"details,withholdings,0,payee_tax_id":    "0105561012346", // ฝั่งบริษัทที่ผู้ใช้กรอก ตัดขีดแล้ว
		"details,withholdings,0,payee_branch_no": "00000",
		"details,withholdings,0,wht_book_no":     "001",
		"details,withholdings,0,remark":          "ลูกค้าส่งหนังสือรับรองทางไปรษณีย์",
		"details,withholdings,0,tax_amount":      "300",
	} {
		if got := detailValue(r.ID, path); got != want {
			t.Fatalf("step 1 PG %s = %q, want %q", path, got, want)
		}
	}

	// 2) ใบร่างอีกใบฝัง snapshot คู่ค้าฉบับที่ 1 (ยังไม่ผ่านบัญชี) แล้วผ่านบัญชีใบขาย
	draft := Journal{DocNo: "SV6901-002", Date: "2026-01-11", BookCode: "JV", FiscalYear: "2026", Kind: "manual", BranchCode: "B1",
		Description: "ขายวัสดุก่อสร้าง รอตรวจเอกสาร", Lines: lines, Details: &JournalDetails{Partners: []SubledgerPartner{f.journal(r.ID).Details.Partners[0]}}}
	draftID := f.run(Command{Resource: "journals", Action: "create", Journal: &draft}).ID
	sale = f.post(f.journal(r.ID))
	postedDetails := details(sale.ID)

	// 3) แก้ทะเบียนคู่ค้าหลังมีเอกสาร: เพิ่มบทบาทเจ้าหนี้ + แก้เลขภาษีที่พิมพ์ผิด (โหลดฉบับที่ 1 ถูกต้อง)
	edited := f.journal(sale.ID).Details.Partners[0]
	edited.IsSupplier, edited.TaxID = true, "0105558012357"
	fix := Journal{DocNo: "JV6901-003", Date: "2026-01-12", BookCode: "JV", FiscalYear: "2026", Kind: "manual", BranchCode: "B1",
		Description: "ปรับปรุงทะเบียนคู่ค้า เพิ่มบทบาทเจ้าหนี้", Lines: lines, Details: &JournalDetails{Partners: []SubledgerPartner{edited}}}
	f.run(Command{Resource: "journals", Action: "create", Journal: &fix})
	p, version = master()
	if version != 2 || !p.IsSupplier || !p.IsCustomer || p.TaxID != "0105558012357" {
		t.Fatalf("step 3 PG partner = %+v v%d", p, version)
	}
	if got := details(sale.ID); got != postedDetails {
		t.Fatalf("step 3 posted journal details changed after partner edit:\nbefore %s\nafter  %s", postedDetails, got)
	}
	if detailValue(sale.ID, "details,vats,0,partner_tax_id") != "0105558012349" || detailValue(sale.ID, "details,withholdings,0,payer_tax_id") != "0105558012349" {
		t.Fatal("step 3 old VAT/WHT rows must keep the tax id recorded at the time")
	}
	// 50 ทวิ อ่านรายการที่บันทึก: snapshot เดิม + ทะเบียนคู่ค้าปัจจุบัน (ใช้เติมเฉพาะช่องที่ว่าง); ใบร่าง/ไม่มีรายการ = ไม่พบ
	item, live, err := RecordedWithholding(f.ctx, f.db, "C", sale.ID, "W1")
	if err != nil || item.PayerTaxID != "0105558012349" || item.BookNo != "001" || live.TaxID != "0105558012357" {
		t.Fatalf("step 3 recorded withholding = %+v partner %+v err=%v", item, live, err)
	}
	if _, _, err = RecordedWithholding(f.ctx, f.db, "C", draftID, "W1"); err != ErrNotFound {
		t.Fatalf("draft journal must not print a certificate: %v", err)
	}

	// 4) ใบร่างที่ฝัง snapshot ฉบับที่ 1 (ไม่ได้แก้คู่ค้า) ต้องแก้และผ่านรายการได้ — ไม่ทับทะเบียนด้วยข้อมูลเก่า
	d := f.journal(draftID)
	d.Description = "ขายวัสดุก่อสร้าง ตรวจเอกสารแล้ว"
	f.run(Command{Resource: "journals", Action: "update", ID: d.ID, Version: d.Version, Journal: &d})
	f.post(f.journal(draftID))
	if p, version = master(); version != 2 || p.TaxID != "0105558012357" || !p.IsSupplier {
		t.Fatalf("step 4 stale draft snapshot overwrote the register: %+v v%d", p, version)
	}

	// 5) optimistic lock ยังทำงาน: ผู้ใช้ถือฉบับที่ 1 แล้วแก้ข้อมูล ขณะที่ทะเบียนเป็นฉบับที่ 2
	stale := partner
	stale.Version, stale.Address = 1, "99 ถนนพหลโยธิน ปทุมธานี 12000"
	conflict := Journal{DocNo: "JV6901-004", Date: "2026-01-12", BookCode: "JV", FiscalYear: "2026", Kind: "manual", BranchCode: "B1", Description: "แก้ที่อยู่คู่ค้า", Lines: lines,
		Details: &JournalDetails{Partners: []SubledgerPartner{stale}}}
	expectCode("stale version edit", Command{Resource: "journals", Action: "create", Journal: &conflict}, "partner_version_conflict")
	if _, version = master(); version != 2 {
		t.Fatalf("step 5 register changed by rejected command: v%d", version)
	}

	// 6) พิมพ์รหัสเดิมเอง (ไม่ได้โหลดทะเบียน, version 0) พร้อมข้อมูลที่แก้ → บันทึกได้ เป็นฉบับที่ 3
	typed := p
	typed.Version, typed.Address = 0, "99 ถนนพหลโยธิน ปทุมธานี 12000"
	conflict.DocNo, conflict.Details = "JV6901-005", &JournalDetails{Partners: []SubledgerPartner{typed}}
	f.run(Command{Resource: "journals", Action: "create", Journal: &conflict})
	if p, version = master(); version != 3 || p.Address != "99 ถนนพหลโยธิน ปทุมธานี 12000" || p.TaxID != "0105558012357" {
		t.Fatalf("step 6 PG partner = %+v v%d", p, version)
	}

	// 7) ยกเลิกบทบาทลูกหนี้ทั้งที่มีเอกสารลูกหนี้แล้ว → ปฏิเสธพร้อมช่อง
	drop := p
	drop.IsCustomer = false
	conflict.DocNo, conflict.Details = "JV6901-006", &JournalDetails{Partners: []SubledgerPartner{drop}}
	expectCode("drop customer role", Command{Resource: "journals", Action: "create", Journal: &conflict}, "partner_customer_role_in_use")

	// 8) กระทบยอดใบที่ผ่านบัญชี: ไม่ส่งคีย์ = ไม่มีอะไรให้ทำ; ส่ง withholdings: [] = ล้างภาษีหักแถวสุดท้าย (VAT คงเดิม)
	sale = f.journal(sale.ID)
	if _, err := f.execute(f.scope, subledgerReconcile(f, sale, JournalDetails{})); err == nil || !strings.Contains(err.Error(), "ไม่มีรายละเอียดกระทบยอด") {
		t.Fatalf("absent keys must change nothing, got %v", err)
	}
	f.run(subledgerReconcile(f, sale, JournalDetails{Withholdings: []SubledgerWithholding{}}))
	if got := detailValue(sale.ID, "details,withholdings"); got != "" {
		t.Fatalf("step 8 withholdings still stored: %s", got)
	}
	if got := detailValue(sale.ID, "details,vats,0,partner_tax_id"); got != "0105558012349" {
		t.Fatalf("step 8 VAT row must stay when its key is absent, got %q", got)
	}
	var before, after, reason string
	if err := f.db.QueryRow(`SELECT payload->'before'->0->>'payer_tax_id', (payload->'after')::text, payload->>'reason' FROM gl_subledger_audit WHERE company='C' AND journal_id=$1 AND action='withholding_replace'`, sale.ID).Scan(&before, &after, &reason); err != nil || before != "0105558012349" || after != "[]" || reason != "reconcile evidence" {
		t.Fatalf("step 8 audit before=%q after=%q reason=%q err=%v", before, after, reason, err)
	}

	// 9) ล้าง VAT แถวสุดท้ายด้วยชุดว่างเช่นกัน; audit เป็น append-only ครบทุกขั้น
	sale = f.journal(sale.ID)
	f.run(subledgerReconcile(f, sale, JournalDetails{Vats: []SubledgerVat{}}))
	if got := detailValue(sale.ID, "details,vats"); got != "" {
		t.Fatalf("step 9 vats still stored: %s", got)
	}
	var actions string
	if err := f.db.QueryRow(`SELECT string_agg(action, ',' ORDER BY event_no) FROM gl_subledger_audit WHERE company='C' AND journal_id=$1`, sale.ID).Scan(&actions); err != nil {
		t.Fatal(err)
	}
	if actions != "create,post,withholding_replace,reconcile,vat_replace,reconcile" {
		t.Fatalf("step 9 audit trail = %s", actions)
	}
	var debit string
	if err := f.db.QueryRow(`SELECT SUM(debit)::text FROM gl_lines WHERE company='C' AND journal_id=$1`, sale.ID).Scan(&debit); err != nil || !strings.HasPrefix(debit, "10700") {
		t.Fatalf("GL after clearing tax rows = %q err=%v (must stay 10700)", debit, err)
	}
}

// ภาษีหักที่ไม่ส่ง wht_rate ต้องถูกปฏิเสธทั้งคำสั่ง ไม่มีใบถูกสร้าง (เดิมบันทึกเป็น 0% ภาษี 0)
func TestPostgresWithholdingWithoutRateRejected(t *testing.T) {
	f := newPGIntegrityFixture(t)
	j := Journal{DocNo: "PV6901-001", Date: "2026-01-10", BookCode: "JV", FiscalYear: "2026", Kind: "manual", BranchCode: "B1", Description: "จ่ายค่าขนส่ง หักภาษี ณ ที่จ่าย",
		Lines: []Line{{AccountCode: "5000", Debit: "10000", Credit: "0"}, {AccountCode: "1000", Debit: "0", Credit: "10000"}},
		Details: &JournalDetails{
			Partners:     []SubledgerPartner{{Code: "TRANS", Name: "บริษัท ขนส่งไทยเร็ว จำกัด", IsSupplier: true, IsActive: true}},
			Withholdings: []SubledgerWithholding{{ID: "W1", Direction: 1, FormType: "PND53", PartnerCode: "TRANS", PaymentDate: "2026-01-10", IncomeType: "3_tres", Condition: 1, BaseAmount: "10000"}},
		}}
	_, err := f.execute(f.scope, Command{Resource: "journals", Action: "create", Journal: &j})
	if user, ok := AsUserError(err); !ok || user.Code != "wht_rate_required" || user.Field != "wht_rate" {
		t.Fatalf("missing wht_rate: %v", err)
	}
	var count int
	if err := f.db.QueryRow(`SELECT COUNT(*) FROM gl_records WHERE company='C' AND kind='journals' AND payload->>'docno'='PV6901-001'`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("rejected journal stored: count=%d err=%v", count, err)
	}
}

// ข้อความไทยยาว (เหตุผลการถอน 200 ตัว, คำอธิบาย Statement 300 ตัว) ต้องบันทึกได้ — เดิมนับไบต์จึงเกิน 500
func TestPostgresThaiTextLengthsCountedInRunes(t *testing.T) {
	f, _, payment := subledgerFixtures(t)
	payment = f.post(payment)
	reason := strings.TrimSpace(strings.Repeat("ถอนการตัดยอดเพราะบันทึกผิดใบ ", 7))
	if n := utf8.RuneCountInString(reason); n > 500 || len(reason) <= 500 {
		t.Fatalf("reason fixture must exceed 500 bytes within 500 runes: %d runes %d bytes", n, len(reason))
	}
	f.run(subledgerReconcile(f, payment, JournalDetails{Withdrawals: []SubledgerWithdrawal{{Kind: "settlement", ID: "SET1", Reason: reason}}}))
	var stored string
	if err := f.db.QueryRow(`SELECT reversal_reason FROM gl_subledger_settlements WHERE company='C' AND id='SET1'`).Scan(&stored); err != nil || stored != reason {
		t.Fatalf("withdraw reason stored = %d runes err=%v", utf8.RuneCountInString(stored), err)
	}
	description := strings.TrimSpace(strings.Repeat("รับโอนเงินจากลูกค้าโครงการก่อสร้าง ", 10))
	if n := utf8.RuneCountInString(description); n > 500 || len(description) <= 500 {
		t.Fatalf("description fixture must exceed 500 bytes within 500 runes: %d runes %d bytes", n, len(description))
	}
	payment = f.journal(payment.ID)
	f.run(subledgerReconcile(f, payment, JournalDetails{StatementLines: []SubledgerStatementLine{{ID: "S-TH", BankAccountCode: "BANK", SourceKey: "KBANK-2569-01-11-ลำดับ-0003", Date: "2026-01-11", Direction: 1, Amount: "0.05", Description: description}}}))
	if err := f.db.QueryRow(`SELECT payload->>'description' FROM gl_subledger_statements WHERE company='C' AND id='S-TH'`).Scan(&stored); err != nil || stored != description {
		t.Fatalf("statement description stored = %d runes err=%v", utf8.RuneCountInString(stored), err)
	}
}
