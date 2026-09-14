//go:build integration

package generalledger

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestLedgerMongoPostgresZeroBalanceYearEnd(t *testing.T) {
	ctx, store, p, db := glAuditStore(t)
	scope := Scope{Holding: "H", Company: "C", Actor: "audit-zero-20260911"}
	var oldYear Result
	for _, code := range []string{"2026", "2027"} {
		cmd := Command{Resource: "fiscal-years", Action: "create", RequestID: "audit-zero-year-" + code, FiscalYear: &FiscalYear{Code: code, StartDate: code + "-01-01", EndDate: code + "-12-31", Currency: "THB", Scale: 2, IsActive: true}}
		result, err := store.Execute(ctx, scope, cmd)
		if err != nil || result.ProjectionPending {
			t.Fatalf("create year %s: %+v %v", code, result, err)
		}
		var persisted FiscalYear
		if err := db.Collection("fiscal_year").FindOne(ctx, scopedID(scope, result.ID)).Decode(&persisted); err != nil || persisted.Code != code || persisted.Closed {
			t.Fatalf("year create not persisted: %+v %v", persisted, err)
		}
		if code == "2026" {
			oldYear = result
		}
	}
	cmd := Command{Resource: "processes", Action: "year-end", ID: "2026", Version: oldYear.Version, TargetYear: "2027", Date: "2027-01-01", DocNo: "ZERO-OPEN", Reason: "ปิดปีไม่มียอดคงเหลือ", RequestID: "audit-zero-year-end-2026"}
	closed, err := store.Execute(ctx, scope, cmd)
	if err != nil || closed.ProjectionPending || closed.CreatedJournals != 0 {
		t.Fatalf("zero balance year-end: %+v %v", closed, err)
	}
	var persisted FiscalYear
	if err := db.Collection("fiscal_year").FindOne(ctx, scopedID(scope, oldYear.ID)).Decode(&persisted); err != nil || !persisted.Closed || persisted.Version != oldYear.Version+1 {
		t.Fatalf("zero year did not close in Mongo: %+v %v", persisted, err)
	}
	if count, err := db.Collection("gl_journals").CountDocuments(ctx, scopeFilter(scope)); err != nil || count != 0 {
		t.Fatalf("zero year created unnecessary journals: %d %v", count, err)
	}
	if replay, err := store.Execute(ctx, scope, cmd); err != nil || replay != closed {
		t.Fatalf("closed zero year replay: %+v != %+v %v", replay, closed, err)
	}
	filter := scopeFilter(scope)
	filter["requestid"] = cmd.RequestID
	if count, err := db.Collection("gl_events").CountDocuments(ctx, filter); err != nil || count != 1 {
		t.Fatalf("zero year audit duplicated: %d %v", count, err)
	}
	var event Event
	if err := db.Collection("gl_events").FindOne(ctx, bson.M{"holdingcode": "H", "businesscode": "C", "requestid": cmd.RequestID}).Decode(&event); err != nil || len(event.Changes) != 1 || event.Changes[0].Kind != "fiscal-years" {
		t.Fatalf("zero year audit must contain only year closure: %+v %v", event, err)
	}
	if sequence, err := p.Version(ctx, scope); err != nil || sequence != closed.Sequence {
		t.Fatalf("zero year replay advanced projection: %d != %d %v", sequence, closed.Sequence, err)
	}
	page, err := p.List(ctx, scope, "journals", "", 1, 10, ListFilter{Kind: "opening"})
	if err != nil || page.Total != 0 {
		t.Fatalf("zero year projected an opening journal: %+v %v", page, err)
	}
}
