//go:build integration

package generalledger

import (
	"encoding/json"
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

func TestLedgerMongoPostgresEightDecimalAmounts(t *testing.T) {
	ctx, store, p, db := glAuditStore(t)
	scope := Scope{Holding: "H", Company: "C", Branch: "B1", Actor: "audit-scale8-20260911"}
	serial := 0
	command := func(resource, action, id string, version int64) Command {
		serial++
		return Command{Resource: resource, Action: action, ID: id, Version: version, RequestID: fmt.Sprintf("audit-scale8-20260911-%06d", serial)}
	}
	run := func(cmd Command) Result {
		t.Helper()
		result, err := store.Execute(ctx, scope, cmd)
		if err != nil || result.ProjectionPending {
			t.Fatalf("%s/%s: %+v %v", cmd.Resource, cmd.Action, result, err)
		}
		return result
	}
	for _, code := range []string{"101", "102"} {
		cmd := command("accounts", "create", "", 0)
		cmd.Account = &Account{AccountCode: code, Names: []Name{{Code: "th", Name: "บัญชีทศนิยมแปดตำแหน่ง"}}, AccountType: "asset", NormalBalance: "debit", IsActive: true, AllowPosting: true}
		run(cmd)
	}
	year := command("fiscal-years", "create", "", 0)
	year.FiscalYear = &FiscalYear{Code: "2026", StartDate: "2026-01-01", EndDate: "2026-12-31", Scale: 8, IsActive: true}
	run(year)
	period := command("periods", "create", "", 0)
	period.Master = &Master{Code: "2026", Name: "งวดทดสอบ", FiscalYear: "2026", StartDate: "2026-01-01", EndDate: "2026-12-31", IsActive: true}
	run(period)
	create := command("journals", "create", "", 0)
	// Decode the same decimal-string JSON accepted at the API boundary.
	var journal Journal
	if err := json.Unmarshal([]byte(`{"docno":"SCALE8","date":"2026-09-11","bookcode":"JV","fiscalyear":"2026","description":"ทดสอบจำนวนเงินแปดตำแหน่ง","kind":"manual","branchcode":"B1","lines":[{"accountcode":"101","debit":"0.00000001"},{"accountcode":"101","debit":"0.00000002"},{"accountcode":"102","credit":"0.00000003"}]}`), &journal); err != nil {
		t.Fatal(err)
	}
	create.Journal = &journal
	created := run(create)
	checkMongo := func(expected Result, status string) {
		t.Helper()
		var raw bson.Raw
		if err := db.Collection("gl_journals").FindOne(ctx, scopedID(scope, created.ID)).Decode(&raw); err != nil {
			t.Fatal(err)
		}
		if raw.Lookup("lines").Array().Index(0).Value().Document().Lookup("debit").Type != bsontype.Decimal128 {
			t.Fatal("small amount lost Decimal128 type")
		}
		var stored Journal
		if err := bson.Unmarshal(raw, &stored); err != nil || stored.Version != expected.Version || stored.Status != status || stored.Lines[0].Debit != "0.00000001" || stored.Lines[1].Debit != "0.00000002" || stored.Lines[2].Credit != "0.00000003" {
			t.Fatalf("small amounts did not round-trip through Mongo: %+v %v", stored, err)
		}
	}
	checkMongo(created, "draft")
	post := command("journals", "post", created.ID, created.Version)
	posted := run(post)
	checkMongo(posted, "posted")
	report, err := p.Report(ctx, scope, "trialbalance", ReportQuery{FiscalYear: "2026", From: "2026-01-01", To: "2026-12-31"})
	if err != nil || report.Totals["debit"] != "0.00000003" || report.Totals["credit"] != "0.00000003" || report.Totals["difference"] != "0.00000000" {
		t.Fatalf("small SQL amounts did not sum exactly: %+v %v", report.Totals, err)
	}
	if replay := run(post); replay != posted {
		t.Fatalf("small amount post replay changed result: %+v != %+v", replay, posted)
	}
	checkMongo(posted, "posted")
}
