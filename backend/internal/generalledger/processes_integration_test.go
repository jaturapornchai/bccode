//go:build integration

package generalledger

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Deterministic interleaving: a real Mongo+PG command commits after the process
// snapshot was read but before the caller receives it. No timing sleeps needed.
type processSnapshotInterleave struct {
	*Postgres
	afterRead func()
}

func (p *processSnapshotInterleave) ProcessBalances(ctx context.Context, scope Scope, year, to string) (ProcessBalanceSnapshot, error) {
	snapshot, err := p.Postgres.ProcessBalances(ctx, scope, year, to)
	if err == nil && p.afterRead != nil {
		hook := p.afterRead
		p.afterRead = nil
		hook()
	}
	return snapshot, err
}

func TestLedgerMongoPostgresProcessLifecycle(t *testing.T) {
	uri := os.Getenv("BC_GL_TEST_MONGO_URI")
	if uri == "" {
		t.Skip("set BC_GL_TEST_MONGO_URI to isolated replica set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })
	db := client.Database("bc_gl_process_uat_" + uuid.NewString())
	t.Cleanup(func() {
		if err := db.Drop(context.Background()); err != nil {
			t.Error(err)
		}
	})
	p, pg := glpgTestDB(t)
	store := newSyncTestStore(db, p)
	scope := Scope{Holding: "H", Company: "PROCESS", Actor: "process-uat-seed-20260911"}
	request := 0
	command := func(resource, action, id string, version int64) Command {
		request++
		return Command{Resource: resource, Action: action, ID: id, Version: version, RequestID: fmt.Sprintf("process-uat-20260911-%06d", request)}
	}
	run := func(cmd Command) Result {
		t.Helper()
		result, err := store.Execute(ctx, scope, cmd)
		if err != nil {
			t.Fatalf("%s/%s: %v", cmd.Resource, cmd.Action, err)
		}
		if result.ProjectionPending {
			t.Fatalf("%s/%s projection pending: %v", cmd.Resource, cmd.Action, store.DeliverPending(ctx))
		}
		return result
	}
	sourceIdentity := func(kind string, result Result) {
		t.Helper()
		var raw bson.Raw
		if err := db.Collection(collectionName(kind)).FindOne(ctx, scopedID(scope, result.ID)).Decode(&raw); err != nil {
			t.Fatal(err)
		}
		if raw.Lookup("__v").Int64() != result.Version {
			t.Fatal("Mongo version differs from result")
		}
	}
	readJournal := func(id, status string) Journal {
		t.Helper()
		var raw bson.Raw
		if err := db.Collection("gl_journals").FindOne(ctx, scopedID(scope, id)).Decode(&raw); err != nil {
			t.Fatal(err)
		}
		var journal Journal
		if err := bson.Unmarshal(raw, &journal); err != nil {
			t.Fatal(err)
		}
		if journal.Status != status {
			t.Fatalf("journal %s status %s want %s", id, journal.Status, status)
		}
		debit, credit := decimal.Zero, decimal.Zero
		for i, line := range journal.Lines {
			document := raw.Lookup("lines").Array().Index(uint(i)).Value().Document()
			if document.Lookup("debit").Type != bsontype.Decimal128 || document.Lookup("credit").Type != bsontype.Decimal128 {
				t.Fatalf("journal %s line %d lost Decimal128", id, i)
			}
			debit = debit.Add(line.Debit.Decimal())
			credit = credit.Add(line.Credit.Decimal())
		}
		if !debit.Equal(credit) {
			t.Fatalf("source journal not balanced: %s %s", debit, credit)
		}
		return journal
	}
	checkPGJournal := func(j Journal) {
		t.Helper()
		var count int
		var difference string
		if err := pg.QueryRow(`SELECT count(*),COALESCE(SUM(debit-credit),0)::text FROM gl_lines WHERE company=$1 AND journal_id=$2`, scope.Company, j.ID).Scan(&count, &difference); err != nil {
			t.Fatal(err)
		}
		expected := len(j.Lines)
		if j.Status == "draft" {
			expected = 0
		}
		if count != expected {
			t.Fatalf("PG lines for %s=%d want %d", j.DocNo, count, expected)
		}
		glpgDecimalEqual(t, difference, "0")
	}
	for _, spec := range []struct {
		code, kind, normal string
		cash               bool
	}{{"101", "asset", "debit", true}, {"301", "equity", "credit", false}, {"302", "equity", "credit", false}, {"401", "income", "credit", false}, {"501", "expense", "debit", false}} {
		cmd := command("accounts", "create", "", 0)
		cmd.Account = &Account{AccountCode: spec.code, Names: []Name{{Code: "th", Name: "บัญชีทดสอบ " + spec.code}}, AccountType: spec.kind, NormalBalance: spec.normal, AllowPosting: true, IsActive: true, IsCash: spec.cash}
		result := run(cmd)
		sourceIdentity("accounts", result)
	}
	years := map[string]Result{}
	for _, code := range []string{"2026", "2027"} {
		cmd := command("fiscal-years", "create", "", 0)
		cmd.FiscalYear = &FiscalYear{Code: code, StartDate: code + "-01-01", EndDate: code + "-12-31", Currency: "THB", Scale: 2, IsActive: true, ProfitLossAccount: "301", RetainedEarningsAccount: "302"}
		years[code] = run(cmd)
		sourceIdentity("fiscal-years", years[code])
		period := command("periods", "create", "", 0)
		period.Master = &Master{Code: code + "-FULL", Name: "งวดทดสอบ " + code, FiscalYear: code, StartDate: code + "-01-01", EndDate: code + "-12-31", IsActive: true}
		sourceIdentity("periods", run(period))
	}
	// Cash lines intentionally have no department/project while their income
	// and expense counterparts do. Only each branch must balance in total.
	for _, spec := range []struct {
		docno, branch, department, project, amount string
		expense                                    bool
	}{{"B1-IN", "B1", "D1", "P1", "100.30", false}, {"B1-OUT", "B1", "D1", "P1", "20.10", true}, {"B2-IN", "B2", "D2", "P2", "200.20", false}, {"B2-OUT", "B2", "D2", "P2", "30.20", true}} {
		cmd := command("journals", "create", "", 0)
		j := &Journal{DocNo: spec.docno, Date: "2026-09-11", BookCode: "JV", FiscalYear: "2026", Description: "ทดสอบกระบวนการ " + spec.docno, BranchCode: spec.branch, Currency: "THB", Kind: "manual"}
		if spec.expense {
			j.Lines = []Line{{AccountCode: "501", Debit: Amount(spec.amount), DepartmentCode: spec.department, ProjectCode: spec.project}, {AccountCode: "101", Credit: Amount(spec.amount), CashFlow: "operating"}}
		} else {
			j.Lines = []Line{{AccountCode: "101", Debit: Amount(spec.amount), CashFlow: "operating"}, {AccountCode: "401", Credit: Amount(spec.amount), DepartmentCode: spec.department, ProjectCode: spec.project}}
		}
		cmd.Journal = j
		draft := run(cmd)
		sourceIdentity("journals", draft)
		checkPGJournal(readJournal(draft.ID, "draft"))
		posted := run(command("journals", "post", draft.ID, draft.Version))
		sourceIdentity("journals", posted)
		checkPGJournal(readJournal(posted.ID, "posted"))
	}
	query := ReportQuery{FiscalYear: "2026", From: "2026-01-01", To: "2026-12-31"}
	pnl, err := p.Report(ctx, scope, "pnl", query)
	if err != nil {
		t.Fatal(err)
	}
	glpgDecimalEqual(t, pnl.Totals["revenue"], "300.50")
	glpgDecimalEqual(t, pnl.Totals["expense"], "50.30")
	glpgDecimalEqual(t, pnl.Totals["profit"], "250.20")
	closingCommand := func(prefix string) Command {
		cmd := command("processes", "close", "2026", years["2026"].Version)
		cmd.Date = "2026-12-31"
		cmd.DocNo = prefix
		cmd.Reason = "ตรวจยอดปิดงบสองสาขา"
		return cmd
	}
	// A stale fiscal version is rejected before creating any process document.
	staleVersion := closingCommand("STALE-VERSION")
	staleVersion.Version++
	if _, err = store.Execute(ctx, scope, staleVersion); err == nil {
		t.Fatal("accepted stale fiscal year version")
	}
	// A concurrent real command must invalidate the prepared PG snapshot.
	interleaved := &processSnapshotInterleave{Postgres: p}
	interleaved.afterRead = func() {
		cmd := command("account-groups", "create", "", 0)
		cmd.Master = &Master{Code: "INTERLEAVE", Name: "เขียนระหว่างเตรียมปิดงบ", IsActive: true}
		sourceIdentity("account-groups", run(cmd))
	}
	staleSnapshot := closingCommand("STALE-SNAPSHOT")
	if _, err = newSyncTestStore(db, interleaved).Execute(ctx, scope, staleSnapshot); err == nil {
		t.Fatal("accepted process using stale snapshot")
	}
	count, err := db.Collection("gl_journals").CountDocuments(ctx, bson.M{"holdingcode": scope.Holding, "businesscode": scope.Company, "kind": "closing"})
	if err != nil || count != 0 {
		t.Fatalf("failed process created documents: %d %v", count, err)
	}
	generated := func(cmd Command, result Result, kind string) []Journal {
		t.Helper()
		var event Event
		if err := db.Collection("gl_events").FindOne(ctx, bson.M{"holdingcode": scope.Holding, "businesscode": scope.Company, "requestid": cmd.RequestID}).Decode(&event); err != nil {
			t.Fatal(err)
		}
		if event.Sequence != result.Sequence || !event.Delivered {
			t.Fatal("process event not delivered at returned sequence")
		}
		journals := []Journal{}
		for _, change := range event.Changes {
			if change.Kind != "journals" {
				continue
			}
			j := readJournal(change.ID, "draft")
			if j.Kind != kind {
				t.Fatalf("generated kind=%s", j.Kind)
			}
			checkPGJournal(j)
			journals = append(journals, j)
		}
		if len(journals) != 2 || journals[0].BranchCode == journals[1].BranchCode {
			t.Fatalf("expected separate branch drafts: %#v", journals)
		}
		return journals
	}
	closeCmd := closingCommand("CLOSE-2026")
	closedResult := run(closeCmd)
	closings := generated(closeCmd, closedResult, "closing")
	for _, j := range closings {
		department, project := "D1", "P1"
		if j.BranchCode == "B2" {
			department, project = "D2", "P2"
		}
		for _, line := range j.Lines {
			if line.DepartmentCode != department || line.ProjectCode != project {
				t.Fatalf("closing lost dimensions: %#v", line)
			}
		}
		posted := run(command("journals", "post", j.ID, j.Version))
		sourceIdentity("journals", posted)
		checkPGJournal(readJournal(posted.ID, "posted"))
	}
	pnlAfter, err := p.Report(ctx, scope, "pnl", query)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(pnlAfter.Totals, pnl.Totals) {
		t.Fatalf("closing erased performance: before=%#v after=%#v", pnl.Totals, pnlAfter.Totals)
	}
	trial, err := p.Report(ctx, scope, "trialbalance", query)
	if err != nil {
		t.Fatal(err)
	}
	glpgDecimalEqual(t, trial.Totals["difference"], "0")
	for _, row := range trial.Rows {
		if row["accounttype"] == "income" || row["accounttype"] == "expense" {
			glpgDecimalEqual(t, row["balance"], "0")
		}
	}
	carrySource, err := p.ProcessBalances(ctx, scope, "2026", "2026-12-31")
	if err != nil {
		t.Fatal(err)
	}
	endCmd := command("processes", "year-end", "2026", years["2026"].Version)
	endCmd.TargetYear = "2027"
	endCmd.Date = "2027-01-01"
	endCmd.DocNo = "OPEN-2027"
	endCmd.Reason = "ยกยอดจากปีที่ปิดแล้ว"
	ended := run(endCmd)
	// Check the fiscal source immediately after the operation, before posting.
	var closedYear FiscalYear
	if err = db.Collection("fiscal_year").FindOne(ctx, scopedID(scope, years["2026"].ID)).Decode(&closedYear); err != nil || !closedYear.Closed || closedYear.Version != years["2026"].Version+1 {
		t.Fatalf("fiscal close not committed: %#v %v", closedYear, err)
	}
	openings := generated(endCmd, ended, "opening")
	replay := run(endCmd)
	if replay != ended {
		t.Fatalf("replay after fiscal closed differs: %#v %#v", replay, ended)
	}
	for _, j := range openings {
		if j.FiscalYear != "2027" || j.Date != "2027-01-01" || j.Reference != "YEAR-END:2026" {
			t.Fatalf("bad opening identity %#v", j)
		}
		update := command("journals", "update", j.ID, j.Version)
		update.Journal = &j
		for _, cmd := range []Command{update, command("journals", "delete", j.ID, j.Version)} {
			if _, err = store.Execute(ctx, scope, cmd); err == nil {
				t.Fatalf("accepted generated-opening %s", cmd.Action)
			}
			unchanged := readJournal(j.ID, "draft")
			if unchanged.Version != j.Version {
				t.Fatal("rejected command mutated opening")
			}
		}
		posted := run(command("journals", "post", j.ID, j.Version))
		sourceIdentity("journals", posted)
		checkPGJournal(readJournal(posted.ID, "posted"))
	}
	carried, err := p.ProcessBalances(ctx, scope, "2027", "2027-12-31")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(carried.Rows, carrySource.Rows) {
		t.Fatalf("opening changed dimensions: before=%#v after=%#v", carrySource.Rows, carried.Rows)
	}
	if replay = run(endCmd); replay != ended {
		t.Fatal("year-end replay after openings posted changed original result")
	}
	// Recalculation is an audited derived-state rebuild, preserving source IDs,
	// versions, all dimensions and every immutable PostgreSQL event.
	var auditsBefore int
	if err = pg.QueryRow(`SELECT count(*) FROM gl_events WHERE company=$1`, scope.Company).Scan(&auditsBefore); err != nil {
		t.Fatal(err)
	}
	recalc := command("processes", "recalculate", "2027", 1)
	recalc.Reason = "ตรวจประวัติและคำนวณใหม่"
	recalculated := run(recalc)
	var recalcEvent Event
	if err = db.Collection("gl_events").FindOne(ctx, bson.M{"holdingcode": scope.Holding, "businesscode": scope.Company, "requestid": recalc.RequestID}).Decode(&recalcEvent); err != nil || len(recalcEvent.Changes) != 0 || !recalcEvent.Delivered {
		t.Fatalf("recalculate source audit %#v %v", recalcEvent, err)
	}
	rebuilt, err := p.ProcessBalances(ctx, scope, "2027", "2027-12-31")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rebuilt.Rows, carried.Rows) {
		t.Fatal("recalculate changed financial dimensions")
	}
	if replay = run(recalc); replay != recalculated {
		t.Fatal("recalculate replay changed event")
	}
	var auditsAfter int
	if err = pg.QueryRow(`SELECT count(*) FROM gl_events WHERE company=$1`, scope.Company).Scan(&auditsAfter); err != nil || auditsAfter != auditsBefore+1 {
		t.Fatalf("recalculate audit duplicated/lost: before=%d after=%d err=%v", auditsBefore, auditsAfter, err)
	}
	if _, err = pg.Exec(`UPDATE gl_events SET event_hash='changed' WHERE company=$1`, scope.Company); err == nil {
		t.Fatal("audit became mutable after rebuild")
	}
	for _, j := range openings {
		saved := readJournal(j.ID, "posted")
		if saved.Version != j.Version+1 || !strings.HasPrefix(saved.Reference, "YEAR-END:") {
			t.Fatal("recalculate changed source journal")
		}
		checkPGJournal(saved)
	}
}
