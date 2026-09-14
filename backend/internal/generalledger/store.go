package generalledger

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

type Store struct {
	db         *mongo.Database
	projection Projection
	publisher  EventPublisher
	indexMu    sync.Mutex
	indexed    bool
}

type Result struct {
	CreatedJournals   int    `json:"createdjournals"`
	ID                string `json:"id"`
	Version           int64  `json:"version"`
	Sequence          int64  `json:"sequence"`
	ProjectionPending bool   `json:"projectionpending"`
}

func NewStore(db *mongo.Database, projection Projection, publisher EventPublisher) *Store {
	return &Store{db: db, projection: projection, publisher: publisher}
}
func scopeFilter(s Scope) bson.M         { return bson.M{"holdingcode": s.Holding, "businesscode": s.Company} }
func scopedID(s Scope, id string) bson.M { f := scopeFilter(s); f["_id"] = id; return f }
func digest(value []byte) string         { sum := sha256.Sum256(value); return hex.EncodeToString(sum[:]) }
func scopeID(s Scope) string             { return digest([]byte(s.Holding + "\x00" + s.Company)) }

func collectionName(kind string) string {
	switch kind {
	case "accounts":
		return "chart_of_accounts"
	case "fiscal-years":
		return "fiscal_year"
	case "journals":
		return "gl_journals"
	}
	return MasterCollections[kind]
}

func (s *Store) EnsureIndexes(ctx context.Context) error {
	s.indexMu.Lock()
	defer s.indexMu.Unlock()
	if s.indexed {
		return nil
	}
	kinds := []string{"accounts", "fiscal-years", "journals"}
	for kind := range MasterCollections {
		kinds = append(kinds, kind)
	}
	for _, kind := range kinds {
		code := "code"
		if kind == "accounts" {
			code = "accountcode"
		}
		if kind == "journals" {
			code = "docno"
		}
		keys := bson.D{{Key: "holdingcode", Value: 1}, {Key: "businesscode", Value: 1}}
		if kind == "journals" {
			keys = append(keys, bson.E{Key: "bookcode", Value: 1})
		}
		keys = append(keys, bson.E{Key: code, Value: 1})
		_, err := s.db.Collection(collectionName(kind)).Indexes().CreateMany(ctx, []mongo.IndexModel{
			{Keys: keys, Options: options.Index().SetUnique(true).SetName("gl_business_key")},
			{Keys: bson.D{{Key: "holdingcode", Value: 1}, {Key: "businesscode", Value: 1}, {Key: "isdeleted", Value: 1}}},
		})
		if err != nil {
			return err
		}
	}
	_, err := s.db.Collection("gl_events").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "holdingcode", Value: 1}, {Key: "businesscode", Value: 1}, {Key: "requestid", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "holdingcode", Value: 1}, {Key: "businesscode", Value: 1}, {Key: "sequence", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "delivered", Value: 1}, {Key: "occurredat", Value: 1}, {Key: "sequence", Value: 1}}},
		{Keys: bson.D{{Key: "holdingcode", Value: 1}, {Key: "businesscode", Value: 1}, {Key: "delivered", Value: 1}, {Key: "sequence", Value: 1}}, Options: options.Index().SetName("gl_outbox_company_pending")},
	})
	if err == nil {
		s.indexed = true
	}
	return err
}

func (s *Store) Execute(ctx context.Context, scope Scope, cmd Command) (Result, error) {
	if scope.Holding == "" || scope.Company == "" || scope.Actor == "" {
		return Result{}, fmt.Errorf("กรุณาเลือกบริษัทและเข้าสู่ระบบก่อนใช้งานบัญชี")
	}
	if len(cmd.RequestID) < 16 || len(cmd.RequestID) > 80 || !validCode(cmd.RequestID) {
		return Result{}, fmt.Errorf("รหัสคำขอไม่ถูกต้อง กรุณาลองบันทึกอีกครั้ง")
	}
	if collectionName(cmd.Resource) == "" && cmd.Resource != "processes" {
		return Result{}, fmt.Errorf("ไม่รองรับรายการบัญชีนี้")
	}
	if !contains([]string{"create", "update", "delete", "post", "reverse", "lock", "unlock", "close", "year-end", "recalculate", "reprocess"}, cmd.Action) {
		return Result{}, fmt.Errorf("ไม่รองรับคำสั่งบัญชีนี้")
	}
	if cmd.Action != "create" && cmd.ID == "" {
		return Result{}, fmt.Errorf("กรุณาเลือกรายการบัญชี")
	}
	if err := s.EnsureIndexes(ctx); err != nil {
		return Result{}, err
	}
	raw, err := json.Marshal(cmd)
	if err != nil {
		return Result{}, err
	}
	hash := digest(append([]byte(scope.Actor+"\x00"+scope.Branch+"\x00"), raw...))
	eventID := digest([]byte(scopeID(scope) + ":" + cmd.RequestID))
	var existing Event
	lookupErr := s.db.Collection("gl_events").FindOne(ctx, bson.M{"_id": eventID}).Decode(&existing)
	if lookupErr == nil {
		if existing.RequestHash != hash {
			return Result{}, fmt.Errorf("รหัสคำขอนี้ถูกใช้กับข้อมูลอื่นแล้ว")
		}
		return s.result(ctx, existing), nil
	}
	if !errors.Is(lookupErr, mongo.ErrNoDocuments) {
		return Result{}, lookupErr
	}
	if cmd.Resource == "processes" {
		cmd.prepared, err = s.prepareProcess(ctx, scope, cmd)
		if err != nil {
			return Result{}, err
		}
	}
	var committed Event
	_, err = s.db.Collection("gl_controls").UpdateOne(ctx, bson.M{"_id": scopeID(scope)}, bson.M{"$setOnInsert": bson.M{"holdingcode": scope.Holding, "businesscode": scope.Company, "sequence": int64(0)}}, options.Update().SetUpsert(true))
	if err != nil {
		return Result{}, err
	}
	session, err := s.db.Client().StartSession()
	if err != nil {
		return Result{}, err
	}
	defer session.EndSession(ctx)
	_, err = session.WithTransaction(ctx, func(tx mongo.SessionContext) (interface{}, error) {
		var existing Event
		err := s.db.Collection("gl_events").FindOne(tx, bson.M{"_id": eventID}).Decode(&existing)
		if err == nil {
			if existing.RequestHash != hash {
				return nil, fmt.Errorf("รหัสคำขอนี้ถูกใช้กับข้อมูลอื่นแล้ว")
			}
			committed = existing
			return nil, nil
		}
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		// A company guard serializes reference checks, posting, locking and year
		// changes across processes, not just within one HTTP server.
		var counter struct {
			Sequence int64 `bson:"sequence"`
		}
		err = s.db.Collection("gl_controls").FindOneAndUpdate(tx, bson.M{"_id": scopeID(scope)}, bson.M{"$inc": bson.M{"sequence": 1}}, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&counter)
		if err != nil {
			return nil, err
		}
		if cmd.prepared != nil && counter.Sequence != cmd.prepared.Sequence+1 {
			return nil, fmt.Errorf("ข้อมูลเปลี่ยนระหว่างเตรียมประมวลผล กรุณาตรวจยอดแล้วลองใหม่")
		}
		now := time.Now().UTC().Truncate(time.Millisecond)
		changes, err := s.apply(tx, scope, cmd, now)
		if err != nil {
			return nil, err
		}
		committed = Event{ID: eventID, HoldingCode: scope.Holding, BusinessCode: scope.Company, Sequence: counter.Sequence, RequestID: cmd.RequestID, RequestHash: hash, Action: cmd.Resource + ":" + cmd.Action, Actor: scope.Actor, Reason: cmd.Reason, OccurredAt: now, Changes: changes}
		_, err = s.db.Collection("gl_events").InsertOne(tx, committed)
		return nil, err
	}, options.Transaction().SetReadConcern(readconcern.Snapshot()).SetWriteConcern(writeconcern.Majority()))
	if err != nil {
		return Result{}, transactionDuplicateError(err)
	}
	return s.result(ctx, committed), nil
}

func (s *Store) result(ctx context.Context, committed Event) Result {
	result := Result{Sequence: committed.Sequence}
	if committed.Action == "processes:close" || committed.Action == "processes:year-end" {
		for _, change := range committed.Changes {
			if change.Kind == "journals" {
				result.CreatedJournals++
			}
		}
	}
	if len(committed.Changes) > 0 {
		result.ID = committed.Changes[0].ID
		var identity Identity
		_ = json.Unmarshal([]byte(committed.Changes[0].Payload), &identity)
		result.Version = identity.Version
	}
	// Mongo is already committed. Delivery failure is reported as pending and
	// retried by the worker; it never changes a saved command into a failed save.
	if err := s.deliver(ctx, committed); err != nil {
		result.ProjectionPending = true
		return result
	}
	var delivery struct {
		Delivered bool `bson:"delivered"`
	}
	err := s.db.Collection("gl_events").FindOne(ctx, bson.M{"_id": committed.ID}, options.FindOne().SetProjection(bson.M{"delivered": 1})).Decode(&delivery)
	result.ProjectionPending = err != nil || !delivery.Delivered
	return result
}

func (s *Store) deliver(ctx context.Context, event Event) error {
	if event.Delivered {
		return nil
	}
	if s.publisher == nil {
		return fmt.Errorf("ยังไม่ได้เชื่อม Kafka สำหรับระบบบัญชี")
	}
	// A later company event cannot overtake an unprocessed event. Waiting for
	// the consumer's Mongo acknowledgement also makes concurrent relays safe:
	// they may publish the same head twice, but never publish its successor early.
	filter := bson.M{"holdingcode": event.HoldingCode, "businesscode": event.BusinessCode, "delivered": false, "sequence": bson.M{"$lt": event.Sequence}}
	older, err := s.db.Collection("gl_events").CountDocuments(ctx, filter, options.Count().SetLimit(1))
	if err != nil {
		return err
	}
	if older != 0 {
		return ErrProjectionPending
	}
	if err = s.publisher.Publish(ctx, event); err != nil {
		return err
	}
	// Broker ACK is not PostgreSQL completion. The existing retry timestamp
	// bounds duplicate notifications if a consumer is slow or unavailable.
	_, err = s.db.Collection("gl_events").UpdateOne(ctx, bson.M{"_id": event.ID, "delivered": false}, bson.M{"$set": bson.M{"retryafter": time.Now().UTC().Add(2 * time.Second)}})
	return err
}

// ApplyReference is the Kafka consumer boundary. It always loads and verifies
// the immutable Mongo source; the transport cannot inject accounting payloads.
func (s *Store) ApplyReference(ctx context.Context, reference EventReference) error {
	if err := reference.Validate(); err != nil {
		return err
	}
	var event Event
	filter := bson.M{"_id": reference.EventID, "holdingcode": reference.HoldingCode, "businesscode": reference.BusinessCode, "sequence": reference.Sequence}
	if err := s.db.Collection("gl_events").FindOne(ctx, filter).Decode(&event); err != nil {
		return fmt.Errorf("โหลดเหตุการณ์บัญชีจาก MongoDB ไม่สำเร็จ: %w", err)
	}
	hash, _, err := eventDigest(event)
	if err != nil || hash != reference.EventHash {
		return fmt.Errorf("เหตุการณ์ Kafka ไม่ตรงกับต้นฉบับบัญชี")
	}
	if s.projection == nil {
		return fmt.Errorf("ยังไม่ได้เชื่อมระบบรายงาน")
	}
	// Project checks the immutable PostgreSQL hash even on a duplicate, and
	// commits its rows, audit and sequence atomically before acknowledgement.
	if err := s.projection.Project(ctx, event); err != nil {
		return err
	}
	if event.Action == "processes:recalculate" && !event.Delivered {
		rebuilder, ok := s.projection.(interface {
			Rebuild(context.Context, Scope) error
		})
		if !ok {
			return fmt.Errorf("ระบบประมวลผลยังไม่รองรับการคำนวณใหม่")
		}
		if err := rebuilder.Rebuild(ctx, Scope{Holding: event.HoldingCode, Company: event.BusinessCode, Actor: event.Actor}); err != nil {
			return err
		}
	}
	filter["requesthash"] = event.RequestHash
	updated, err := s.db.Collection("gl_events").UpdateOne(ctx, filter, bson.M{"$set": bson.M{"delivered": true}, "$unset": bson.M{"retryafter": ""}})
	if err != nil {
		return err
	}
	if updated.MatchedCount != 1 {
		return fmt.Errorf("ยืนยันการประมวลผลบัญชีไม่สำเร็จ")
	}
	return nil
}

func (s *Store) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = s.DeliverPending(ctx)
		}
	}
}

func (s *Store) DeliverPending(ctx context.Context) error {
	// One earliest event per company prevents a failing company from starving
	// delivery for other companies. Backoff metadata is outside the audit hash.
	cursor, err := s.db.Collection("gl_events").Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"delivered": false}}},
		{{Key: "$sort", Value: bson.D{{Key: "sequence", Value: 1}}}},
		{{Key: "$group", Value: bson.M{"_id": bson.M{"holding": "$holdingcode", "company": "$businesscode"}, "event": bson.M{"$first": "$$ROOT"}}}},
		{{Key: "$replaceRoot", Value: bson.M{"newRoot": "$event"}}},
		{{Key: "$match", Value: bson.M{"$or": bson.A{bson.M{"retryafter": bson.M{"$exists": false}}, bson.M{"retryafter": bson.M{"$lte": time.Now().UTC()}}}}}},
		{{Key: "$sort", Value: bson.D{{Key: "occurredat", Value: 1}}}},
		{{Key: "$limit", Value: 100}},
	})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	var first error
	for cursor.Next(ctx) {
		var e Event
		if err := cursor.Decode(&e); err != nil {
			return err
		}
		if err := s.deliver(ctx, e); err != nil {
			if first == nil {
				first = err
			}
			_, _ = s.db.Collection("gl_events").UpdateOne(ctx, bson.M{"_id": e.ID, "delivered": false}, bson.M{"$set": bson.M{"retryafter": time.Now().UTC().Add(5 * time.Second)}})
		}
	}
	if err := cursor.Err(); err != nil {
		return err
	}
	return first
}

func (s *Store) Ready(ctx context.Context, scope Scope) error {
	var counter struct {
		Sequence int64 `bson:"sequence"`
	}
	err := s.db.Collection("gl_controls").FindOne(ctx, bson.M{"_id": scopeID(scope)}).Decode(&counter)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}
	version, err := s.projection.Version(ctx, scope)
	if err != nil {
		return err
	}
	if version != counter.Sequence {
		return ErrProjectionPending
	}
	// A recalculation may have projected its audit before its rebuild completes.
	// Do not expose reports until delivery (including rebuild) is confirmed.
	f := scopeFilter(scope)
	f["delivered"] = false
	pending, err := s.db.Collection("gl_events").CountDocuments(ctx, f, options.Count().SetLimit(1))
	if err != nil {
		return err
	}
	if pending > 0 {
		return ErrProjectionPending
	}
	return nil
}

func (s *Store) save(ctx context.Context, scope Scope, kind, id, code string, record interface{}, expected int64) (Change, error) {
	col := s.db.Collection(collectionName(kind))
	if expected == 0 {
		if _, err := col.InsertOne(ctx, record); err != nil {
			return Change{}, duplicateKeyError(kind, err)
		}
	} else {
		filter := scopedID(scope, id)
		filter["__v"] = expected
		result, err := col.ReplaceOne(ctx, filter, record)
		if err != nil {
			return Change{}, duplicateKeyError(kind, err)
		}
		if result.MatchedCount != 1 {
			return Change{}, userError(CodeStaleVersion, "รายการนี้มีการแก้ไขแล้ว กรุณาโหลดข้อมูลล่าสุด")
		}
	}
	payload, err := json.Marshal(record)
	return Change{Kind: kind, ID: id, Code: code, Payload: string(payload)}, err
}

func identityFor(scope Scope, cmd Command, old Identity, now time.Time) Identity {
	if cmd.Action == "create" {
		return Identity{ID: digest([]byte(scopeID(scope) + ":" + cmd.Resource + ":" + cmd.RequestID))[:32], HoldingCode: scope.Holding, BusinessCode: scope.Company, Version: 1, CreatedAt: now, CreatedBy: scope.Actor, UpdatedAt: now, UpdatedBy: scope.Actor}
	}
	old.Version++
	old.UpdatedAt = now
	old.UpdatedBy = scope.Actor
	return old
}
