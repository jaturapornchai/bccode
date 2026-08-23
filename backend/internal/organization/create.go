package organization

import (
	"context"
	"errors"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type organizationCreatePersister interface {
	Transaction(context.Context, func(context.Context) error) error
	Create(context.Context, interface{}, interface{}) (primitive.ObjectID, error)
	CreateIndex(context.Context, interface{}, string, interface{}) (string, error)
	CreatePartialUniqueIndex(context.Context, interface{}, string, interface{}, interface{}) (string, error)
}

type OrganizationCreate struct {
	TargetModel interface{}
	Target      interface{}
	CodeClaim   OrganizationCodeClaim
	Audit       OrganizationAudit
	Outbox      OrganizationOutboxEvent
}

type UniqueIndexSpec struct {
	Model         interface{}
	Name          string
	Keys          bson.D
	PartialFilter interface{}
}

type OrganizationCreateResponse struct {
	Entity    interface{} `json:"entity"`
	KafkaSync string      `json:"kafka_sync"`
}

// ApplyOrganizationCreate commits the permanent business-code claim, entity,
// audit, and outbox record together. The code claim is written first so its
// unique index is the concurrency fence for codes that may never be reused.
func ApplyOrganizationCreate(ctx context.Context, pst organizationCreatePersister, change OrganizationCreate) error {
	if err := validateOrganizationCreate(change); err != nil {
		return err
	}
	return pst.Transaction(ctx, func(transactionContext context.Context) error {
		if _, err := pst.Create(transactionContext, OrganizationCodeClaim{}, change.CodeClaim); err != nil {
			return err
		}
		if _, err := pst.Create(transactionContext, change.TargetModel, change.Target); err != nil {
			return err
		}
		if _, err := pst.Create(transactionContext, OrganizationAudit{}, change.Audit); err != nil {
			return err
		}
		_, err := pst.Create(transactionContext, OrganizationOutboxEvent{}, change.Outbox)
		return err
	})
}

func EnsureOrganizationCreateIndexes(ctx context.Context, pst organizationCreatePersister, targetIndexes ...UniqueIndexSpec) error {
	indexes := append([]UniqueIndexSpec{
		{Model: OrganizationCodeClaim{}, Name: "uniq_organizationcodeclaims_scope_code", Keys: bson.D{{Key: "entitytype", Value: 1}, {Key: "scopeuid", Value: 1}, {Key: "normalizedcode", Value: 1}}},
		{Model: OrganizationAudit{}, Name: "uniq_organizationaudits_audituid", Keys: bson.D{{Key: "audituid", Value: 1}}, PartialFilter: CanonicalStringFieldsFilter("audituid")},
		{Model: OrganizationOutboxEvent{}, Name: "uniq_outboxevents_eventuid", Keys: bson.D{{Key: "eventuid", Value: 1}}, PartialFilter: CanonicalStringFieldsFilter("eventuid")},
		{Model: OrganizationOutboxEvent{}, Name: "uniq_outboxevents_aggregate_version", Keys: bson.D{{Key: "aggregateuid", Value: 1}, {Key: "version", Value: 1}}, PartialFilter: CanonicalStringFieldsFilter("aggregateuid")},
	}, targetIndexes...)
	for _, index := range indexes {
		if index.Model == nil || strings.TrimSpace(index.Name) == "" || len(index.Keys) == 0 {
			return errors.New("organization unique index is incomplete")
		}
		if index.PartialFilter != nil {
			if _, err := pst.CreatePartialUniqueIndex(ctx, index.Model, index.Name, index.Keys, index.PartialFilter); err != nil {
				return err
			}
		} else if _, err := pst.CreateIndex(ctx, index.Model, index.Name, index.Keys); err != nil {
			return err
		}
	}
	return nil
}

func CanonicalStringFieldsFilter(fields ...string) bson.M {
	// Transitional DEV records without canonical stable IDs are excluded. New
	// canonical writes are still protected; the code-claim index remains global.
	filter := bson.M{}
	for _, field := range fields {
		filter[field] = bson.M{"$type": "string", "$gt": ""}
	}
	return filter
}

func validateOrganizationCreate(change OrganizationCreate) error {
	if change.TargetModel == nil || change.Target == nil ||
		strings.TrimSpace(change.CodeClaim.EntityType) == "" ||
		strings.TrimSpace(change.CodeClaim.ScopeUID) == "" ||
		strings.TrimSpace(change.CodeClaim.NormalizedCode) == "" ||
		strings.TrimSpace(change.CodeClaim.EntityUID) == "" ||
		strings.TrimSpace(change.CodeClaim.ClaimedBy) == "" ||
		change.CodeClaim.ClaimedAt.IsZero() ||
		strings.TrimSpace(change.Audit.AuditUID) == "" ||
		strings.TrimSpace(change.Audit.ActorUID) == "" ||
		strings.TrimSpace(change.Audit.Action) == "" ||
		strings.TrimSpace(change.Audit.TargetType) == "" ||
		strings.TrimSpace(change.Audit.TargetUID) == "" ||
		strings.TrimSpace(change.Audit.HoldingUID) == "" ||
		change.Audit.OccurredAt.IsZero() ||
		strings.TrimSpace(change.Outbox.EventUID) == "" ||
		strings.TrimSpace(change.Outbox.AggregateType) == "" ||
		strings.TrimSpace(change.Outbox.AggregateUID) == "" ||
		strings.TrimSpace(change.Outbox.EventType) == "" ||
		strings.TrimSpace(change.Outbox.Status) == "" ||
		change.Outbox.OccurredAt.IsZero() ||
		change.Outbox.Version < 0 {
		return errors.New("organization create identity is incomplete")
	}
	return nil
}
