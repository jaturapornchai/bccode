//go:build integration

package kafkatransport

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	gl "smlcloudplatform/internal/generalledger"
)

type flowMQ string

func (m flowMQ) URI() string            { return string(m) }
func (flowMQ) SecurityProtocol() string { return "PLAINTEXT" }
func (flowMQ) SSLKeyFile() string       { return "" }
func (flowMQ) SSLCAFile() string        { return "" }
func (flowMQ) SSLCertFile() string      { return "" }

func flowTopic(t *testing.T, ctx context.Context, broker string) string {
	t.Helper()
	connection, err := kafka.DialContext(ctx, "tcp", broker)
	if err != nil {
		t.Fatal(err)
	}
	controller, err := connection.Controller()
	connection.Close()
	if err != nil {
		t.Fatal(err)
	}
	admin, err := kafka.DialContext(ctx, "tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		t.Fatal(err)
	}
	topic := "bc-gl-flow-" + uuid.NewString()
	if err := admin.CreateTopics(kafka.TopicConfig{Topic: topic, NumPartitions: 1, ReplicationFactor: 1}); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = admin.SetDeadline(time.Now().Add(10 * time.Second))
		if err := admin.DeleteTopics(topic); err != nil {
			t.Error(err)
		}
		admin.Close()
	})
	// CreateTopics acknowledges before the newly elected leader can serve requests.
	readyCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	for {
		leader, err := kafka.DialLeader(readyCtx, "tcp", broker, topic, 0)
		if err == nil {
			deadline, _ := readyCtx.Deadline()
			_ = leader.SetDeadline(deadline)
			_, _, err = leader.ReadOffsets()
			leader.Close()
		}
		if err == nil {
			break
		}
		if !errors.Is(err, kafka.UnknownTopicOrPartition) && !errors.Is(err, kafka.LeaderNotAvailable) && !errors.Is(err, kafka.NotLeaderForPartition) {
			t.Fatal(err)
		}
		select {
		case <-readyCtx.Done():
			t.Fatalf("new topic leader did not become ready: %v", err)
		case <-time.After(100 * time.Millisecond):
		}
	}
	return topic
}

func flowDatabases(t *testing.T, ctx context.Context) (*mongo.Database, *gl.Postgres, *sql.DB) {
	t.Helper()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("BC_GL_TEST_MONGO_URI")))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })
	db := client.Database("bc_gl_kafka_" + strings.ReplaceAll(uuid.NewString(), "-", ""))
	t.Cleanup(func() {
		if err := db.Drop(context.Background()); err != nil {
			t.Error(err)
		}
	})
	dsn := os.Getenv("BC_GL_TEST_POSTGRES_DSN")
	admin, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := "bc_gl_kafka_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := admin.ExecContext(ctx, `CREATE SCHEMA `+pq.QuoteIdentifier(schema)); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec(`DROP SCHEMA ` + pq.QuoteIdentifier(schema) + ` CASCADE`); err != nil {
			t.Error(err)
		}
		admin.Close()
	})
	address, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	query := address.Query()
	query.Set("search_path", schema)
	address.RawQuery = query.Encode()
	pg, err := sql.Open("postgres", address.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pg.Close() })
	projection := gl.NewPostgres(func(holding string) (*sql.DB, error) {
		if holding != "H" {
			return nil, errors.New("isolated test holding denied")
		}
		return pg, nil
	})
	return db, projection, pg
}

type flowPublisher struct {
	bus      *Bus
	db       *mongo.Database
	mu       sync.Mutex
	attempts map[int64]int
	acks     map[int64]int
	ordering error
}

func (p *flowPublisher) Publish(ctx context.Context, event gl.Event) error {
	p.mu.Lock()
	p.attempts[event.Sequence]++
	p.mu.Unlock()
	count, err := p.db.Collection("gl_events").CountDocuments(ctx, bson.M{"holdingcode": event.HoldingCode, "businesscode": event.BusinessCode, "delivered": false, "sequence": bson.M{"$lt": event.Sequence}})
	if err != nil {
		return err
	}
	if count != 0 {
		orderingErr := fmt.Errorf("sequence %d published before its predecessor completed", event.Sequence)
		p.mu.Lock()
		p.ordering = orderingErr
		p.mu.Unlock()
		return orderingErr
	}
	if err := p.bus.Publish(ctx, event); err != nil {
		return err
	}
	p.mu.Lock()
	p.acks[event.Sequence]++
	p.mu.Unlock()
	return nil
}

type flowProjection struct {
	*gl.Postgres
	failOnce atomic.Bool
	failed   chan gl.Event
}

func (p *flowProjection) Project(ctx context.Context, event gl.Event) error {
	if err := p.Postgres.Project(ctx, event); err != nil {
		return err
	}
	if p.failOnce.CompareAndSwap(true, false) {
		p.failed <- event
		return errors.New("synthetic failure after PostgreSQL commit before Mongo acknowledgement")
	}
	return nil
}

type flowReader struct {
	messageReader
	db         *mongo.Database
	projection *gl.Postgres
	fetched    *sync.Map
	committed  *sync.Map
	violations chan error
}

func (r flowReader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	message, err := r.messageReader.FetchMessage(ctx)
	if err == nil {
		ref, decodeErr := DecodeReference(message)
		if decodeErr != nil {
			return message, decodeErr
		}
		r.fetched.Store(ref.EventID, message.Offset)
	}
	return message, err
}

func (r flowReader) CommitMessages(ctx context.Context, messages ...kafka.Message) error {
	for _, message := range messages {
		ref, err := DecodeReference(message)
		if err != nil {
			return err
		}
		var event gl.Event
		err = r.db.Collection("gl_events").FindOne(ctx, bson.M{"_id": ref.EventID}).Decode(&event)
		if err == nil && !event.Delivered {
			err = errors.New("Kafka offset committed before Mongo delivery acknowledgement")
		}
		if err == nil {
			version, versionErr := r.projection.Version(ctx, gl.Scope{Holding: ref.HoldingCode, Company: ref.BusinessCode})
			err = versionErr
			if err == nil && version < ref.Sequence {
				err = errors.New("Kafka offset committed before PostgreSQL sequence")
			}
		}
		if err != nil {
			r.violations <- err
			return err
		}
	}
	if err := r.messageReader.CommitMessages(ctx, messages...); err != nil {
		return err
	}
	for _, message := range messages {
		ref, _ := DecodeReference(message)
		r.committed.Store(ref.EventID, message.Offset+1)
	}
	return nil
}

func flowOffset(ctx context.Context, broker, group, topic string) (int64, error) {
	client := kafka.Client{Addr: kafka.TCP(broker), Timeout: 5 * time.Second}
	response, err := client.OffsetFetch(ctx, &kafka.OffsetFetchRequest{GroupID: group, Topics: map[string][]int{topic: {0}}})
	if err != nil {
		return 0, err
	}
	if response.Error != nil {
		return 0, response.Error
	}
	partitions := response.Topics[topic]
	if len(partitions) != 1 {
		return -1, nil
	}
	return partitions[0].CommittedOffset, partitions[0].Error
}

func TestLedgerKafkaMongoPostgresFlow(t *testing.T) {
	for _, name := range []string{"BC_GL_TEST_KAFKA", "BC_GL_TEST_MONGO_URI", "BC_GL_TEST_POSTGRES_DSN"} {
		if os.Getenv(name) == "" {
			t.Skip("set " + name + " to an isolated test service")
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 160*time.Second)
	defer cancel()
	db, postgres, sqlDB := flowDatabases(t, ctx)
	broker := os.Getenv("BC_GL_TEST_KAFKA")
	topic := flowTopic(t, ctx, broker)
	group := topic + "-group"
	bus, err := NewForTopic(flowMQ(broker), group, topic)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = bus.Close() })
	publisher := &flowPublisher{bus: bus, db: db, attempts: map[int64]int{}, acks: map[int64]int{}}
	projection := &flowProjection{Postgres: postgres, failed: make(chan gl.Event, 1)}
	store := gl.NewStore(db, projection, publisher)
	scope := gl.Scope{Holding: "H", Company: "C", Branch: "B1", Actor: "kafka-flow-test"}
	violations := make(chan error, 8)
	var fetched, committed sync.Map
	originalReader := bus.newReader
	bus.newReader = func() messageReader {
		return flowReader{originalReader(), db, postgres, &fetched, &committed, violations}
	}
	wait := func(label string, condition func() (bool, error)) {
		t.Helper()
		deadline := time.NewTimer(25 * time.Second)
		defer deadline.Stop()
		ticker := time.NewTicker(25 * time.Millisecond)
		defer ticker.Stop()
		for {
			if ok, err := condition(); err != nil {
				t.Fatalf("%s: %v", label, err)
			} else if ok {
				return
			}
			select {
			case err := <-violations:
				t.Fatal(err)
			case <-deadline.C:
				t.Fatalf("timed out: %s", label)
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			case <-ticker.C:
			}
		}
	}
	serial := 0
	command := func(resource, action, id string, version int64) gl.Command {
		serial++
		return gl.Command{Resource: resource, Action: action, ID: id, Version: version, RequestID: fmt.Sprintf("kafka-flow-20260911-%06d", serial)}
	}
	execute := func(cmd gl.Command) gl.Result {
		t.Helper()
		result, err := store.Execute(ctx, scope, cmd)
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	loadEvent := func(sequence int64) gl.Event {
		t.Helper()
		var event gl.Event
		if err := db.Collection("gl_events").FindOne(ctx, bson.M{"holdingcode": "H", "businesscode": "C", "sequence": sequence}).Decode(&event); err != nil {
			t.Fatal(err)
		}
		return event
	}
	checkReady := func(sequence int64) {
		t.Helper()
		wait("Mongo acknowledgement and PostgreSQL readiness", func() (bool, error) {
			if err := store.Ready(ctx, scope); err != nil {
				if errors.Is(err, gl.ErrProjectionPending) {
					return false, nil
				}
				return false, err
			}
			version, err := postgres.Version(ctx, scope)
			return version == sequence, err
		})
	}
	for _, spec := range []struct{ code, kind, normal string }{{"101", "asset", "debit"}, {"401", "income", "credit"}} {
		cmd := command("accounts", "create", "", 0)
		cmd.Account = &gl.Account{AccountCode: spec.code, Names: []gl.Name{{Code: "th", Name: "บัญชี Kafka สมมติ"}}, AccountType: spec.kind, NormalBalance: spec.normal, IsActive: true, AllowPosting: true}
		result := execute(cmd)
		if !result.ProjectionPending || loadEvent(result.Sequence).Delivered {
			t.Fatal("broker publish acknowledged PostgreSQL before a consumer ran")
		}
		var source gl.Account
		if err := db.Collection("chart_of_accounts").FindOne(ctx, bson.M{"_id": result.ID}).Decode(&source); err != nil || source.AccountCode != spec.code {
			t.Fatalf("Mongo source missing immediately after create: %+v %v", source, err)
		}
	}
	wait("first Kafka broker acknowledgement", func() (bool, error) {
		if err := store.DeliverPending(ctx); err != nil {
			return false, err
		}
		publisher.mu.Lock()
		defer publisher.mu.Unlock()
		return publisher.acks[1] > 0, nil
	})
	publisher.mu.Lock()
	secondPublished := publisher.attempts[2]
	publisher.mu.Unlock()
	if secondPublished != 0 {
		t.Fatal("sequence 2 overtook unacknowledged sequence 1")
	}
	if version, err := postgres.Version(ctx, scope); err != nil || version != 0 {
		t.Fatalf("Kafka ACK changed PostgreSQL without a consumer: %d %v", version, err)
	}
	if !errors.Is(store.Ready(ctx, scope), gl.ErrProjectionPending) {
		t.Fatal("pending source exposed reports")
	}
	observer := kafka.NewReader(kafka.ReaderConfig{Brokers: []string{broker}, Topic: topic, Partition: 0, MinBytes: 1, MaxBytes: 4096})
	message, err := observer.ReadMessage(ctx)
	observer.Close()
	if err != nil {
		t.Fatal(err)
	}
	firstRef, err := DecodeReference(message)
	if err != nil || firstRef.Sequence != 1 || string(message.Key) != gl.ProjectionKey("H", "C") {
		t.Fatalf("broker did not receive canonical head reference: %+v %v", firstRef, err)
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(message.Value, &payload); err != nil || len(payload) != 6 || payload["changes"] != nil {
		t.Fatal("Kafka carried financial payload instead of compact source reference")
	}
	var failAfterMongo atomic.Int64
	afterMongoFailure := make(chan gl.EventReference, 1)
	reported := make(chan error, 32)
	startConsumer := func() func() {
		consumerCtx, stop := context.WithCancel(ctx)
		finished := make(chan struct{})
		go func() {
			defer close(finished)
			bus.Run(consumerCtx, func(c context.Context, ref gl.EventReference) error {
				if err := store.ApplyReference(c, ref); err != nil {
					return err
				}
				if failAfterMongo.CompareAndSwap(ref.Sequence, 0) {
					afterMongoFailure <- ref
					return errors.New("synthetic failure after Mongo acknowledgement before Kafka commit")
				}
				return nil
			}, func(err error) { reported <- err })
		}()
		return func() {
			stop()
			select {
			case <-finished:
			case <-time.After(15 * time.Second):
				t.Fatal("Kafka consumer did not stop")
			}
		}
	}
	stopConsumer := startConsumer()
	t.Cleanup(func() { stopConsumer() })
	relayCtx, stopRelay := context.WithCancel(ctx)
	relayDone := make(chan struct{})
	go func() { defer close(relayDone); store.Run(relayCtx) }()
	t.Cleanup(func() { stopRelay(); <-relayDone })
	checkReady(2)
	year := command("fiscal-years", "create", "", 0)
	year.FiscalYear = &gl.FiscalYear{Code: "2026", StartDate: "2026-01-01", EndDate: "2026-12-31", Scale: 8, IsActive: true}
	checkReady(execute(year).Sequence)
	period := command("periods", "create", "", 0)
	period.Master = &gl.Master{Code: "2026", Name: "งวด Kafka สมมติ", FiscalYear: "2026", StartDate: "2026-01-01", EndDate: "2026-12-31", IsActive: true}
	checkReady(execute(period).Sequence)
	journalCommand := func(docno, first, second, total string) gl.Command {
		cmd := command("journals", "create", "", 0)
		cmd.Journal = &gl.Journal{DocNo: docno, Date: "2026-09-11", Description: "ข้อมูล Kafka สมมติ", BookCode: "JV", FiscalYear: "2026", BranchCode: "B1", Kind: "manual", Lines: []gl.Line{{AccountCode: "101", Debit: gl.Amount(first)}, {AccountCode: "101", Debit: gl.Amount(second)}, {AccountCode: "401", Credit: gl.Amount(total)}}}
		return cmd
	}
	checkJournal := func(result gl.Result, first string, status string) {
		t.Helper()
		var raw bson.Raw
		if err := db.Collection("gl_journals").FindOne(ctx, bson.M{"_id": result.ID}).Decode(&raw); err != nil {
			t.Fatal(err)
		}
		if raw.Lookup("lines").Array().Index(0).Value().Document().Lookup("debit").Type != bsontype.Decimal128 {
			t.Fatal("Mongo money is not Decimal128")
		}
		var journal gl.Journal
		if err := bson.Unmarshal(raw, &journal); err != nil || string(journal.Lines[0].Debit) != first || journal.Status != status || journal.Version != result.Version {
			t.Fatalf("Mongo journal round-trip failed: %+v %v", journal, err)
		}
	}
	first := execute(journalCommand("KAFKA-EXACT", "0.1", "0.2", "0.3"))
	checkJournal(first, "0.1", "draft")
	checkReady(first.Sequence)
	projection.failOnce.Store(true)
	first = execute(command("journals", "post", first.ID, first.Version))
	checkJournal(first, "0.1", "posted")
	select {
	case failed := <-projection.failed:
		if failed.Sequence != first.Sequence || loadEvent(failed.Sequence).Delivered {
			t.Fatal("post-PG failure acknowledged Mongo")
		}
		if version, err := postgres.Version(ctx, scope); err != nil || version != first.Sequence {
			t.Fatalf("failed apply did not reach committed PG state: %d %v", version, err)
		}
		if _, exists := committed.Load(failed.ID); exists {
			t.Fatal("failed apply committed its Kafka offset")
		}
		if !errors.Is(store.Ready(ctx, scope), gl.ErrProjectionPending) {
			t.Fatal("equal PG sequence bypassed pending Mongo acknowledgement")
		}
	case <-time.After(20 * time.Second):
		t.Fatal("post-PG failure hook did not execute")
	}
	checkReady(first.Sequence)
	stopConsumer()
	stopConsumer = func() {}
	second := execute(journalCommand("KAFKA-SMALL", "0.00000001", "0.00000002", "0.00000003"))
	checkJournal(second, "0.00000001", "draft")
	postSecond := command("journals", "post", second.ID, second.Version)
	second = execute(postSecond)
	checkJournal(second, "0.00000001", "posted")
	if !second.ProjectionPending || loadEvent(second.Sequence).Delivered {
		t.Fatal("stopped consumer acknowledged a new journal")
	}
	if version, err := postgres.Version(ctx, scope); err != nil || version != first.Sequence {
		t.Fatalf("stopped consumer changed PG: %d %v", version, err)
	}
	failAfterMongo.Store(second.Sequence)
	stopConsumer = startConsumer()
	var failedReference gl.EventReference
	select {
	case failedReference = <-afterMongoFailure:
		if failedReference.Sequence != second.Sequence || !loadEvent(second.Sequence).Delivered {
			t.Fatal("after-Mongo failure did not preserve source completion")
		}
		if _, exists := committed.Load(failedReference.EventID); exists {
			t.Fatal("handler error advanced Kafka offset")
		}
	case <-time.After(25 * time.Second):
		t.Fatal("pending journal did not recover through restarted Kafka consumer")
	}
	stopConsumer()
	stopConsumer = func() {}
	failedOffset, ok := fetched.Load(failedReference.EventID)
	if !ok {
		t.Fatal("failed reference offset was not observed")
	}
	if offset, err := flowOffset(ctx, broker, group, topic); err != nil || offset > failedOffset.(int64) {
		t.Fatalf("broker committed failed offset: %d > %v: %v", offset, failedOffset, err)
	}
	stopConsumer = startConsumer()
	wait("replay commits the original Kafka offset", func() (bool, error) {
		_, found := committed.Load(failedReference.EventID)
		return found, nil
	})
	checkReady(second.Sequence)
	report, err := postgres.Report(ctx, scope, "trialbalance", gl.ReportQuery{FiscalYear: "2026", From: "2026-01-01", To: "2026-12-31"})
	if err != nil || report.Totals["debit"] != "0.30000003" || report.Totals["credit"] != "0.30000003" || report.Totals["difference"] != "0.00000000" {
		t.Fatalf("Kafka→Mongo→PG exact totals: %+v %v", report.Totals, err)
	}
	var lines, events int64
	if err := sqlDB.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM gl_lines),(SELECT count(*) FROM gl_events)`).Scan(&lines, &events); err != nil || lines != 6 || events != second.Sequence {
		t.Fatalf("Kafka replay duplicated ledger: lines=%d events=%d %v", lines, events, err)
	}
	if replay := execute(postSecond); replay.ID != second.ID || replay.Version != second.Version || replay.Sequence != second.Sequence || replay.ProjectionPending {
		t.Fatalf("command replay changed result: %+v", replay)
	}
	publisher.mu.Lock()
	orderingErr := publisher.ordering
	publisher.mu.Unlock()
	if orderingErr != nil {
		t.Fatal(orderingErr)
	}
	select {
	case err := <-violations:
		t.Fatal(err)
	default:
	}
	stopConsumer()
	stopConsumer = func() {}
	validReference, err := gl.ReferenceFor(loadEvent(second.Sequence))
	if err != nil {
		t.Fatal(err)
	}
	for _, mutation := range []string{"forged-hash", "wrong-company", "wrong-holding"} {
		t.Run(mutation, func(t *testing.T) {
			bad := validReference
			switch mutation {
			case "forged-hash":
				bad.EventHash = strings.Repeat("f", 64)
			case "wrong-company":
				bad.BusinessCode = "OTHER"
			case "wrong-holding":
				bad.HoldingCode = "OTHER"
			}
			badTopic := flowTopic(t, ctx, broker)
			badGroup := badTopic + "-group"
			badBus, err := NewForTopic(flowMQ(broker), badGroup, badTopic)
			if err != nil {
				t.Fatal(err)
			}
			defer badBus.Close()
			encoded, err := json.Marshal(bad)
			if err != nil {
				t.Fatal(err)
			}
			// Use the bus's own transport; DefaultTransport caches previously
			// discovered topics across these isolated UUID-topic subtests.
			if err := badBus.writer.WriteMessages(ctx, kafka.Message{Key: []byte(gl.ProjectionKey(bad.HoldingCode, bad.BusinessCode)), Value: encoded}); err != nil {
				t.Fatal(err)
			}
			badCtx, cancelBad := context.WithCancel(ctx)
			badDone, rejected := make(chan struct{}), make(chan error, 1)
			go func() {
				defer close(badDone)
				badBus.Run(badCtx, func(c context.Context, ref gl.EventReference) error {
					err := store.ApplyReference(c, ref)
					if err != nil {
						select {
						case rejected <- err:
						default:
						}
					}
					return err
				}, nil)
			}()
			select {
			case <-rejected:
			case <-time.After(20 * time.Second):
				cancelBad()
				<-badDone
				t.Fatal("forged source reference was not rejected")
			}
			cancelBad()
			<-badDone
			if offset, err := flowOffset(ctx, broker, badGroup, badTopic); err != nil || offset > 0 {
				t.Fatalf("forged reference was acknowledged: %d %v", offset, err)
			}
			if version, err := postgres.Version(ctx, scope); err != nil || version != second.Sequence {
				t.Fatalf("forged reference changed PG sequence: %d %v", version, err)
			}
			var afterLines, afterEvents int64
			if err := sqlDB.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM gl_lines),(SELECT count(*) FROM gl_events)`).Scan(&afterLines, &afterEvents); err != nil || afterLines != lines || afterEvents != events {
				t.Fatalf("forged reference changed ledger rows: %d %d %v", afterLines, afterEvents, err)
			}
		})
	}
	if !t.Failed() {
		t.Log("real Kafka reference ACK, company head ordering, Mongo Decimal128, PG NUMERIC, both failure boundaries, broker offset replay, stopped consumer recovery, and forged reference rejection passed")
	}
}

func TestLedgerNilPublisherKeepsMongoPending(t *testing.T) {
	for _, name := range []string{"BC_GL_TEST_MONGO_URI", "BC_GL_TEST_POSTGRES_DSN"} {
		if os.Getenv(name) == "" {
			t.Skip("set " + name + " to an isolated test service")
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db, postgres, sqlDB := flowDatabases(t, ctx)
	store := gl.NewStore(db, postgres, nil)
	scope := gl.Scope{Holding: "H", Company: "C", Branch: "B1", Actor: "nil-publisher-test"}
	cmd := gl.Command{Resource: "accounts", Action: "create", RequestID: "nil-publisher-20260911-account", Account: &gl.Account{AccountCode: "101", Names: []gl.Name{{Code: "th", Name: "บัญชีทดสอบไม่มี publisher"}}, AccountType: "asset", NormalBalance: "debit", IsActive: true, AllowPosting: true}}
	result, err := store.Execute(ctx, scope, cmd)
	if err != nil || !result.ProjectionPending {
		t.Fatalf("Mongo command should remain durable and pending: %+v %v", result, err)
	}
	var account gl.Account
	if err := db.Collection("chart_of_accounts").FindOne(ctx, bson.M{"_id": result.ID}).Decode(&account); err != nil || account.AccountCode != "101" {
		t.Fatalf("Mongo source was not preserved: %+v %v", account, err)
	}
	var event gl.Event
	if err := db.Collection("gl_events").FindOne(ctx, bson.M{"sequence": result.Sequence}).Decode(&event); err != nil || event.Delivered {
		t.Fatalf("missing publisher acknowledged Mongo: %+v %v", event, err)
	}
	if err := store.DeliverPending(ctx); err == nil {
		t.Fatal("missing publisher silently accepted pending delivery")
	}
	if !errors.Is(store.Ready(ctx, scope), gl.ErrProjectionPending) {
		t.Fatal("missing publisher exposed stale reports")
	}
	if version, err := postgres.Version(ctx, scope); err != nil || version != 0 {
		t.Fatalf("missing publisher changed PostgreSQL: %d %v", version, err)
	}
	if replay, err := store.Execute(ctx, scope, cmd); err != nil || replay.Sequence != result.Sequence || replay.ID != result.ID || !replay.ProjectionPending {
		t.Fatalf("pending command replay changed source identity: %+v %v", replay, err)
	}
	var records, events int64
	if err := sqlDB.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM gl_records),(SELECT count(*) FROM gl_events)`).Scan(&records, &events); err != nil || records != 0 || events != 0 {
		t.Fatalf("nil publisher wrote a PG projection: records=%d events=%d %v", records, events, err)
	}
}
