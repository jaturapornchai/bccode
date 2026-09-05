package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"smlcloudplatform/internal/product/projection"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Only Product projection events belong to this dispatcher. Other outbox users
// have different payload contracts and must never be claimed here.
const aggregateType = "productprojection"
const leaseDuration = time.Minute

type Persister interface {
	Transaction(context.Context, func(context.Context) error) error
	Exec(context.Context, interface{}) (*mongo.Collection, error)
}

type Message struct {
	Topic   string          `json:"topic"`
	Key     string          `json:"key"`
	Payload json.RawMessage `json:"payload"`
}

func NewMessage(topic, key string, payload interface{}) (Message, error) {
	if topic == "" || key == "" {
		return Message{}, errors.New("outbox topic and key are required")
	}
	data, err := json.Marshal(payload)
	return Message{Topic: topic, Key: key, Payload: data}, err
}

// Payload is JSON text so BSON never coerces accounting strings or JSON numbers
// while buffering an existing event. This does not fix legacy float sources.
type Event struct {
	ID            primitive.ObjectID `bson:"_id"`
	EventUID      string             `bson:"eventuid"`
	AggregateType string             `bson:"aggregatetype"`
	AggregateUID  string             `bson:"aggregateuid"`
	Version       int64              `bson:"version"`
	EventType     string             `bson:"eventtype"`
	Payload       string             `bson:"payload"`
	Status        string             `bson:"status"`
	Attempts      int64              `bson:"attempts"`
	OccurredAt    time.Time          `bson:"occurredat"`
	PublishedAt   *time.Time         `bson:"publishedat,omitempty"`
	LastError     string             `bson:"lasterror,omitempty"`
	LeaseOwner    string             `bson:"leaseowner,omitempty"`
	LeaseUntil    time.Time          `bson:"leaseuntil,omitempty"`
}

func (Event) CollectionName() string { return "outboxevents" }

type Store struct {
	pst          Persister
	indexesMu    sync.Mutex
	indexesReady bool
}

func New(pst Persister) *Store { return &Store{pst: pst} }

func (s *Store) collection(ctx context.Context) (*mongo.Collection, error) {
	col, err := s.pst.Exec(ctx, Event{})
	if err != nil {
		return nil, err
	}
	s.indexesMu.Lock()
	defer s.indexesMu.Unlock()
	if !s.indexesReady {
		_, err = col.Indexes().CreateMany(ctx, []mongo.IndexModel{
			{Keys: bson.D{{Key: "aggregateuid", Value: 1}, {Key: "version", Value: 1}},
				Options: options.Index().SetName("uniq_product_outbox_version").SetUnique(true).SetPartialFilterExpression(bson.M{"aggregatetype": aggregateType})},
			{Keys: bson.D{{Key: "aggregatetype", Value: 1}, {Key: "status", Value: 1}, {Key: "aggregateuid", Value: 1}, {Key: "version", Value: 1}},
				Options: options.Index().SetName("product_outbox_pending")},
		})
		if err != nil {
			return nil, err
		}
		s.indexesReady = true
	}
	return col, nil
}

func AggregateKey(holding, business, guid string) string {
	return projection.AggregateKey(holding, business, guid)
}

// Commit buffers messages in the same MongoDB transaction as the business
// mutation. There is no Kafka side effect inside the transaction callback.
func (s *Store) Commit(ctx context.Context, holding, business, guid string, mutate func(context.Context) ([]Message, error)) error {
	return s.commit(ctx, holding, business, guid, AggregateKey(holding, business, guid), mutate)
}

// Barcode intents share delivery/lease machinery but have a separate sequence.
// A Product created as part of that transaction is sent as a source-refresh signal.
func BarcodeAggregateKey(holding, business, guid string) string {
	return "barcode:" + strings.TrimPrefix(AggregateKey(holding, business, guid), "product:")
}

func (s *Store) CommitBarcode(ctx context.Context, holding, business, guid string, mutate func(context.Context) ([]Message, error)) error {
	return s.commit(ctx, holding, business, guid, BarcodeAggregateKey(holding, business, guid), mutate)
}

func (s *Store) commit(ctx context.Context, holding, business, guid, key string, mutate func(context.Context) ([]Message, error)) error {
	if holding == "" || business == "" || guid == "" {
		return errors.New("outbox company and product identity are required")
	}
	col, err := s.collection(ctx)
	if err != nil {
		return err
	}
	return s.pst.Transaction(ctx, func(tx context.Context) error {
		var previous Event
		err := col.FindOne(tx, bson.M{"aggregatetype": aggregateType, "aggregateuid": key},
			options.FindOne().SetSort(bson.D{{Key: "version", Value: -1}})).Decode(&previous)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return err
		}
		messages, err := mutate(tx)
		if err != nil {
			return err
		}
		if len(messages) == 0 {
			return errors.New("outbox mutation must produce an event")
		}
		for _, m := range messages {
			if m.Topic == "" || m.Key != key || !json.Valid(m.Payload) {
				return errors.New("invalid product outbox message")
			}
		}
		payload, err := json.Marshal(messages)
		if err != nil {
			return err
		}
		id := primitive.NewObjectID()
		_, err = col.InsertOne(tx, Event{
			ID: id, EventUID: id.Hex(), AggregateType: aggregateType, AggregateUID: key,
			Version: previous.Version + 1, EventType: "product.projection", Payload: string(payload),
			Status: "PENDING", OccurredAt: time.Now().UTC(),
		})
		return err
	})
}

type SendFunc func(topic, key string, payload interface{}) error

// Dispatch sends at most the oldest pending event for each aggregate in a
// bounded page. A leased/retrying head blocks later events for that aggregate.
func (s *Store) Dispatch(ctx context.Context, send SendFunc) error {
	col, err := s.collection(ctx)
	if err != nil {
		return err
	}
	// Group before filtering leases: a leased head must block later versions,
	// but must not starve unrelated products behind a large failed backlog.
	cursor, err := col.Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"aggregatetype": aggregateType, "status": "PENDING"}}},
		{{Key: "$sort", Value: bson.D{{Key: "aggregateuid", Value: 1}, {Key: "version", Value: 1}}}},
		{{Key: "$group", Value: bson.M{"_id": "$aggregateuid", "head": bson.M{"$first": "$$ROOT"}}}},
		{{Key: "$replaceRoot", Value: bson.M{"newRoot": "$head"}}},
		{{Key: "$match", Value: bson.M{"$or": bson.A{bson.M{"leaseuntil": bson.M{"$exists": false}}, bson.M{"leaseuntil": bson.M{"$lte": time.Now().UTC()}}}}}},
		{{Key: "$sort", Value: bson.D{{Key: "occurredat", Value: 1}}}},
		{{Key: "$limit", Value: 100}},
	}, options.Aggregate().SetAllowDiskUse(true))
	if err != nil {
		return err
	}
	var events []Event
	if err = cursor.All(ctx, &events); err != nil {
		return err
	}
	seen := make(map[string]bool)
	var failures []error
	for _, event := range events {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if seen[event.AggregateUID] {
			continue
		}
		seen[event.AggregateUID] = true
		owner := primitive.NewObjectID().Hex()
		now := time.Now().UTC()
		claim := bson.M{"_id": event.ID, "status": "PENDING", "$or": bson.A{
			bson.M{"leaseuntil": bson.M{"$exists": false}}, bson.M{"leaseuntil": bson.M{"$lte": now}},
		}}
		result, err := col.UpdateOne(ctx, claim, bson.M{
			"$set": bson.M{"leaseowner": owner, "leaseuntil": now.Add(leaseDuration)}, "$inc": bson.M{"attempts": 1},
		})
		if err != nil {
			return err
		}
		if result.MatchedCount == 0 {
			continue
		}
		if err := s.deliver(ctx, col, event, owner, send); err != nil {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}

func (s *Store) deliver(ctx context.Context, col *mongo.Collection, event Event, owner string, send SendFunc) error {
	deliveryCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	owned := bson.M{"_id": event.ID, "status": "PENDING", "leaseowner": owner}
	done := make(chan struct{})
	// Renew independently while the synchronous producer awaits its delivery report.
	go func() {
		defer close(done)
		ticker := time.NewTicker(leaseDuration / 3)
		defer ticker.Stop()
		for {
			select {
			case <-deliveryCtx.Done():
				return
			case <-ticker.C:
				renewCtx, stop := context.WithTimeout(deliveryCtx, 5*time.Second)
				result, err := col.UpdateOne(renewCtx, owned, bson.M{"$set": bson.M{"leaseuntil": time.Now().UTC().Add(leaseDuration)}})
				stop()
				if err != nil || result.MatchedCount == 0 {
					cancel()
					return
				}
			}
		}
	}()
	var messages []Message
	sendErr := json.Unmarshal([]byte(event.Payload), &messages)
	if sendErr == nil && len(messages) == 0 {
		sendErr = errors.New("empty outbox payload")
	}
	if sendErr == nil {
		for _, message := range messages {
			if deliveryCtx.Err() != nil {
				sendErr = deliveryCtx.Err()
				break
			}
			if message.Topic == "" || message.Key != event.AggregateUID || !json.Valid(message.Payload) {
				sendErr = errors.New("invalid outbox payload")
				break
			}
			payload := message.Payload
			switch message.Topic {
			case "when-product-created", "when-product-updated", "when-product-deleted":
				if !strings.HasPrefix(event.AggregateUID, "barcode:") {
					payload, sendErr = projection.WithMetadata(payload, projection.Metadata{AggregateUID: event.AggregateUID, EventUID: event.EventUID, Version: event.Version})
				}
			}
			if sendErr != nil {
				break
			}
			if sendErr = send(message.Topic, message.Key, payload); sendErr != nil {
				break
			}
		}
	}
	if deliveryCtx.Err() != nil && sendErr == nil {
		sendErr = deliveryCtx.Err()
	}
	cancel()
	<-done
	if ctx.Err() != nil {
		return ctx.Err()
	} // Leave the lease for recovery after shutdown.
	update := bson.M{"$unset": bson.M{"leaseowner": "", "leaseuntil": "", "lasterror": ""}}
	if sendErr == nil {
		update["$set"] = bson.M{"status": "PUBLISHED", "publishedat": time.Now().UTC()}
	} else {
		// Persist a safe category, never a raw producer error containing credentials/payload.
		delay := time.Second * time.Duration(min(event.Attempts+1, 60))
		update = bson.M{"$set": bson.M{"lasterror": "DELIVERY_FAILED", "leaseuntil": time.Now().UTC().Add(delay)},
			"$unset": bson.M{"leaseowner": ""}}
	}
	result, err := col.UpdateOne(ctx, owned, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("product outbox delivery lease lost")
	}
	if sendErr != nil {
		return fmt.Errorf("product outbox event %s requires retry", event.EventUID)
	}
	return nil
}

func (s *Store) Run(ctx context.Context, send SendFunc, report func(error)) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		if err := s.Dispatch(ctx, send); err != nil && ctx.Err() == nil {
			report(err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
