//go:build integration

package generalledger

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

// Existing domain integration tests isolate financial behavior from broker
// scheduling. Production constructors always receive the Kafka publisher.
func newSyncTestStore(db *mongo.Database, projection Projection) *Store {
	store := NewStore(db, projection, nil)
	store.publisher = EventPublisherFunc(func(ctx context.Context, event Event) error {
		reference, err := ReferenceFor(event)
		if err != nil {
			return err
		}
		return store.ApplyReference(ctx, reference)
	})
	return store
}
