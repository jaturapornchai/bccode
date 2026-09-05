//go:build integration

package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"smlcloudplatform/pkg/microservice"
)

func TestProductOutboxIntegration(t *testing.T) {
	uri := os.Getenv("BC_OUTBOX_TEST_MONGODB_URI")
	if uri == "" {
		t.Skip("set BC_OUTBOX_TEST_MONGODB_URI to an isolated replica set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		t.Fatal(err)
	}
	// Always create a new test database; never accept an existing target DB name.
	db := client.Database("bc_outbox_test_" + primitive.NewObjectID().Hex())
	t.Cleanup(func() {
		cleanup, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		if err := db.Drop(cleanup); err != nil {
			t.Error(err)
		}
		if err := client.Disconnect(cleanup); err != nil {
			t.Error(err)
		}
	})
	pst := microservice.NewPersisterMongoWithDBContext(db)
	store := New(pst)
	col, err := store.collection(ctx)
	if err != nil {
		t.Fatal(err)
	}
	enqueue := func(guid string, count int) error {
		return store.Commit(ctx, "holding", "company", guid, func(tx context.Context) ([]Message, error) {
			_, err := db.Collection("business").UpdateOne(tx, bson.M{"_id": guid}, bson.M{"$set": bson.M{"value": count}}, options.Update().SetUpsert(true))
			if err != nil {
				return nil, err
			}
			message, err := NewMessage("when-product-updated", AggregateKey("holding", "company", guid), map[string]int{"value": count})
			return []Message{message}, err
		})
	}
	expire := func(guid string) {
		_, err := col.UpdateMany(ctx, bson.M{"aggregateuid": AggregateKey("holding", "company", guid)}, bson.M{"$set": bson.M{"leaseuntil": time.Now().Add(-time.Hour)}})
		if err != nil {
			t.Fatal(err)
		}
	}
	status := func(guid string, version int64) Event {
		var event Event
		if err := col.FindOne(ctx, bson.M{"aggregateuid": AggregateKey("holding", "company", guid), "version": version}).Decode(&event); err != nil {
			t.Fatal(err)
		}
		return event
	}
	t.Run("rollback_business_when_event_invalid", func(t *testing.T) {
		err := store.Commit(ctx, "holding", "company", "rollback", func(tx context.Context) ([]Message, error) {
			_, err := db.Collection("business").InsertOne(tx, bson.M{"_id": "rollback"})
			if err != nil {
				return nil, err
			}
			return []Message{{Topic: "topic", Key: "wrong-scope", Payload: json.RawMessage("{}")}}, nil
		})
		if err == nil {
			t.Fatal("invalid event accepted")
		}
		count, err := db.Collection("business").CountDocuments(ctx, bson.M{"_id": "rollback"})
		if err != nil || count != 0 {
			t.Fatalf("business not rolled back: %d %v", count, err)
		}
		count, err = col.CountDocuments(ctx, bson.M{"aggregateuid": AggregateKey("holding", "company", "rollback")})
		if err != nil || count != 0 {
			t.Fatalf("orphan event: %d %v", count, err)
		}
	})
	t.Run("failure_retry_and_order", func(t *testing.T) {
		if err := enqueue("retry", 1); err != nil {
			t.Fatal(err)
		}
		if err := enqueue("retry", 2); err != nil {
			t.Fatal(err)
		}
		if status("retry", 1).Status != "PENDING" {
			t.Fatal("missing pending event")
		}
		if err := store.Dispatch(ctx, func(string, string, interface{}) error { return errors.New("secret must not be persisted") }); err == nil {
			t.Fatal("failure swallowed")
		}
		e := status("retry", 1)
		if e.Attempts != 1 || e.LastError != "DELIVERY_FAILED" || e.Status != "PENDING" {
			t.Fatalf("incorrect retry state: %+v", e)
		}
		var values []int
		send := func(_ string, _ string, payload interface{}) error {
			var body struct{ Value int }
			if err := json.Unmarshal(payload.(json.RawMessage), &body); err != nil {
				return err
			}
			values = append(values, body.Value)
			return nil
		}
		if err := store.Dispatch(ctx, send); err != nil {
			t.Fatal(err)
		}
		if len(values) != 0 {
			t.Fatal("later event overtook retrying head")
		}
		expire("retry")
		if err := New(pst).Dispatch(ctx, send); err != nil {
			t.Fatal(err)
		}
		if err := store.Dispatch(ctx, send); err != nil {
			t.Fatal(err)
		}
		if len(values) != 2 || values[0] != 1 || values[1] != 2 {
			t.Fatalf("wrong order %v", values)
		}
		if status("retry", 2).Status != "PUBLISHED" {
			t.Fatal("not acknowledged")
		}
	})
	t.Run("recover_after_delivery_before_ack", func(t *testing.T) {
		if err := enqueue("crash", 3); err != nil {
			t.Fatal(err)
		}
		deliveryCtx, stop := context.WithCancel(ctx)
		calls := 0
		_ = store.Dispatch(deliveryCtx, func(string, string, interface{}) error { calls++; stop(); return nil })
		if status("crash", 1).Status != "PENDING" {
			t.Fatal("cancelled delivery incorrectly acknowledged")
		}
		expire("crash")
		if err := New(pst).Dispatch(ctx, func(string, string, interface{}) error { calls++; return nil }); err != nil {
			t.Fatal(err)
		}
		if calls != 2 || status("crash", 1).Status != "PUBLISHED" {
			t.Fatal("restart did not recover")
		}
	})
	t.Run("only_one_worker_owns_head", func(t *testing.T) {
		if err := enqueue("lease", 4); err != nil {
			t.Fatal(err)
		}
		entered, release := make(chan struct{}), make(chan struct{})
		done := make(chan error, 1)
		var sends atomic.Int32
		go func() {
			done <- store.Dispatch(ctx, func(string, string, interface{}) error { sends.Add(1); close(entered); <-release; return nil })
		}()
		select {
		case <-entered:
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
		err := New(pst).Dispatch(ctx, func(string, string, interface{}) error { sends.Add(1); return nil })
		close(release)
		if err != nil {
			t.Fatal(err)
		}
		if err := <-done; err != nil {
			t.Fatal(err)
		}
		if sends.Load() != 1 {
			t.Fatal("two workers sent the same leased head")
		}
	})
	t.Run("retry_entire_product_and_barcode_intent", func(t *testing.T) {
		key := AggregateKey("holding", "company", "batch")
		err := store.Commit(ctx, "holding", "company", "batch", func(tx context.Context) ([]Message, error) {
			product, err := NewMessage("when-product-updated", key, map[string]string{"type": "product"})
			if err != nil {
				return nil, err
			}
			barcode, err := NewMessage("when-product-barcode-bulk-updated", key, []map[string]string{{"type": "barcode"}})
			return []Message{product, barcode}, err
		})
		if err != nil {
			t.Fatal(err)
		}
		var topics []string
		err = store.Dispatch(ctx, func(topic, _ string, _ interface{}) error {
			topics = append(topics, topic)
			if topic == "when-product-barcode-bulk-updated" {
				return errors.New("broker unavailable")
			}
			return nil
		})
		if err == nil || status("batch", 1).Status != "PENDING" {
			t.Fatal("partial delivery acknowledged")
		}
		expire("batch")
		if err := store.Dispatch(ctx, func(topic, _ string, _ interface{}) error { topics = append(topics, topic); return nil }); err != nil {
			t.Fatal(err)
		}
		if len(topics) != 4 || topics[0] != topics[2] || topics[1] != topics[3] || status("batch", 1).Status != "PUBLISHED" {
			t.Fatal("incomplete replay of linked intent")
		}
	})
	t.Run("ignore_other_outbox_contracts", func(t *testing.T) {
		_, err := col.InsertOne(ctx, bson.M{"aggregatetype": "company", "aggregateuid": "other", "version": 1, "status": "PENDING"})
		if err != nil {
			t.Fatal(err)
		}
		if err := store.Dispatch(ctx, func(string, string, interface{}) error { t.Error("dispatched unrelated event"); return nil }); err != nil {
			t.Fatal(err)
		}
	})
}
