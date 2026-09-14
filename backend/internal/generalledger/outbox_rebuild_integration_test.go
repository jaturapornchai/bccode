//go:build integration

package generalledger

import (
	"context"
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

type auditRebuildFailOnce struct {
	*Postgres
	attempts int
}

func (p *auditRebuildFailOnce) Rebuild(ctx context.Context, scope Scope) error {
	p.attempts++
	if p.attempts == 1 {
		return errors.New("synthetic rebuild failure")
	}
	return p.Postgres.Rebuild(ctx, scope)
}

func TestLedgerMongoPostgresRebuildFailureReadiness(t *testing.T) {
	ctx, _, p, db := glAuditStore(t)
	projection := &auditRebuildFailOnce{Postgres: p}
	store := newSyncTestStore(db, projection)
	scope := Scope{Holding: "H", Company: "C", Actor: "audit-outbox-20260911"}
	account := Command{Resource: "accounts", Action: "create", RequestID: "audit-rebuild-account-2026", Account: &Account{AccountCode: "101", Names: []Name{{Code: "th", Name: "บัญชีทดสอบคำนวณใหม่"}}, AccountType: "asset", NormalBalance: "debit", AllowPosting: true, IsActive: true}}
	created, err := store.Execute(ctx, scope, account)
	if err != nil || created.ProjectionPending {
		t.Fatalf("seed account: %+v %v", created, err)
	}
	before, err := p.Get(ctx, scope, "accounts", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	command := Command{Resource: "processes", Action: "recalculate", ID: "C", RequestID: "audit-rebuild-fail-once-2026", Reason: "ทดสอบการส่งซ้ำเมื่อคำนวณใหม่ล้มเหลว"}
	pending, err := store.Execute(ctx, scope, command)
	if err != nil || !pending.ProjectionPending || projection.attempts != 1 {
		t.Fatalf("rebuild failure was not pending: %+v attempts=%d %v", pending, projection.attempts, err)
	}
	if sequence, err := p.Version(ctx, scope); err != nil || sequence != pending.Sequence {
		t.Fatalf("recalculate audit should already be projected: %d != %d %v", sequence, pending.Sequence, err)
	}
	filter := scopeFilter(scope)
	filter["requestid"] = command.RequestID
	var event Event
	if err := db.Collection("gl_events").FindOne(ctx, filter).Decode(&event); err != nil || event.Delivered || len(event.Changes) != 0 {
		t.Fatalf("failed rebuild audit state: %+v %v", event, err)
	}
	if err := store.Ready(ctx, scope); err == nil {
		t.Fatal("readiness accepted equal sequence while rebuild was incomplete")
	}
	if err := store.DeliverPending(ctx); err != nil || projection.attempts != 2 {
		t.Fatalf("pending delivery did not retry rebuild: attempts=%d %v", projection.attempts, err)
	}
	if err := db.Collection("gl_events").FindOne(ctx, filter).Decode(&event); err != nil || !event.Delivered {
		t.Fatalf("successful retry not marked delivered: %+v %v", event, err)
	}
	if err := store.Ready(ctx, scope); err != nil {
		t.Fatalf("readiness did not recover: %v", err)
	}
	if replay, err := store.Execute(ctx, scope, command); err != nil || replay.ProjectionPending || replay.Sequence != pending.Sequence || projection.attempts != 2 {
		t.Fatalf("delivered recalculation replay reran rebuild: %+v attempts=%d %v", replay, projection.attempts, err)
	}
	if count, err := db.Collection("gl_events").CountDocuments(ctx, bson.M{"holdingcode": "H", "businesscode": "C", "requestid": command.RequestID}); err != nil || count != 1 {
		t.Fatalf("rebuild retry duplicated audit: %d %v", count, err)
	}
	after, err := p.Get(ctx, scope, "accounts", created.ID)
	if err != nil || string(before) != string(after) {
		t.Fatalf("retry changed account truth: %s != %s %v", after, before, err)
	}
}
