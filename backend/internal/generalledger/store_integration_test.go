//go:build integration

package generalledger

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestLedgerMongoPostgresLifecycle(t *testing.T) {
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
	// A new database name and a separate PostgreSQL schema isolate every run.
	db := client.Database("bc_gl_uat_" + uuid.NewString())
	t.Cleanup(func() {
		if err := db.Drop(context.Background()); err != nil {
			t.Error(err)
		}
	})
	p, pg := glpgTestDB(t)
	store := newSyncTestStore(db, p)
	scope := Scope{Holding: "H", Company: "C", Branch: "B1", Actor: "uat-seed-20260911"}
	var sequence int
	command := func(resource, action, id string, version int64) Command {
		sequence++
		return Command{Resource: resource, Action: action, ID: id, Version: version, RequestID: fmt.Sprintf("uat-20260911-%06d", sequence)}
	}
	run := func(cmd Command) Result {
		t.Helper()
		r, err := store.Execute(ctx, scope, cmd)
		if err != nil {
			t.Fatal(err)
		}
		if r.ProjectionPending {
			t.Fatalf("projection pending after %s/%s: %v", cmd.Resource, cmd.Action, store.DeliverPending(ctx))
		}
		return r
	}
	checkSource := func(kind string, r Result) {
		t.Helper()
		var raw bson.Raw
		if err := db.Collection(collectionName(kind)).FindOne(ctx, scopedID(scope, r.ID)).Decode(&raw); err != nil {
			t.Fatal(err)
		}
		if v := raw.Lookup("__v").Int64(); v != r.Version {
			t.Fatalf("Mongo version %d != %d", v, r.Version)
		}
	}
	accounts := map[string]Result{}
	for _, spec := range []struct {
		code, kind, normal string
		cash               bool
	}{{"101", "asset", "debit", true}, {"301", "equity", "credit", false}, {"302", "equity", "credit", false}, {"401", "income", "credit", false}, {"501", "expense", "debit", false}} {
		cmd := command("accounts", "create", "", 0)
		cmd.Account = &Account{AccountCode: spec.code, Names: []Name{{Code: "th", Name: "บัญชีทดสอบ " + spec.code}}, AccountType: spec.kind, NormalBalance: spec.normal, AllowPosting: true, IsActive: true, IsCash: spec.cash}
		r := run(cmd)
		checkSource("accounts", r)
		accounts[spec.code] = r
	}
	yearCmd := command("fiscal-years", "create", "", 0)
	yearCmd.FiscalYear = &FiscalYear{Code: "2026", StartDate: "2026-01-01", EndDate: "2026-12-31", Scale: 2, IsActive: true, ProfitLossAccount: "301", RetainedEarningsAccount: "302"}
	year := run(yearCmd)
	checkSource("fiscal-years", year)
	periodCmd := command("periods", "create", "", 0)
	periodCmd.Master = &Master{Code: "2026-09", Name: "กันยายน", FiscalYear: "2026", StartDate: "2026-09-01", EndDate: "2026-09-30", IsActive: true}
	period := run(periodCmd)
	checkSource("periods", period)

	// Real create/read/update/delete; check Mongo immediately after every step.
	groupCmd := command("account-groups", "create", "", 0)
	groupCmd.Master = &Master{Code: "UAT-G", Name: "กลุ่ม & <ทดสอบ>", IsActive: true}
	group := run(groupCmd)
	checkSource("account-groups", group)
	data, err := p.Get(ctx, scope, "account-groups", group.ID)
	if err != nil {
		t.Fatal(err)
	}
	var groupRead Master
	if err = json.Unmarshal(data, &groupRead); err != nil || groupRead.Name != groupCmd.Master.Name {
		t.Fatalf("read %s %v", data, err)
	}
	update := command("account-groups", "update", group.ID, group.Version)
	update.Master = &Master{Code: "UAT-G", Name: "กลุ่มแก้ไข", IsActive: true}
	group = run(update)
	checkSource("account-groups", group)
	var rawGroup Master
	if err = db.Collection("gl_account_groups").FindOne(ctx, scopedID(scope, group.ID)).Decode(&rawGroup); err != nil || rawGroup.Name != "กลุ่มแก้ไข" {
		t.Fatalf("Mongo update not persisted: %v", err)
	}
	group = run(command("account-groups", "delete", group.ID, group.Version))
	checkSource("account-groups", group)
	if err = db.Collection("gl_account_groups").FindOne(ctx, scopedID(scope, group.ID)).Decode(&rawGroup); err != nil || !rawGroup.IsDeleted {
		t.Fatal("Mongo soft delete missing")
	}

	journalCmd := command("journals", "create", "", 0)
	journalCmd.Journal = &Journal{DocNo: "JV-UAT-001", Date: "2026-09-11", BookCode: "JV", FiscalYear: "2026", Description: "ทดสอบรับเงิน", BranchCode: "B1", Kind: "manual", Lines: []Line{{AccountCode: "101", Debit: "0.1", CashFlow: "operating"}, {AccountCode: "101", Debit: "0.2", CashFlow: "operating"}, {AccountCode: "401", Credit: "0.3"}}}
	journal := run(journalCmd)
	checkSource("journals", journal)
	var source bson.Raw
	if err = db.Collection("gl_journals").FindOne(ctx, scopedID(scope, journal.ID)).Decode(&source); err != nil {
		t.Fatal(err)
	}
	if source.Lookup("lines").Array().Index(0).Value().Document().Lookup("debit").Type != bsontype.Decimal128 {
		t.Fatal("Mongo journal amount is not Decimal128")
	}
	var count int
	if err = pg.QueryRow(`SELECT count(*) FROM gl_lines`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("draft affected ledger: %d %v", count, err)
	}
	post := command("journals", "post", journal.ID, journal.Version)
	journal = run(post)
	checkSource("journals", journal)
	if err = store.Ready(ctx, scope); err != nil {
		t.Fatal(err)
	}
	report, err := p.Report(ctx, scope, "trialbalance", ReportQuery{FiscalYear: "2026", From: "2026-09-01", To: "2026-09-30"})
	if err != nil {
		t.Fatal(err)
	}
	if report.Totals["debit"] != "0.3" && report.Totals["debit"] != "0.30000000" {
		t.Fatalf("exact persisted total: %+v", report.Totals)
	}
	// Repeating the identical command returns the first result, even after delivery.
	replay := run(post)
	if replay.ID != journal.ID || replay.Version != journal.Version || replay.Sequence != journal.Sequence {
		t.Fatal("idempotent command changed result")
	}
	post.Reason = "changed request"
	if _, err = store.Execute(ctx, scope, post); err == nil {
		t.Fatal("accepted reused request ID with changed payload")
	}
	for _, action := range []string{"update", "delete"} {
		bad := command("journals", action, journal.ID, journal.Version)
		bad.Journal = journalCmd.Journal
		if _, err = store.Execute(ctx, scope, bad); err == nil {
			t.Fatalf("allowed %s of posted journal", action)
		}
	}
	accountUpdate := command("accounts", "update", accounts["101"].ID, accounts["101"].Version)
	accountUpdate.Account = &Account{AccountCode: "101", Names: []Name{{Code: "th", Name: "เปลี่ยนหมวด"}}, AccountType: "expense", NormalBalance: "debit", AllowPosting: true, IsActive: true}
	if _, err = store.Execute(ctx, scope, accountUpdate); err == nil {
		t.Fatal("changed historical account type")
	}

	// Tenant and branch boundaries are checked before a mutation.
	other := scope
	other.Company = "OTHER"
	if _, err = store.Execute(ctx, other, command("journals", "delete", journal.ID, journal.Version)); err == nil {
		t.Fatal("cross company mutation")
	}
	other = scope
	other.Branch = "B2"
	if _, err = store.Execute(ctx, other, command("journals", "delete", journal.ID, journal.Version)); err == nil {
		t.Fatal("cross branch mutation")
	}
	otherReport, err := p.List(ctx, other, "journals", "", 1, 20, ListFilter{})
	if err != nil || otherReport.Total != 0 {
		t.Fatalf("branch data leak: %+v %v", otherReport, err)
	}

	reverse := command("journals", "reverse", journal.ID, journal.Version)
	reverse.DocNo = "JV-UAT-REV"
	reverse.Date = "2026-09-12"
	reverse.Reason = "ทดสอบกลับรายการ"
	journal = run(reverse)
	checkSource("journals", journal)
	var original Journal
	if err = db.Collection("gl_journals").FindOne(ctx, scopedID(scope, journal.ID)).Decode(&original); err != nil || original.Status != "reversed" || original.Lines[0].Debit != "0.1" {
		t.Fatal("original audit changed")
	}
	report, err = p.Report(ctx, scope, "trialbalance", ReportQuery{FiscalYear: "2026"})
	if err != nil {
		t.Fatal(err)
	}
	if report.Totals["balance"] != "0" && report.Totals["difference"] != "0" && report.Totals["difference"] != "0.00000000" {
		t.Fatalf("reversal not reconciled: %+v", report.Totals)
	}
	lock := command("periods", "lock", period.ID, period.Version)
	lock.Reason = "ทดสอบล็อกงวด"
	period = run(lock)
	checkSource("periods", period)
	blocked := journalCmd
	blocked.RequestID = "uat-locked-period-20260911"
	blocked.Journal = &Journal{}
	*blocked.Journal = *journalCmd.Journal
	blocked.Journal.DocNo = "JV-UAT-BLOCK"
	if _, err = store.Execute(ctx, scope, blocked); err == nil {
		t.Fatal("saved into locked period")
	}
	unlock := command("periods", "unlock", period.ID, period.Version)
	unlock.Reason = "คืนงวดทดสอบ"
	period = run(unlock)
	checkSource("periods", period)

	// Concurrent duplicate creates have one command/audit and one document.
	concurrent := command("account-groups", "create", "", 0)
	concurrent.Master = &Master{Code: "UAT-CONCURRENT", Name: "บันทึกซ้ำ", IsActive: true}
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for range 2 {
		wg.Add(1)
		go func() { defer wg.Done(); _, e := store.Execute(ctx, scope, concurrent); errs <- e }()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	n, err := db.Collection("gl_account_groups").CountDocuments(ctx, bson.M{"code": "UAT-CONCURRENT"})
	if err != nil || n != 1 {
		t.Fatalf("duplicate docs: %d %v", n, err)
	}
	n, err = db.Collection("gl_events").CountDocuments(ctx, bson.M{"requestid": concurrent.RequestID})
	if err != nil || n != 1 {
		t.Fatalf("duplicate events: %d %v", n, err)
	}
	if err = store.DeliverPending(ctx); err != nil {
		t.Fatal(err)
	}
	if err = store.Ready(ctx, scope); err != nil {
		t.Fatal(err)
	}
}
