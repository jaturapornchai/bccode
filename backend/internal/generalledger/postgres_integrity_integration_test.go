//go:build integration

package generalledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

type pgIntegrityFixture struct {
	t        *testing.T
	ctx      context.Context
	db       *sql.DB
	store    *PostgresStore
	scope    Scope
	accounts map[string]Account
	years    map[string]FiscalYear
}

func newPGIntegrityFixture(t *testing.T) *pgIntegrityFixture {
	pg, db := glpgTestDB(t)
	f := &pgIntegrityFixture{t: t, ctx: context.Background(), db: db, store: NewPostgresStore(pg), scope: Scope{Holding: "H", Company: "C", Actor: "tester"}, accounts: map[string]Account{}, years: map[string]FiscalYear{}}
	for _, row := range []struct{ code, kind, normal string }{{"1000", "asset", "debit"}, {"3000", "equity", "credit"}, {"3100", "equity", "credit"}, {"3200", "equity", "credit"}, {"4000", "income", "credit"}, {"5000", "expense", "debit"}} {
		a := Account{AccountCode: row.code, AccountType: row.kind, NormalBalance: row.normal, AllowPosting: true, IsActive: true, Names: []Name{{Code: "th", Name: row.code}}}
		result := f.run(Command{Resource: "accounts", Action: "create", Account: &a})
		a.ID = result.ID
		a.Version = result.Version
		f.accounts[row.code] = a
	}
	for _, code := range []string{"2026", "2027"} {
		y := FiscalYear{Code: code, StartDate: code + "-01-01", EndDate: code + "-12-31", IsActive: true, Scale: 2, ProfitLossAccount: "3200", RetainedEarningsAccount: "3100"}
		r := f.run(Command{Resource: "fiscal-years", Action: "create", FiscalYear: &y})
		y.ID = r.ID
		y.Version = r.Version
		f.years[code] = y
	}
	return f
}
func (f *pgIntegrityFixture) execute(scope Scope, cmd Command) (Result, error) {
	if cmd.RequestID == "" {
		cmd.RequestID = uuid.NewString()
	}
	return f.store.Execute(f.ctx, scope, cmd)
}
func (f *pgIntegrityFixture) run(cmd Command) Result {
	f.t.Helper()
	r, err := f.execute(f.scope, cmd)
	if err != nil {
		f.t.Fatalf("%s/%s failed: %v", cmd.Resource, cmd.Action, err)
	}
	return r
}
func (f *pgIntegrityFixture) deny(scope Scope, cmd Command) {
	f.t.Helper()
	if _, err := f.execute(scope, cmd); err == nil {
		f.t.Fatalf("accepted forbidden %s/%s", cmd.Resource, cmd.Action)
	}
}
func (f *pgIntegrityFixture) journal(id string) Journal {
	f.t.Helper()
	raw, err := f.store.Get(f.ctx, f.scope, "journals", id)
	if err != nil {
		f.t.Fatal(err)
	}
	var j Journal
	if err = json.Unmarshal(raw, &j); err != nil {
		f.t.Fatal(err)
	}
	return j
}
func (f *pgIntegrityFixture) draft(doc, branch string) Journal {
	j := Journal{DocNo: doc, Date: "2026-01-10", BookCode: "JV", FiscalYear: "2026", Description: "test", Kind: "manual", BranchCode: branch, Lines: []Line{{AccountCode: "1000", Debit: Amount("100"), Credit: Amount("0"), DepartmentCode: "D", ProjectCode: "P"}, {AccountCode: "4000", Debit: Amount("0"), Credit: Amount("100"), DepartmentCode: "D", ProjectCode: "P"}}}
	r := f.run(Command{Resource: "journals", Action: "create", Journal: &j})
	return f.journal(r.ID)
}
func (f *pgIntegrityFixture) post(j Journal) Journal {
	f.run(Command{Resource: "journals", Action: "post", ID: j.ID, Version: j.Version})
	return f.journal(j.ID)
}
func TestPostgresIntegrityLifecycleAndBranch(t *testing.T) {
	f := newPGIntegrityFixture(t)
	j := f.draft("JV1", "B1")
	other := f.scope
	other.Branch = "B2"
	f.deny(other, Command{Resource: "journals", Action: "post", ID: j.ID, Version: j.Version})
	f.deny(other, Command{Resource: "journals", Action: "create", Journal: &j})
	changed := j
	changed.BranchCode = "B2"
	own := f.scope
	own.Branch = "B1"
	f.deny(own, Command{Resource: "journals", Action: "update", ID: j.ID, Version: j.Version, Journal: &changed})
	j = f.post(j)
	f.deny(other, Command{Resource: "journals", Action: "reverse", ID: j.ID, Version: j.Version, DocNo: "REV-BAD", Date: "2026-01-11", Reason: "test"})
	f.run(Command{Resource: "journals", Action: "reverse", ID: j.ID, Version: j.Version, DocNo: "REV1", Date: "2026-01-11", Reason: "correct"})
	j = f.journal(j.ID)
	f.deny(f.scope, Command{Resource: "journals", Action: "post", ID: j.ID, Version: j.Version})
	var raw []byte
	if err := f.db.QueryRow(`SELECT payload FROM gl_records WHERE company='C' AND kind='journals' AND code='REV1'`).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	var rev Journal
	if err := json.Unmarshal(raw, &rev); err != nil {
		t.Fatal(err)
	}
	if rev.BranchCode != "B1" || rev.ReversalOf != j.ID || rev.Kind != "reversal" || rev.Lines[0].DepartmentCode != "D" || rev.Lines[0].ProjectCode != "P" {
		t.Fatalf("reversal lost identity/dimensions: %+v", rev)
	}
	f.deny(f.scope, Command{Resource: "journals", Action: "reverse", ID: rev.ID, Version: rev.Version, DocNo: "REV2", Date: "2026-01-12", Reason: "again"})
	var total string
	if err := f.db.QueryRow(`SELECT SUM(debit-credit)::text FROM gl_lines WHERE company='C' AND branch_code='B1' AND account_code='1000'`).Scan(&total); err != nil {
		t.Fatal(err)
	}
	if total != "0.00000000" {
		t.Fatalf("branch not reconciled: %s", total)
	}
}
func TestPostgresIntegrityPeriodAndValidation(t *testing.T) {
	f := newPGIntegrityFixture(t)
	j := f.draft("JV1", "B1")
	p := Master{Code: "2026-01", FiscalYear: "2026", StartDate: "2026-01-01", EndDate: "2026-01-31", IsActive: true}
	r := f.run(Command{Resource: "periods", Action: "create", Master: &p})
	f.deny(f.scope, Command{Resource: "periods", Action: "lock", ID: r.ID, Version: r.Version, Reason: "pending"})
	j = f.post(j)
	r = f.run(Command{Resource: "periods", Action: "lock", ID: r.ID, Version: r.Version, Reason: "close month"})
	f.deny(f.scope, Command{Resource: "journals", Action: "reverse", ID: j.ID, Version: j.Version, DocNo: "REV1", Date: "2026-01-11", Reason: "locked"})
	next := j
	next.DocNo = "JV2"
	next.Status = "draft"
	next.PostedAt = nil
	next.PostedBy = ""
	f.deny(f.scope, Command{Resource: "journals", Action: "create", Journal: &next})
	r = f.run(Command{Resource: "periods", Action: "unlock", ID: r.ID, Version: r.Version, Reason: "correction"})
	pending := f.draft("JV2", "B1")
	// Simulate a pre-existing locked period containing a legacy draft.
	if _, err := f.db.Exec(`UPDATE gl_records SET payload=jsonb_set(payload,'{locked}','true') WHERE company='C' AND kind='periods'`); err != nil {
		t.Fatal(err)
	}
	f.deny(f.scope, Command{Resource: "journals", Action: "post", ID: pending.ID, Version: pending.Version})
	if _, err := f.db.Exec(`UPDATE gl_records SET payload=jsonb_set(payload,'{locked}','false') WHERE company='C' AND kind='periods'`); err != nil {
		t.Fatal(err)
	}
	a := f.accounts["1000"]
	a.IsActive = false
	f.run(Command{Resource: "accounts", Action: "update", ID: a.ID, Version: a.Version, Account: &a})
	f.deny(f.scope, Command{Resource: "journals", Action: "post", ID: pending.ID, Version: pending.Version})
	y := f.years["2026"]
	y.Scale = 3
	f.deny(f.scope, Command{Resource: "fiscal-years", Action: "update", ID: y.ID, Version: y.Version, FiscalYear: &y})
	a = f.accounts["4000"]
	a.AccountType = "liability"
	f.deny(f.scope, Command{Resource: "accounts", Action: "update", ID: a.ID, Version: a.Version, Account: &a})
	f.deny(f.scope, Command{Resource: "accounts", Action: "delete", ID: a.ID, Version: a.Version})
}
func TestPostgresIntegrityCloseYearEndAndReplay(t *testing.T) {
	f := newPGIntegrityFixture(t)
	f.post(f.draft("JV1", "B1"))
	year := f.years["2026"]
	closeCmd := Command{Resource: "processes", Action: "close", ID: "2026", Version: year.Version, DocNo: "CLOSE26", Date: "2026-12-31", Reason: "close", RequestID: uuid.NewString()}
	closeResult := f.run(closeCmd)
	if closeResult.CreatedJournals != 1 {
		t.Fatalf("close generated %d", closeResult.CreatedJournals)
	}
	retry := f.run(closeCmd)
	if retry.ID != closeResult.ID || retry.CreatedJournals != closeResult.CreatedJournals {
		t.Fatal("close retry changed result")
	}
	closing := f.journal(closeResult.ID)
	if closing.Kind != "closing" || closing.Status != "draft" {
		t.Fatal("close did not create draft")
	}
	f.deny(f.scope, Command{Resource: "processes", Action: "year-end", ID: "2026", Version: year.Version, TargetYear: "2027", DocNo: "OPEN27", Date: "2027-01-01", Reason: "carry"})
	f.post(closing)
	endCmd := Command{Resource: "processes", Action: "year-end", ID: "2026", Version: year.Version, TargetYear: "2027", DocNo: "OPEN27", Date: "2027-01-01", Reason: "carry", RequestID: uuid.NewString()}
	result := f.run(endCmd)
	if result.CreatedJournals != 1 {
		t.Fatal("no opening")
	}
	opening := f.journal(result.ID)
	if opening.Kind != "opening" || opening.BranchCode != "B1" {
		t.Fatal("invalid opening")
	}
	raw, err := f.store.Get(f.ctx, f.scope, "fiscal-years", year.ID)
	if err != nil {
		t.Fatal(err)
	}
	var closed FiscalYear
	_ = json.Unmarshal(raw, &closed)
	if !closed.Closed {
		t.Fatal("year not closed")
	}
	retry = f.run(endCmd)
	if retry.ID != result.ID || retry.CreatedJournals != result.CreatedJournals {
		t.Fatal("yearend retry changed result")
	} // Same request is still idempotent after the year is closed.
	endCmd.RequestID = uuid.NewString()
	endCmd.Version = closed.Version
	f.deny(f.scope, endCmd)
	f.deny(f.scope, Command{Resource: "journals", Action: "update", ID: opening.ID, Version: opening.Version, Journal: &opening})
	f.deny(f.scope, Command{Resource: "journals", Action: "delete", ID: opening.ID, Version: opening.Version})
	f.post(opening)
	for _, action := range []string{"recalculate", "reprocess"} {
		if _, err = f.db.Exec(`DELETE FROM gl_lines WHERE company='C'`); err != nil {
			t.Fatal(err)
		}
		f.run(Command{Resource: "processes", Action: action, ID: "2026", Reason: "restore exact audit"})
		var count int
		if err = f.db.QueryRow(`SELECT COUNT(*) FROM gl_lines WHERE company='C'`).Scan(&count); err != nil || count == 0 {
			t.Fatalf("rebuild did nothing: %d %v", count, err)
		}
	}
}

func TestPostgresIntegrityConcurrentPostAndCrossYearReversal(t *testing.T) {
	f := newPGIntegrityFixture(t)
	j := f.draft("CONCURRENT", "B1")
	results := make(chan error, 2)
	for range 2 {
		go func() {
			_, err := f.execute(f.scope, Command{Resource: "journals", Action: "post", ID: j.ID, Version: j.Version})
			results <- err
		}()
	}
	succeeded := 0
	for range 2 {
		if <-results == nil {
			succeeded++
		}
	}
	if succeeded != 1 {
		t.Fatalf("concurrent post accepted %d requests", succeeded)
	}
	j = f.journal(j.ID)
	f.run(Command{Resource: "journals", Action: "reverse", ID: j.ID, Version: j.Version, DocNo: "REV2027", Date: "2027-01-01", Reason: "next year adjustment"})
	var raw []byte
	if err := f.db.QueryRow(`SELECT payload FROM gl_records WHERE company='C' AND kind='journals' AND code='REV2027'`).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	var rev Journal
	_ = json.Unmarshal(raw, &rev)
	if rev.FiscalYear != "2027" || rev.BranchCode != "B1" {
		t.Fatalf("cross-year reversal used wrong scope: %+v", rev)
	}
	pending := f.draft("LEGACYDRAFT", "B1")
	if _, err := f.db.Exec(`UPDATE gl_records SET payload=jsonb_set(payload,'{closed}','true') WHERE company='C' AND kind='fiscal-years' AND code='2026'`); err != nil {
		t.Fatal(err)
	}
	f.deny(f.scope, Command{Resource: "journals", Action: "post", ID: pending.ID, Version: pending.Version})
	y := f.years["2026"]
	f.deny(f.scope, Command{Resource: "fiscal-years", Action: "update", ID: y.ID, Version: y.Version, FiscalYear: &y})
}

func TestPostgresIntegrityManualReferenceAndManagedMetadata(t *testing.T) {
	f := newPGIntegrityFixture(t)
	j := f.draft("REFERENCE", "B1")
	j.Reference = "YEAR-END:customer reference"
	f.run(Command{Resource: "journals", Action: "update", ID: j.ID, Version: j.Version, Journal: &j})
	j = f.journal(j.ID)
	j.Description = "edited legitimate reference"
	f.run(Command{Resource: "journals", Action: "update", ID: j.ID, Version: j.Version, Journal: &j})
	j = f.journal(j.ID)
	j.ReversalOf = "forged"
	j.PostedBy = "forged"
	f.run(Command{Resource: "journals", Action: "update", ID: j.ID, Version: j.Version, Journal: &j})
	j = f.journal(j.ID)
	if j.ReversalOf != "" || j.PostedBy != "" || j.PostedAt != nil {
		t.Fatal("draft accepted forged audit metadata")
	}
	forged := j
	forged.DocNo = "FORGED"
	forged.ReversalOf = "unrelated"
	f.deny(f.scope, Command{Resource: "journals", Action: "create", Journal: &forged})
	f.run(Command{Resource: "journals", Action: "delete", ID: j.ID, Version: j.Version})
}
