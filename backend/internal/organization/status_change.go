package organization

import (
	"context"
	"errors"
	"strings"
	"time"

	"smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/pkg/microservice"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ErrStatusChangeConflict = errors.New("organization status changed concurrently")

type StatusChange struct {
	TargetModel      interface{}
	TargetFilter     bson.M
	MembershipFilter bson.M
	TargetType       string
	TargetUID        string
	HoldingUID       string
	CompanyUID       string
	BranchUID        string
	ActorUID         string
	Reason           string
	Before           bool
	After            bool
	Version          int64
	Set              bson.M
	OccurredAt       time.Time
}

type OrganizationAudit struct {
	AuditUID   string      `bson:"audituid"`
	ActorUID   string      `bson:"actoruid"`
	Action     string      `bson:"action"`
	TargetType string      `bson:"targettype"`
	TargetUID  string      `bson:"targetuid"`
	HoldingUID string      `bson:"holdinguid"`
	CompanyUID string      `bson:"companyuid,omitempty"`
	BranchUID  string      `bson:"branchuid,omitempty"`
	Before     interface{} `bson:"before,omitempty"`
	After      interface{} `bson:"after,omitempty"`
	Reason     string      `bson:"reason,omitempty"`
	OccurredAt time.Time   `bson:"occurredat"`
}

func (OrganizationAudit) CollectionName() string { return "organizationaudits" }

type OrganizationOutboxEvent struct {
	EventUID      string      `bson:"eventuid"`
	AggregateType string      `bson:"aggregatetype"`
	AggregateUID  string      `bson:"aggregateuid"`
	Version       int64       `bson:"version"`
	EventType     string      `bson:"eventtype"`
	Payload       interface{} `bson:"payload"`
	Status        string      `bson:"status"`
	Attempts      int64       `bson:"attempts"`
	OccurredAt    time.Time   `bson:"occurredat"`
}

func (OrganizationOutboxEvent) CollectionName() string { return "outboxevents" }

type OrganizationCodeClaim struct {
	EntityType     string    `bson:"entitytype"`
	ScopeUID       string    `bson:"scopeuid"`
	NormalizedCode string    `bson:"normalizedcode"`
	EntityUID      string    `bson:"entityuid"`
	ClaimedAt      time.Time `bson:"claimedat"`
	ClaimedBy      string    `bson:"claimedby"`
}

func (OrganizationCodeClaim) CollectionName() string { return "organizationcodeclaims" }

// ApplyStatusChange makes the organization state, permission fence, audit, and
// outbox event one MongoDB transaction. Redis invalidation is only an
// optimization: every request still compares the persisted permission version.
func ApplyStatusChange(ctx context.Context, pst microservice.IPersisterMongo, change StatusChange) error {
	change.Reason = strings.TrimSpace(change.Reason)
	change.ActorUID = strings.TrimSpace(change.ActorUID)
	change.TargetUID = strings.TrimSpace(change.TargetUID)
	change.HoldingUID = strings.TrimSpace(change.HoldingUID)
	if change.Reason == "" {
		return ErrStatusReasonRequired
	}
	if change.ActorUID == "" || change.TargetUID == "" || change.HoldingUID == "" || change.TargetModel == nil {
		return errors.New("organization status change identity is incomplete")
	}
	if change.OccurredAt.IsZero() {
		change.OccurredAt = time.Now().UTC()
	} else {
		change.OccurredAt = change.OccurredAt.UTC()
	}
	if change.Version < 0 {
		return errors.New("organization status version is invalid")
	}

	targetCollection, err := pst.Exec(ctx, change.TargetModel)
	if err != nil {
		return err
	}
	membershipCollection, err := pst.Exec(ctx, &models.ShopUser{})
	if err != nil {
		return err
	}

	return pst.Transaction(ctx, func(transactionContext context.Context) error {
		filter := bson.M{"$and": bson.A{
			change.TargetFilter,
			bson.M{"isactive": change.Before, "__v": change.Version},
		}}
		set := bson.M{}
		for key, value := range change.Set {
			set[key] = value
		}
		set["isactive"] = change.After

		result, err := targetCollection.UpdateOne(transactionContext, filter, bson.M{
			"$set": set,
			"$inc": bson.M{"__v": 1},
		})
		if err != nil {
			return err
		}
		if result.MatchedCount != 1 {
			return ErrStatusChangeConflict
		}
		if _, err := membershipCollection.UpdateMany(transactionContext, change.MembershipFilter, bson.M{
			"$inc": bson.M{"permissionversion": 1},
		}); err != nil {
			return err
		}

		auditUID := primitive.NewObjectID().Hex()
		eventUID := primitive.NewObjectID().Hex()
		nextVersion := change.Version + 1
		if _, err := pst.Create(transactionContext, OrganizationAudit{}, OrganizationAudit{
			AuditUID: auditUID, ActorUID: change.ActorUID, Action: "organization.status.changed",
			TargetType: change.TargetType, TargetUID: change.TargetUID, HoldingUID: change.HoldingUID,
			CompanyUID: change.CompanyUID, BranchUID: change.BranchUID,
			Before: bson.M{"isactive": change.Before}, After: bson.M{"isactive": change.After},
			Reason: change.Reason, OccurredAt: change.OccurredAt,
		}); err != nil {
			return err
		}
		_, err = pst.Create(transactionContext, OrganizationOutboxEvent{}, OrganizationOutboxEvent{
			EventUID: eventUID, AggregateType: change.TargetType, AggregateUID: change.TargetUID,
			Version: nextVersion, EventType: "organization.status.changed",
			Payload: bson.M{
				"audituid": auditUID, "holdinguid": change.HoldingUID,
				"companyuid": change.CompanyUID, "branchuid": change.BranchUID,
				"before": change.Before, "after": change.After,
			},
			Status: "PENDING", Attempts: 0, OccurredAt: change.OccurredAt,
		})
		return err
	})
}

// ApplyMetadataUpdate updates organization metadata without allowing the
// ordinary edit path to change isactive. The version and current status fence
// make a concurrent status change win instead of being silently overwritten.
func ApplyMetadataUpdate(
	ctx context.Context,
	pst microservice.IPersisterMongo,
	targetModel interface{},
	targetFilter bson.M,
	expectedVersion int64,
	expectedActive bool,
	set bson.M,
) error {
	if targetModel == nil || expectedVersion < 0 {
		return errors.New("organization metadata update identity is incomplete")
	}
	collection, err := pst.Exec(ctx, targetModel)
	if err != nil {
		return err
	}
	safeSet := bson.M{}
	for key, value := range set {
		if key == "isactive" || key == "__v" {
			continue
		}
		safeSet[key] = value
	}
	result, err := collection.UpdateOne(ctx, bson.M{"$and": bson.A{
		targetFilter,
		bson.M{"isactive": expectedActive, "__v": expectedVersion},
	}}, bson.M{
		"$set": safeSet,
		"$inc": bson.M{"__v": 1},
	})
	if err != nil {
		return err
	}
	if result.MatchedCount != 1 {
		return ErrStatusChangeConflict
	}
	return nil
}

func HoldingMembershipStatusFilter(holdingCode string) bson.M {
	return bson.M{"holdingcode": holdingCode, "isdeleted": bson.M{"$ne": true}}
}

func CompanyMembershipStatusFilter(holdingCode, companyUID string) bson.M {
	return bson.M{
		"holdingcode":  holdingCode,
		"isdeleted":    bson.M{"$ne": true},
		"accessscopes": bson.M{"$elemMatch": bson.M{"companyuid": companyUID}},
	}
}

func BranchMembershipStatusFilter(holdingCode, companyUID, branchUID string) bson.M {
	return bson.M{
		"holdingcode": holdingCode,
		"isdeleted":   bson.M{"$ne": true},
		"$or": bson.A{
			bson.M{"accessscopes": bson.M{"$elemMatch": bson.M{"branchuid": branchUID}}},
			bson.M{"accessscopes": bson.M{"$elemMatch": bson.M{"companyuid": companyUID, "allbranches": true}}},
		},
	}
}
