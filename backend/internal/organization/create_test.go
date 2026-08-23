package organization

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type createTarget struct{ UID string }

func (createTarget) CollectionName() string { return "create_targets" }

type createPersisterStub struct {
	mu           sync.Mutex
	claims       map[string]bool
	writes       []string
	failOn       string
	indexNames   []string
	partialNames []string
	transactionN int
}

func (stub *createPersisterStub) Transaction(ctx context.Context, callback func(context.Context) error) error {
	stub.mu.Lock()
	defer stub.mu.Unlock()
	stub.transactionN++
	beforeWrites := len(stub.writes)
	beforeClaims := make(map[string]bool, len(stub.claims))
	for key, value := range stub.claims {
		beforeClaims[key] = value
	}
	if err := callback(ctx); err != nil {
		stub.writes = stub.writes[:beforeWrites]
		stub.claims = beforeClaims
		return err
	}
	return nil
}

func (stub *createPersisterStub) Create(_ context.Context, model interface{}, data interface{}) (primitive.ObjectID, error) {
	kind := fmt.Sprintf("%T", model)
	if stub.failOn == kind {
		return primitive.NilObjectID, fmt.Errorf("forced %s failure", kind)
	}
	if claim, ok := data.(OrganizationCodeClaim); ok {
		key := claim.EntityType + "|" + claim.ScopeUID + "|" + claim.NormalizedCode
		if stub.claims[key] {
			return primitive.NilObjectID, mongo.WriteException{WriteErrors: mongo.WriteErrors{{Code: 11000, Message: "duplicate claim"}}}
		}
		stub.claims[key] = true
	}
	stub.writes = append(stub.writes, kind)
	return primitive.NewObjectID(), nil
}

func (stub *createPersisterStub) CreateIndex(_ context.Context, _ interface{}, name string, _ interface{}) (string, error) {
	stub.indexNames = append(stub.indexNames, name)
	return name, nil
}

func (stub *createPersisterStub) CreatePartialUniqueIndex(_ context.Context, _ interface{}, name string, _ interface{}, _ interface{}) (string, error) {
	stub.indexNames = append(stub.indexNames, name)
	stub.partialNames = append(stub.partialNames, name)
	return name, nil
}

func TestApplyOrganizationCreateRollsBackEveryWrite(t *testing.T) {
	stub := &createPersisterStub{claims: map[string]bool{}, failOn: "organization.OrganizationAudit"}
	change := validOrganizationCreate("company-1", "001")
	if err := ApplyOrganizationCreate(context.Background(), stub, change); err == nil {
		t.Fatal("expected forced audit failure")
	}
	if len(stub.writes) != 0 || len(stub.claims) != 0 {
		t.Fatalf("transaction leaked writes=%v claims=%v", stub.writes, stub.claims)
	}

	stub.failOn = ""
	if err := ApplyOrganizationCreate(context.Background(), stub, change); err != nil {
		t.Fatal(err)
	}
	if len(stub.writes) != 4 || len(stub.claims) != 1 {
		t.Fatalf("committed writes=%v claims=%v", stub.writes, stub.claims)
	}
}

func TestApplyOrganizationCreateConcurrentDuplicateHasOneWinner(t *testing.T) {
	stub := &createPersisterStub{claims: map[string]bool{}}
	results := make(chan error, 2)
	for _, uid := range []string{"company-1", "company-2"} {
		uid := uid
		go func() {
			results <- ApplyOrganizationCreate(context.Background(), stub, validOrganizationCreate(uid, "001"))
		}()
	}

	successes, duplicates := 0, 0
	for range 2 {
		err := <-results
		switch {
		case err == nil:
			successes++
		case mongo.IsDuplicateKeyError(err):
			duplicates++
		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if successes != 1 || duplicates != 1 || len(stub.writes) != 4 {
		t.Fatalf("successes=%d duplicates=%d writes=%v", successes, duplicates, stub.writes)
	}
}

func TestEnsureOrganizationCreateIndexesIncludesConcurrencyFences(t *testing.T) {
	stub := &createPersisterStub{}
	err := EnsureOrganizationCreateIndexes(context.Background(), stub, UniqueIndexSpec{
		Model: createTarget{}, Name: "uniq_create_target_uid", Keys: bson.D{{Key: "uid", Value: 1}}, PartialFilter: CanonicalStringFieldsFilter("uid"),
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"uniq_organizationcodeclaims_scope_code",
		"uniq_organizationaudits_audituid",
		"uniq_outboxevents_eventuid",
		"uniq_outboxevents_aggregate_version",
		"uniq_create_target_uid",
	}
	if fmt.Sprint(stub.indexNames) != fmt.Sprint(want) {
		t.Fatalf("indexes=%v want=%v", stub.indexNames, want)
	}
	if len(stub.partialNames) != 4 || stub.partialNames[0] != "uniq_organizationaudits_audituid" || stub.partialNames[3] != "uniq_create_target_uid" {
		t.Fatalf("partial indexes=%v", stub.partialNames)
	}
}

func validOrganizationCreate(uid, code string) OrganizationCreate {
	now := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	return OrganizationCreate{
		TargetModel: createTarget{}, Target: createTarget{UID: uid},
		CodeClaim: OrganizationCodeClaim{EntityType: "company", ScopeUID: "holding-1", NormalizedCode: code, EntityUID: uid, ClaimedBy: "user-1", ClaimedAt: now},
		Audit:     OrganizationAudit{AuditUID: "audit-" + uid, ActorUID: "user-1", Action: "company.created", TargetType: "company", TargetUID: uid, HoldingUID: "holding-1", OccurredAt: now},
		Outbox:    OrganizationOutboxEvent{EventUID: "event-" + uid, AggregateType: "company", AggregateUID: uid, EventType: "company.created", Status: "PENDING", OccurredAt: now},
	}
}
