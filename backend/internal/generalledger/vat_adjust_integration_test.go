//go:build integration

package generalledger

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// Champ 5498 "ปรับปรุงภาษีซื้อ": list purchase VAT rows of posted journals and change claim period/status of many rows in one command.
type vatAdjustFixture struct {
	*pgIntegrityFixture
	j1, j2, j3, draftJournal Journal
}

func newVatAdjustFixture(t *testing.T) *vatAdjustFixture {
	f := &vatAdjustFixture{pgIntegrityFixture: newPGIntegrityFixture(t)}
	create := func(doc, date string, partners []SubledgerPartner, vats ...SubledgerVat) Journal {
		t.Helper()
		j := Journal{DocNo: doc, Date: date, BookCode: "JV", FiscalYear: "2026", Description: "ซื้อวัสดุก่อสร้าง " + doc, Kind: "manual", BranchCode: "B1",
			Lines:   []Line{{AccountCode: "1000", Debit: "1070", Credit: "0"}, {AccountCode: "4000", Debit: "0", Credit: "1070"}},
			Details: &JournalDetails{Partners: partners, Vats: vats}}
		return f.journal(f.run(Command{Resource: "journals", Action: "create", Journal: &j}).ID)
	}
	supplier := func(code string) SubledgerPartner {
		return SubledgerPartner{Code: code, Name: "บริษัท สยามพาณิชย์ โฮลดิ้ง จำกัด " + code, TaxID: "0105558012349", IsSupplier: true, IsActive: true}
	}
	withPartner := func(v SubledgerVat, code string) SubledgerVat { v.PartnerCode = code; return v }
	// J1: purchase P1 (invoice Jan, claimed in Mar), P2 (invoice Feb, claim pending, no period)
	f.j1 = f.post(create("AJ-1", "2026-01-10", []SubledgerPartner{supplier("AP001")},
		withPartner(vatItem("P1", 1, 1, "PI-1", "2026-01-10", 2026, 3, 1, "1000", nil), "AP001"),
		vatItem("P2", 1, 1, "PI-2", "2026-02-03", 0, 0, 3, "2000", nil)))
	// J2: purchase P3 (invoice Jan, claimed in Apr) + a sales row S1 that must never be listed or changed
	f.j2 = f.post(create("AJ-2", "2026-01-20", []SubledgerPartner{supplier("AP002")},
		withPartner(vatItem("P3", 1, 1, "PI-3", "2026-01-20", 2026, 4, 1, "3000", nil), "AP002"),
		vatItem("S1", 2, 1, "SV-1", "2026-01-20", 2026, 1, 0, "500", nil)))
	// J3: purchase P4 (invoice Mar, claimed in May) — outside the January filter of the "invoice date" tests
	f.j3 = f.post(create("AJ-3", "2026-03-05", nil, vatItem("P4", 1, 1, "PI-4", "2026-03-05", 2026, 5, 1, "4000", nil)))
	// A draft journal and another company's journal must not be listed
	f.draftJournal = create("AJ-DRAFT", "2026-01-12", nil, vatItem("P5", 1, 1, "PI-5", "2026-01-12", 2026, 1, 1, "100", nil))
	if _, err := f.db.Exec(`INSERT INTO gl_records(company,kind,id,code,version,payload) SELECT 'OTHER',kind,id,code,version,payload FROM gl_records WHERE company='C' AND id=$1`, f.j1.ID); err != nil {
		t.Fatal(err)
	}
	return f
}

func (f *vatAdjustFixture) list(filter VatAdjustFilter, page, limit int) (Page, []VatAdjustRow) {
	f.t.Helper()
	p, err := f.store.pg.VatAdjustList(context.Background(), f.scope, filter, page, limit)
	if err != nil {
		f.t.Fatalf("VatAdjustList: %v", err)
	}
	rows := []VatAdjustRow{}
	for _, raw := range p.Items {
		var row VatAdjustRow
		if err := json.Unmarshal(raw, &row); err != nil {
			f.t.Fatal(err)
		}
		rows = append(rows, row)
	}
	return p, rows
}

func vatAdjustIDs(rows []VatAdjustRow) string {
	out := ""
	for _, r := range rows {
		out += r.ID + ","
	}
	return out
}

// stored returns "year-month/claim" of one VAT row as saved in PostgreSQL, e.g. "2026-3/1".
func (f *vatAdjustFixture) stored(journalID, vatID string) string {
	f.t.Helper()
	var raw []byte
	if err := f.db.QueryRow(`SELECT item FROM gl_records r, jsonb_array_elements(r.payload->'details'->'vats') item WHERE r.company='C' AND r.kind='journals' AND r.id=$1 AND item->>'id'=$2`, journalID, vatID).Scan(&raw); err != nil {
		f.t.Fatalf("read %s: %v", vatID, err)
	}
	var v SubledgerVat
	if err := json.Unmarshal(raw, &v); err != nil {
		f.t.Fatal(err)
	}
	return fmt.Sprintf("%d-%d/%d", v.TaxPeriodYear, v.TaxPeriodMonth, v.ClaimStatus)
}

func (f *vatAdjustFixture) version(journalID string) int64 { return f.journal(journalID).Version }

func (f *vatAdjustFixture) rowsCommand(reason string, items ...VatAdjustItem) Command {
	return Command{Resource: VatAdjustKind, Action: "update", Reason: reason, VatAdjust: &VatAdjustInput{Mode: VatAdjustModeRows, Items: items}}
}

func (f *vatAdjustFixture) invoiceDateCommand(filter VatAdjustFilter, expected int) Command {
	return Command{Resource: VatAdjustKind, Action: "update", Reason: "ตั้งงวดตามวันที่ใบกำกับภาษี", VatAdjust: &VatAdjustInput{Mode: VatAdjustModeInvoiceDate, Filter: filter, ExpectedCount: expected}}
}

func TestVatAdjustListFilters(t *testing.T) {
	f := newVatAdjustFixture(t)
	jan := VatAdjustFilter{InvDateFrom: "2026-01-01", InvDateTo: "2026-01-31"}
	page, rows := f.list(jan, 1, 100)
	// purchase rows only (S1 is output VAT), posted only (draft P5), own company only, ordered by invoice date then document no
	if got := vatAdjustIDs(rows); got != "P1,P3," || page.Total != 2 {
		t.Fatalf("january = %s total %d", got, page.Total)
	}
	if rows[0].JournalID != f.j1.ID || rows[0].Version != f.j1.Version || rows[0].DocNo != "AJ-1" || rows[0].DocDate != "2026-01-10" || string(rows[0].BaseAmount) != "1000" || rows[0].TaxPeriodMonth != 3 {
		t.Fatalf("row 0 = %+v", rows[0])
	}
	// paging with a fixed order: LIMIT/OFFSET
	if _, second := f.list(jan, 2, 1); vatAdjustIDs(second) != "P3," {
		t.Fatalf("page 2 = %s", vatAdjustIDs(second))
	}
	// document date, and Champ "from only = exact": doc date 2026-01-20 is journal AJ-2
	if _, byDoc := f.list(VatAdjustFilter{DocDateFrom: "2026-01-20"}, 1, 100); vatAdjustIDs(byDoc) != "P3," {
		t.Fatalf("doc date exact = %s", vatAdjustIDs(byDoc))
	}
	// creditor code: from-only is an exact match, from-to a range
	if _, one := f.list(VatAdjustFilter{InvDateFrom: "2026-01-01", InvDateTo: "2026-12-31", PartnerFrom: "AP001"}, 1, 100); vatAdjustIDs(one) != "P1," {
		t.Fatalf("partner exact = %s", vatAdjustIDs(one))
	}
	if _, both := f.list(VatAdjustFilter{InvDateFrom: "2026-01-01", InvDateTo: "2026-12-31", PartnerFrom: "AP001", PartnerTo: "AP002"}, 1, 100); vatAdjustIDs(both) != "P1,P3," {
		t.Fatalf("partner range = %s", vatAdjustIDs(both))
	}
	// all purchase rows of the year, claim-pending row included
	if all, rows := f.list(VatAdjustFilter{InvDateTo: "2026-12-31"}, 1, 100); vatAdjustIDs(rows) != "P1,P3,P2,P4," || all.Total != 4 {
		t.Fatalf("year = %s", vatAdjustIDs(rows))
	}
	// a branch-scoped user sees only that branch
	other := f.scope
	other.Branch = "B2"
	if p, err := f.store.pg.VatAdjustList(context.Background(), other, jan, 1, 100); err != nil || p.Total != 0 {
		t.Fatalf("other branch total %d err %v", p.Total, err)
	}
	// Champ refuses to load without a date; "to" must not precede "from"; dates must be real dates
	for name, bad := range map[string]struct {
		filter VatAdjustFilter
		code   string
	}{
		"no date":       {VatAdjustFilter{PartnerFrom: "AP001"}, "vat_adjust_filter_required"},
		"empty":         {VatAdjustFilter{}, "vat_adjust_filter_required"},
		"reversed date": {VatAdjustFilter{InvDateFrom: "2026-02-01", InvDateTo: "2026-01-01"}, "vat_adjust_range_invalid"},
		"reversed code": {VatAdjustFilter{InvDateFrom: "2026-01-01", PartnerFrom: "AP9", PartnerTo: "AP1"}, "vat_adjust_range_invalid"},
		"bad date":      {VatAdjustFilter{DocDateFrom: "2026-02-30"}, "vat_adjust_date_invalid"},
	} {
		_, err := f.store.pg.VatAdjustList(context.Background(), f.scope, bad.filter, 1, 100)
		if user, ok := AsUserError(err); !ok || user.Code != bad.code {
			t.Fatalf("%s: err = %v, want %s", name, err, bad.code)
		}
	}
}

func TestVatAdjustRowsUpdatesManyRowsInOneTransaction(t *testing.T) {
	f := newVatAdjustFixture(t)
	v1, v2 := f.version(f.j1.ID), f.version(f.j2.ID)
	reason := "ย้ายงวดตามใบกำกับภาษีที่ได้รับ"
	// P1: Mar -> Feb (still claimed), P2: pending -> not claimed, P3: Apr -> Feb — three rows in two journals
	command := f.rowsCommand(reason,
		VatAdjustItem{JournalID: f.j1.ID, Version: v1, VatID: "P1", TaxPeriodYear: 2026, TaxPeriodMonth: 2, ClaimStatus: 1},
		VatAdjustItem{JournalID: f.j1.ID, Version: v1, VatID: "P2", TaxPeriodYear: 0, TaxPeriodMonth: 0, ClaimStatus: 4},
		VatAdjustItem{JournalID: f.j2.ID, Version: v2, VatID: "P3", TaxPeriodYear: 2026, TaxPeriodMonth: 2, ClaimStatus: 1})
	command.RequestID = "11111111-1111-4111-8111-111111111111"
	result := f.run(command)
	if result.AdjustedJournals != 2 {
		t.Fatalf("adjusted journals = %d", result.AdjustedJournals)
	}
	for _, want := range []struct{ journal, vat, state string }{
		{f.j1.ID, "P1", "2026-2/1"}, {f.j1.ID, "P2", "0-0/4"}, {f.j2.ID, "P3", "2026-2/1"}, {f.j3.ID, "P4", "2026-5/1"},
		{f.j2.ID, "S1", "2026-1/0"}, // the sales row of J2 is untouched
	} {
		if got := f.stored(want.journal, want.vat); got != want.state {
			t.Fatalf("%s = %s, want %s", want.vat, got, want.state)
		}
	}
	if f.version(f.j1.ID) != v1+1 || f.version(f.j2.ID) != v2+1 || f.version(f.j3.ID) != f.j3.Version {
		t.Fatalf("versions = %d %d %d", f.version(f.j1.ID), f.version(f.j2.ID), f.version(f.j3.ID))
	}
	// audit: one vat_replace per changed journal with the reason and the rows before/after; one event holding both journals
	if n := f.countRows(`SELECT count(*) FROM gl_subledger_audit WHERE company='C' AND action='vat_replace' AND payload->>'reason'=$3 AND journal_id IN ($1,$2)`, f.j1.ID, f.j2.ID, reason); n != 2 {
		t.Fatalf("vat_replace audit rows = %d", n)
	}
	var beforePeriod, afterPeriod string
	if err := f.db.QueryRow(`SELECT payload->'before'->0->>'tax_period_month', payload->'after'->0->>'tax_period_month' FROM gl_subledger_audit WHERE company='C' AND action='vat_replace' AND journal_id=$1`, f.j1.ID).Scan(&beforePeriod, &afterPeriod); err != nil || beforePeriod != "3" || afterPeriod != "2" {
		t.Fatalf("audit before/after = %s/%s err %v", beforePeriod, afterPeriod, err)
	}
	if n := f.countRows(`SELECT count(*) FROM gl_events WHERE company='C' AND payload->>'action'='vat-adjustments:update' AND jsonb_array_length(payload->'changes')=2`); n != 1 {
		t.Fatalf("events = %d", n)
	}
	// the same request id answers from the saved event and writes nothing twice
	if again := f.run(command); again.AdjustedJournals != 2 {
		t.Fatalf("retry adjusted journals = %d", again.AdjustedJournals)
	}
	if n := f.countRows(`SELECT count(*) FROM gl_subledger_audit WHERE company='C' AND action='vat_replace'`); n != 2 {
		t.Fatalf("audit rows after retry = %d", n)
	}
	// the screen reads the new periods and versions back
	if _, rows := f.list(VatAdjustFilter{InvDateTo: "2026-12-31"}, 1, 100); rows[0].TaxPeriodMonth != 2 || rows[0].Version != v1+1 {
		t.Fatalf("list after save = %+v", rows[0])
	}
}

func TestVatAdjustRowsRejectionRollsBackEveryRow(t *testing.T) {
	f := newVatAdjustFixture(t)
	v1, v2 := f.version(f.j1.ID), f.version(f.j2.ID)
	// P1 -> Feb is fine, P3 (invoice January 2026) claimed in September 2026 is past the 6-month window (July)
	_, err := f.execute(f.scope, f.rowsCommand("test window",
		VatAdjustItem{JournalID: f.j1.ID, Version: v1, VatID: "P1", TaxPeriodYear: 2026, TaxPeriodMonth: 2, ClaimStatus: 1},
		VatAdjustItem{JournalID: f.j2.ID, Version: v2, VatID: "P3", TaxPeriodYear: 2026, TaxPeriodMonth: 9, ClaimStatus: 1}))
	user, ok := AsUserError(err)
	if !ok || user.Code != "vat_claim_window_exceeded" || user.Field != "tax_period_month" {
		t.Fatalf("window error = %#v", err)
	}
	if !strings.HasPrefix(user.Message, "ใบสำคัญ AJ-2: ") {
		t.Fatalf("message does not name the voucher: %s", user.Message)
	}
	if f.stored(f.j1.ID, "P1") != "2026-3/1" || f.stored(f.j2.ID, "P3") != "2026-4/1" || f.version(f.j1.ID) != v1 || f.version(f.j2.ID) != v2 {
		t.Fatal("a rejected command left a changed row behind")
	}
	if n := f.countRows(`SELECT count(*) FROM gl_subledger_audit WHERE company='C' AND action IN('vat_replace','reconcile')`); n != 0 {
		t.Fatalf("audit rows after rollback = %d", n)
	}
	if n := f.countRows(`SELECT count(*) FROM gl_events WHERE company='C' AND payload->>'action'='vat-adjustments:update'`); n != 0 {
		t.Fatalf("events after rollback = %d", n)
	}
}

func TestVatAdjustRowsVersionConflictAndInputRules(t *testing.T) {
	f := newVatAdjustFixture(t)
	v1, v2 := f.version(f.j1.ID), f.version(f.j2.ID)
	item := func(journal Journal, version int64, vat string, month int) VatAdjustItem {
		return VatAdjustItem{JournalID: journal.ID, Version: version, VatID: vat, TaxPeriodYear: 2026, TaxPeriodMonth: month, ClaimStatus: 1}
	}
	// somebody changed J2 after the screen loaded it: the whole command is refused, J1 stays as it was
	f.failWith(f.rowsCommand("test", item(f.j1, v1, "P1", 2), item(f.j2, v2-1, "P3", 2)), CodeStaleVersion)
	if f.stored(f.j1.ID, "P1") != "2026-3/1" || f.version(f.j1.ID) != v1 {
		t.Fatal("version conflict changed J1")
	}
	// the change reason is required, like the single-journal save
	f.failWith(f.rowsCommand("   ", item(f.j1, v1, "P1", 2)), "tax_edit_reason_required")
	// unknown row, and the sales row (not adjustable here)
	f.failWith(f.rowsCommand("test", item(f.j1, v1, "NOPE", 2)), "vat_adjust_row_not_found")
	f.failWith(f.rowsCommand("test", item(f.j2, v2, "S1", 2)), "vat_adjust_row_not_found")
	// empty, duplicated row, two versions of one journal, unknown mode
	f.failWith(f.rowsCommand("test"), "vat_adjust_rows_invalid")
	f.failWith(f.rowsCommand("test", item(f.j1, v1, "P1", 2), item(f.j1, v1, "P1", 2)), "vat_adjust_rows_invalid")
	f.failWith(f.rowsCommand("test", item(f.j1, v1, "P1", 2), item(f.j1, v1+1, "P2", 2)), "vat_adjust_rows_invalid")
	f.failWith(Command{Resource: VatAdjustKind, Action: "update", Reason: "test", VatAdjust: &VatAdjustInput{Mode: "all"}}, "vat_adjust_rows_invalid")
	// nothing differs from what is stored
	f.failWith(f.rowsCommand("test", item(f.j1, v1, "P1", 3)), "vat_adjust_nothing_changed")
	// more journals than one command may rewrite
	many := []VatAdjustItem{}
	for i := 0; i <= VatAdjustMaxJournals; i++ {
		many = append(many, VatAdjustItem{JournalID: fmt.Sprintf("J-%d", i), Version: 1, VatID: "P1", TaxPeriodYear: 2026, TaxPeriodMonth: 2, ClaimStatus: 1})
	}
	f.failWith(f.rowsCommand("test", many...), "vat_adjust_too_many")
	// a user bound to another branch cannot touch the journals of branch B1
	other := f.scope
	other.Branch = "B2"
	f.deny(other, f.rowsCommand("test", item(f.j1, v1, "P1", 2)))
	// a draft journal is not adjustable (the reconcile path needs a posted journal)
	f.failWith(f.rowsCommand("test", item(f.draftJournal, f.draftJournal.Version, "P5", 2)), "reconcile_requires_posted")
}

func TestVatAdjustInvoiceDateModeTouchesOnlyFilteredRows(t *testing.T) {
	f := newVatAdjustFixture(t)
	jan := VatAdjustFilter{InvDateFrom: "2026-01-01", InvDateTo: "2026-01-31"}
	// the screen showed 2 rows; a different count means the data changed, nothing is written
	f.failWith(f.invoiceDateCommand(jan, 3), "vat_adjust_filter_changed")
	f.failWith(f.invoiceDateCommand(VatAdjustFilter{PartnerFrom: "AP001"}, 1), "vat_adjust_filter_required")
	f.failWith(f.invoiceDateCommand(jan, 0), "vat_adjust_rows_invalid")
	if f.stored(f.j1.ID, "P1") != "2026-3/1" {
		t.Fatal("refused command changed a row")
	}
	// January invoices: P1 (Mar) and P3 (Apr) become January; P2 (Feb invoice, pending) and P4 (Mar invoice claimed in May) are outside the filter
	result := f.run(f.invoiceDateCommand(jan, 2))
	if result.AdjustedJournals != 2 {
		t.Fatalf("adjusted journals = %d", result.AdjustedJournals)
	}
	for _, want := range []struct{ journal, vat, state string }{
		{f.j1.ID, "P1", "2026-1/1"}, {f.j2.ID, "P3", "2026-1/1"},
		{f.j1.ID, "P2", "0-0/3"}, {f.j3.ID, "P4", "2026-5/1"}, {f.j2.ID, "S1", "2026-1/0"},
	} {
		if got := f.stored(want.journal, want.vat); got != want.state {
			t.Fatalf("%s = %s, want %s", want.vat, got, want.state)
		}
	}
	if f.version(f.j3.ID) != f.j3.Version {
		t.Fatal("a journal outside the filter was rewritten")
	}
	// running it again finds nothing to change
	f.failWith(f.invoiceDateCommand(jan, 2), "vat_adjust_nothing_changed")
	// the creditor filter narrows the listing too: P4 has no creditor, so a creditor range finds nothing among March invoices
	if _, rows := f.list(VatAdjustFilter{InvDateFrom: "2026-03-01", InvDateTo: "2026-03-31", PartnerFrom: "AP001", PartnerTo: "AP002"}, 1, 100); len(rows) != 0 {
		t.Fatalf("march rows with creditors = %s", vatAdjustIDs(rows))
	}
	if got := f.stored(f.j3.ID, "P4"); got != "2026-5/1" {
		t.Fatalf("P4 = %s", got)
	}
}

func TestVatAdjustInvoiceDateModeAcrossTheYear(t *testing.T) {
	f := newVatAdjustFixture(t)
	f.run(f.invoiceDateCommand(VatAdjustFilter{InvDateTo: "2026-12-31"}, 4))
	for _, want := range []struct{ journal, vat, state string }{
		{f.j1.ID, "P1", "2026-1/1"}, {f.j1.ID, "P2", "2026-2/3"}, {f.j2.ID, "P3", "2026-1/1"}, {f.j3.ID, "P4", "2026-3/1"},
	} {
		if got := f.stored(want.journal, want.vat); got != want.state {
			t.Fatalf("%s = %s, want %s", want.vat, got, want.state)
		}
	}
}
