//go:build integration

package generalledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func (f *pgIntegrityFixture) inTx(fn func(tx *sql.Tx)) {
	f.t.Helper()
	tx, err := f.db.BeginTx(f.ctx, nil)
	if err != nil {
		f.t.Fatal(err)
	}
	defer tx.Rollback()
	fn(tx)
	if err = tx.Commit(); err != nil {
		f.t.Fatal(err)
	}
}

func (f *pgIntegrityFixture) mappingCmd(action, id string, version int64, m Master) Command {
	m.Kind = "mappings"
	return Command{Resource: "mappings", Action: action, ID: id, Version: version, Master: &m}
}

func saleRules() []MappingRule {
	return []MappingRule{{AccountCode: "1000", Side: MappingSideDebit, Source: MappingSourceTotal}, {AccountCode: "4000", Side: MappingSideCredit, Source: MappingSourceGoods}, {AccountCode: "5000", Side: MappingSideCredit, Source: MappingSourceVat}}
}

// รูปแบบลงบัญชีของเอกสาร ERP: ประเภทเอกสารต้องอยู่ในทะเบียน, ซ้ำไม่ได้, สมุด/บัญชีต้องใช้ได้, และเดบิต = เครดิตเสมอ
func TestMappingSourceDocValidationIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	good := Master{Code: "SALE", Name: "ลงบัญชีขายสินค้า", SourceDoc: "sale-invoice", BookCode: "SV", IsActive: true, Rules: saleRules()}

	created := f.run(f.mappingCmd("create", "", 0, good))
	raw, err := f.store.Get(f.ctx, f.scope, "mappings", created.ID)
	var stored Master
	if err != nil || json.Unmarshal(raw, &stored) != nil || stored.SourceDoc != "sale-invoice" || stored.BookCode != "SV" {
		t.Fatalf("stored mapping = %s (err %v), want sourcedoc and book kept", raw, err)
	}

	bad := func(mut func(m *Master), code, field string) *UserError {
		m := good
		m.Code = "OTHER"
		m.Rules = append([]MappingRule(nil), good.Rules...)
		mut(&m)
		return f.failCode(f.mappingCmd("create", "", 0, m), code, field)
	}
	bad(func(m *Master) { m.SourceDoc = "sale-order" }, "mapping_source_doc_invalid", "sourcedoc")
	bad(func(m *Master) { m.SourceDoc = "" }, "mapping_source_doc_invalid", "sourcedoc")
	if dup := bad(func(m *Master) {}, "mapping_source_doc_duplicate", "sourcedoc"); len(dup.Args) != 1 || dup.Args[0] != "SALE" {
		t.Fatalf("duplicate args = %v, want the code of the existing pattern", dup.Args)
	}
	bad(func(m *Master) { m.SourceDoc = "purchase"; m.BookCode = "NOPE" }, "mapping_book_invalid", "bookcode")
	bad(func(m *Master) { m.SourceDoc = "purchase"; m.Rules = m.Rules[:1] }, "mapping_rules_required", "rules")
	bad(func(m *Master) { m.SourceDoc = "purchase"; m.Rules[1].Side = "left" }, "mapping_rule_side_invalid", "rules")
	bad(func(m *Master) { m.SourceDoc = "purchase"; m.Rules[2].Source = "net" }, "mapping_rule_source_invalid", "rules")
	if acc := bad(func(m *Master) { m.SourceDoc = "purchase"; m.Rules[1].AccountCode = "9999" }, "mapping_account_invalid", "rules"); len(acc.Args) != 1 || acc.Args[0] != "9999" {
		t.Fatalf("account args = %v", acc.Args)
	}
	bad(func(m *Master) { m.SourceDoc = "purchase"; m.Rules = m.Rules[:2] }, "mapping_unbalanced", "rules")

	// แก้รูปแบบของตัวเองต้องไม่ชนกับตัวเอง; เปลี่ยนไปประเภทที่มีอยู่แล้วต้องชน
	other := good
	other.Code, other.SourceDoc = "BUY", "purchase"
	other.Rules = []MappingRule{{AccountCode: "5000", Side: MappingSideDebit, Source: MappingSourceGoods}, {AccountCode: "1000", Side: MappingSideDebit, Source: MappingSourceVat}, {AccountCode: "4000", Side: MappingSideCredit, Source: MappingSourceTotal}}
	buy := f.run(f.mappingCmd("create", "", 0, other))
	self := good
	self.Name = "ลงบัญชีขายสินค้า (แก้ไข)"
	f.run(f.mappingCmd("update", created.ID, created.Version, self))
	clash := good
	clash.SourceDoc = "purchase"
	f.failCode(f.mappingCmd("update", created.ID, created.Version+1, clash), "mapping_source_doc_duplicate", "sourcedoc")
	_ = buy
}

// ExecuteSourceTx: ใบสำคัญของเอกสาร ERP อยู่ในธุรกรรมเดียวกับเอกสาร — rollback แล้วไม่เหลือใบสำคัญ และผ่านกฎบัญชีครบ
func TestExecuteSourceTxCommitsAndRollsBackWithCallerIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	journal := func(doc, date string) *Journal {
		return &Journal{DocNo: doc, Date: date, BookCode: "SV", FiscalYear: date[:4], Description: "ขายสินค้า " + doc, Kind: "manual", BranchCode: "B1", SourceSystem: "erp", SourceRecordID: "sale-invoice:" + doc + ":r1",
			Lines: []Line{{AccountCode: "1000", Debit: Amount("107.00"), Credit: Amount("0")}, {AccountCode: "4000", Debit: Amount("0"), Credit: Amount("100.00")}, {AccountCode: "5000", Debit: Amount("0"), Credit: Amount("7.00")}}}
	}
	count := func(doc string) (n int) {
		if err := f.db.QueryRow(`SELECT count(*) FROM gl_records WHERE company='C' AND kind='journals' AND code=$1`, doc).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return
	}

	// rollback ของผู้เรียก → ไม่มีใบสำคัญ
	tx, _ := f.db.BeginTx(f.ctx, nil)
	if _, err := ExecuteSourceTx(f.ctx, tx, f.scope, Command{Resource: "journals", Action: "create", RequestID: uuid.NewString(), Journal: journal("INV-RB", "2026-03-01")}); err != nil {
		t.Fatal(err)
	}
	tx.Rollback()
	if n := count("INV-RB"); n != 0 {
		t.Fatalf("rolled-back voucher left %d rows", n)
	}

	// commit → create แล้ว post ในธุรกรรมเดียว; request id เดิมซ้ำ = ผลเดิม ไม่ซ้ำใบสำคัญ
	rid := uuid.NewString()
	var created Result
	f.inTx(func(tx *sql.Tx) {
		var err error
		if created, err = ExecuteSourceTx(f.ctx, tx, f.scope, Command{Resource: "journals", Action: "create", RequestID: rid, Journal: journal("INV-1", "2026-03-01")}); err != nil {
			t.Fatal(err)
		}
		if _, err = ExecuteSourceTx(f.ctx, tx, f.scope, Command{Resource: "journals", Action: "post", RequestID: uuid.NewString(), ID: created.ID, Version: created.Version}); err != nil {
			t.Fatal(err)
		}
	})
	if j := f.journal(created.ID); j.Status != "posted" {
		t.Fatalf("status = %s, want posted", j.Status)
	}
	f.inTx(func(tx *sql.Tx) {
		again, err := ExecuteSourceTx(f.ctx, tx, f.scope, Command{Resource: "journals", Action: "create", RequestID: rid, Journal: journal("INV-1", "2026-03-01")})
		if err != nil || again.ID != created.ID {
			t.Fatalf("replay = %+v (%v), want same journal", again, err)
		}
	})
	if n := count("INV-1"); n != 1 {
		t.Fatalf("INV-1 rows = %d, want 1", n)
	}

	// ไม่ใช่ command ของ source → ปฏิเสธ; request id สั้น → ปฏิเสธ
	tx, _ = f.db.BeginTx(f.ctx, nil)
	defer tx.Rollback()
	for _, cmd := range []Command{
		{Resource: "accounts", Action: "create", RequestID: uuid.NewString()},
		{Resource: "journals", Action: "delete", RequestID: uuid.NewString(), ID: created.ID},
		{Resource: "journals", Action: "create", RequestID: "short", Journal: journal("INV-X", "2026-03-01")},
	} {
		if _, err := ExecuteSourceTx(f.ctx, tx, f.scope, cmd); err == nil {
			t.Fatalf("ExecuteSourceTx accepted %s/%s", cmd.Resource, cmd.Action)
		}
	}
}

// ปีบัญชี/วันที่กลับรายการ: ไม่เดาปี, ปีปิดไม่รับ, ช่วงที่ล็อกกลับไปวันนี้ (ไม่ก่อนวันเดิม)
func TestSourceFiscalYearAndReversalDateIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	ctx := context.Background()
	f.inTx(func(tx *sql.Tx) {
		if code, err := SourceFiscalYear(ctx, tx, f.scope, "2026-05-10"); err != nil || code != "2026" {
			t.Fatalf("SourceFiscalYear = %q, %v", code, err)
		}
		if _, err := SourceFiscalYear(ctx, tx, f.scope, "2030-01-01"); !isCode(err, "source_fiscal_year_not_found") {
			t.Fatalf("no year: %v", err)
		}
		if _, err := SourceFiscalYear(ctx, tx, f.scope, "2026-13-40"); !isCode(err, "source_date_invalid") {
			t.Fatalf("bad date: %v", err)
		}
		// ช่วงเปิดอยู่ → กลับในวันเดิม
		if d, err := SourceReversalDate(ctx, tx, f.scope, "2026-05-10", "2026-06-01"); err != nil || d != "2026-05-10" {
			t.Fatalf("open period reversal date = %q, %v", d, err)
		}
	})

	// ปิดปี 2026 → กลับไปวันนี้ (ถ้าวันนี้อยู่ในปี 2027 ที่เปิด); วันนี้อยู่ในปีปิด → error ชัดเจน
	if _, err := f.db.Exec(`UPDATE gl_records SET payload=jsonb_set(payload,'{closed}','true') WHERE company='C' AND kind='fiscal-years' AND code='2026'`); err != nil {
		t.Fatal(err)
	}
	f.inTx(func(tx *sql.Tx) {
		if _, err := SourceFiscalYear(ctx, tx, f.scope, "2026-05-10"); !isCode(err, "source_fiscal_year_closed") {
			t.Fatalf("closed year: %v", err)
		}
		if d, err := SourceReversalDate(ctx, tx, f.scope, "2026-05-10", "2027-02-03"); err != nil || d != "2027-02-03" {
			t.Fatalf("closed-year reversal date = %q, %v", d, err)
		}
		if d, err := SourceReversalDate(ctx, tx, f.scope, "2027-05-10", "2027-02-03"); err != nil || d != "2027-05-10" {
			t.Fatalf("reversal never goes before the original: got %q, %v", d, err)
		}
		if _, err := SourceReversalDate(ctx, tx, f.scope, "2026-05-10", "2026-12-31"); !isCode(err, "source_fiscal_year_closed") {
			t.Fatalf("today inside the closed year: %v", err)
		}
	})
}

func isCode(err error, code string) bool {
	user, ok := AsUserError(err)
	return ok && user.Code == code
}
