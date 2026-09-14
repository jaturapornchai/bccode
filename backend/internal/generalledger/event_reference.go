package generalledger

import (
	"context"
	"fmt"
	"regexp"
)

// EventPublisher only acknowledges Kafka receipt. PostgreSQL completion is
// recorded separately by ApplyReference after the consumer finishes its work.
type EventPublisher interface {
	Publish(context.Context, Event) error
}

type EventPublisherFunc func(context.Context, Event) error

func (publish EventPublisherFunc) Publish(ctx context.Context, event Event) error {
	return publish(ctx, event)
}

// Kafka carries a compact reference; the complete immutable command and its
// decimal-string payload remain in MongoDB, including large year-end batches.
type EventReference struct {
	SchemaVersion int    `json:"schemaversion"`
	EventID       string `json:"eventid"`
	HoldingCode   string `json:"holdingcode"`
	BusinessCode  string `json:"businesscode"`
	Sequence      int64  `json:"sequence"`
	EventHash     string `json:"eventhash"`
}

var referenceHash = regexp.MustCompile(`^[0-9a-f]{64}$`)

func (reference EventReference) Validate() error {
	if reference.SchemaVersion != 1 || !referenceHash.MatchString(reference.EventID) || !referenceHash.MatchString(reference.EventHash) || !validCode(reference.HoldingCode) || !validCode(reference.BusinessCode) || reference.Sequence < 1 {
		return fmt.Errorf("ข้อมูลอ้างอิงเหตุการณ์บัญชีไม่ถูกต้อง")
	}
	return nil
}

func ReferenceFor(event Event) (EventReference, error) {
	hash, _, err := eventDigest(event)
	if err != nil {
		return EventReference{}, err
	}
	reference := EventReference{SchemaVersion: 1, EventID: event.ID, HoldingCode: event.HoldingCode, BusinessCode: event.BusinessCode, Sequence: event.Sequence, EventHash: hash}
	return reference, reference.Validate()
}

func ProjectionKey(holding, company string) string {
	return scopeID(Scope{Holding: holding, Company: company})
}
