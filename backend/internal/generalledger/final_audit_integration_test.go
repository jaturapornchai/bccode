//go:build integration

package generalledger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func glAuditStore(t *testing.T) (context.Context, *Store, *Postgres, *mongo.Database) {
	t.Helper()
	uri := os.Getenv("BC_GL_TEST_MONGO_URI")
	if uri == "" {
		t.Skip("set BC_GL_TEST_MONGO_URI to isolated replica set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	t.Cleanup(cancel)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })
	db := client.Database("bc_gl_audit_" + uuid.NewString())
	t.Cleanup(func() {
		if err := db.Drop(context.Background()); err != nil {
			t.Error(err)
		}
	})
	p, _ := glpgTestDB(t)
	return ctx, newSyncTestStore(db, p), p, db
}

func TestLedgerMongoPostgresFinancialMasterYearGuards(t *testing.T) {
	for _, kind := range []string{"budgets", "forecast"} {
		t.Run(kind, func(t *testing.T) {
			ctx, store, p, db := glAuditStore(t)
			scope := Scope{Holding: "H", Company: "C", Actor: "audit-seed-20260911"}
			serial := 0
			command := func(resource, action, id string, version int64) Command {
				serial++
				return Command{Resource: resource, Action: action, ID: id, Version: version, RequestID: fmt.Sprintf("audit-master-20260911-%06d", serial)}
			}
			run := func(cmd Command) Result {
				t.Helper()
				result, err := store.Execute(ctx, scope, cmd)
				if err != nil || result.ProjectionPending {
					t.Fatalf("%s/%s: %+v %v", cmd.Resource, cmd.Action, result, err)
				}
				return result
			}
			for _, account := range []Account{
				{AccountCode: "101", AccountType: "asset", NormalBalance: "debit", IsCash: true},
				{AccountCode: "301", AccountType: "equity", NormalBalance: "credit"},
			} {
				account.IsActive, account.AllowPosting = true, true
				account.Names = []Name{{Code: "th", Name: "บัญชีทดสอบ " + account.AccountCode}}
				cmd := command("accounts", "create", "", 0)
				cmd.Account = &account
				run(cmd)
			}
			var oldYear Result
			yearSpec := FiscalYear{Code: "2026", StartDate: "2026-01-01", EndDate: "2026-12-31", Scale: 2, IsActive: true}
			for _, year := range []FiscalYear{yearSpec, {Code: "2027", StartDate: "2027-01-01", EndDate: "2027-12-31", Scale: 2, IsActive: true}} {
				cmd := command("fiscal-years", "create", "", 0)
				cmd.FiscalYear = &year
				result := run(cmd)
				if year.Code == "2026" {
					oldYear = result
				}
				period := command("periods", "create", "", 0)
				period.Master = &Master{Code: year.Code, Name: "งวดทดสอบ", FiscalYear: year.Code, StartDate: year.StartDate, EndDate: year.EndDate, IsActive: true}
				run(period)
			}
			planSpec := Master{Code: "PLAN", Name: "แผนทดสอบ", FiscalYear: "2026", StartDate: "2026-09-01", EndDate: "2026-09-30", Amount: "12.34", AccountCode: "101", Direction: "in", BranchCode: "B1", IsActive: true}
			createPlan := command(kind, "create", "", 0)
			createPlan.Master = &planSpec
			plan := run(createPlan)
			checkPlan := func() {
				t.Helper()
				var raw bson.Raw
				if err := db.Collection(collectionName(kind)).FindOne(ctx, scopedID(scope, plan.ID)).Decode(&raw); err != nil {
					t.Fatal(err)
				}
				amount := raw.Lookup("amount")
				if amount.Type != bsontype.Decimal128 || amount.Decimal128().String() != "12.34" || raw.Lookup("fiscalyear").StringValue() != "2026" || raw.Lookup("isdeleted").Boolean() || raw.Lookup("__v").Int64() != plan.Version {
					t.Fatal("financial master changed or lost its exact amount")
				}
			}
			checkPlan()
			if count, err := db.Collection("gl_journals").CountDocuments(ctx, scopeFilter(scope)); err != nil || count != 0 {
				t.Fatalf("scale fixture must have no journals: %d %v", count, err)
			}
			before, err := p.Version(ctx, scope)
			if err != nil {
				t.Fatal(err)
			}
			for _, change := range []string{"scale"} {
				next := yearSpec
				next.Scale = 0
				cmd := command("fiscal-years", "update", oldYear.ID, oldYear.Version)
				cmd.FiscalYear = &next
				if _, err := store.Execute(ctx, scope, cmd); err == nil {
					t.Fatalf("accepted %s change with persisted %s", change, kind)
				}
				var persisted FiscalYear
				if err := db.Collection("fiscal_year").FindOne(ctx, scopedID(scope, oldYear.ID)).Decode(&persisted); err != nil || persisted.Scale != 2 || persisted.Version != oldYear.Version {
					t.Fatalf("rejected year update changed Mongo: %+v %v", persisted, err)
				}
				checkPlan()
			}
			if after, err := p.Version(ctx, scope); err != nil || after != before {
				t.Fatalf("rejected updates advanced projection: %d != %d %v", after, before, err)
			}

			// Close the year through the real process, with only balance-sheet balances.
			createJournal := command("journals", "create", "", 0)
			createJournal.Journal = &Journal{DocNo: "CAPITAL", Description: "ทดสอบเงินทุน", Date: "2026-09-11", BookCode: "JV", FiscalYear: "2026", Kind: "manual", BranchCode: "B1", Lines: []Line{{AccountCode: "101", Debit: "100"}, {AccountCode: "301", Credit: "100"}}}
			journal := run(createJournal)
			run(command("journals", "post", journal.ID, journal.Version))
			yearEnd := command("processes", "year-end", "2026", oldYear.Version)
			yearEnd.TargetYear, yearEnd.Date, yearEnd.DocNo, yearEnd.Reason = "2027", "2027-01-01", "OPEN2027", "ทดสอบป้องกันแก้แผนปีปิด"
			run(yearEnd)
			var closed FiscalYear
			if err := db.Collection("fiscal_year").FindOne(ctx, scopedID(scope, oldYear.ID)).Decode(&closed); err != nil || !closed.Closed {
				t.Fatalf("year-end did not close Mongo year: %+v %v", closed, err)
			}
			before, err = p.Version(ctx, scope)
			if err != nil {
				t.Fatal(err)
			}
			for _, action := range []string{"delete", "move", "update"} {
				cmd := command(kind, "delete", plan.ID, plan.Version)
				if action != "delete" {
					next := planSpec
					cmd.Action = "update"
					if action == "move" {
						next.FiscalYear, next.StartDate, next.EndDate = "2027", "2027-09-01", "2027-09-30"
					} else {
						next.Amount = "99.99"
					}
					cmd.Master = &next
				}
				if _, err := store.Execute(ctx, scope, cmd); err == nil {
					t.Fatalf("accepted %s of %s in closed year", action, kind)
				}
				checkPlan()
			}
			if after, err := p.Version(ctx, scope); err != nil || after != before {
				t.Fatalf("rejected plan changes advanced projection: %d != %d %v", after, before, err)
			}
		})
	}
}

func TestLedgerMongoPostgresJournalDeleteAuthorizationReplay(t *testing.T) {
	ctx, store, p, db := glAuditStore(t)
	scope := Scope{Holding: "H", Company: "C", Branch: "B1", Actor: "audit-seed-20260911"}
	serial := 0
	command := func(resource, action, id string, version int64) Command {
		serial++
		return Command{Resource: resource, Action: action, ID: id, Version: version, RequestID: fmt.Sprintf("audit-delete-20260911-%06d", serial)}
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
		cmd.Account = &Account{AccountCode: code, Names: []Name{{Code: "th", Name: "บัญชีทดสอบ " + code}}, AccountType: "asset", NormalBalance: "debit", IsActive: true, AllowPosting: true}
		run(cmd)
	}
	year := command("fiscal-years", "create", "", 0)
	year.FiscalYear = &FiscalYear{Code: "2026", StartDate: "2026-01-01", EndDate: "2026-12-31", Scale: 2, IsActive: true}
	run(year)
	period := command("periods", "create", "", 0)
	period.Master = &Master{Code: "2026", Name: "งวดทดสอบ", FiscalYear: "2026", StartDate: "2026-01-01", EndDate: "2026-12-31", IsActive: true}
	run(period)
	create := command("journals", "create", "", 0)
	create.Journal = &Journal{DocNo: "DELETE-REPLAY", Description: "ทดสอบลบซ้ำ", Date: "2026-09-11", BookCode: "JV", FiscalYear: "2026", Kind: "manual", BranchCode: "B1", Lines: []Line{{AccountCode: "101", Debit: "0.30"}, {AccountCode: "102", Credit: "0.30"}}}
	journal := run(create)
	if record, err := store.JournalForAuthorization(ctx, scope, journal.ID); err != nil || record.IsDeleted || record.BookCode != "JV" {
		t.Fatalf("live authorization record: %+v %v", record, err)
	}
	remove := command("journals", "delete", journal.ID, journal.Version)
	deleted := run(remove)
	var tombstone Journal
	if err := db.Collection("gl_journals").FindOne(ctx, scopedID(scope, journal.ID)).Decode(&tombstone); err != nil || !tombstone.IsDeleted || tombstone.Status != "void" || tombstone.Version != deleted.Version {
		t.Fatalf("delete did not persist exact tombstone: %+v %v", tombstone, err)
	}
	if _, err := p.Get(ctx, scope, "journals", journal.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("normal read exposed deleted journal: %v", err)
	}
	if record, err := store.JournalForAuthorization(ctx, scope, journal.ID); err != nil || !record.IsDeleted || record.BookCode != "JV" || record.Kind != "manual" {
		t.Fatalf("authorization lost original book/kind after delete: %+v %v", record, err)
	}
	for _, boundary := range []string{"branch", "company", "holding"} {
		other := scope
		switch boundary {
		case "branch":
			other.Branch = "B2"
		case "company":
			other.Company = "OTHER"
		case "holding":
			other.Holding = "OTHER"
		}
		if _, err := store.JournalForAuthorization(ctx, other, journal.ID); !errors.Is(err, ErrNotFound) {
			t.Fatalf("authorization crossed %s boundary: %v", boundary, err)
		}
	}
	if replay := run(remove); replay != deleted {
		t.Fatalf("delete replay changed result: %+v != %+v", replay, deleted)
	}
	var after Journal
	if err := db.Collection("gl_journals").FindOne(ctx, scopedID(scope, journal.ID)).Decode(&after); err != nil {
		t.Fatal(err)
	}
	beforeJSON, _ := json.Marshal(tombstone)
	afterJSON, _ := json.Marshal(after)
	if string(beforeJSON) != string(afterJSON) {
		t.Fatal("delete retry changed the source tombstone")
	}
	filter := scopeFilter(scope)
	filter["requestid"] = remove.RequestID
	if count, err := db.Collection("gl_events").CountDocuments(ctx, filter); err != nil || count != 1 {
		t.Fatalf("delete retry duplicated audit: %d %v", count, err)
	}
	if sequence, err := p.Version(ctx, scope); err != nil || sequence != deleted.Sequence {
		t.Fatalf("delete retry advanced projection: %d != %d %v", sequence, deleted.Sequence, err)
	}
}
