//go:build integration

package kafka

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/lib/pq"
	kafkago "github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/mykafkaconsumer"
	applogger "smlcloudplatform/internal/logger"
	"smlcloudplatform/internal/product/product/outbox"
	"smlcloudplatform/pkg/microservice"
)

type projectionTestLogger struct{ applogger.ILogger }

func (projectionTestLogger) Error(...interface{}) {}
func (projectionTestLogger) Debug(...interface{}) {}

type commitObservedReader struct {
	*kafkago.Reader
	cancel  context.CancelFunc
	commits int
}

func (r *commitObservedReader) CommitMessages(ctx context.Context, messages ...kafkago.Message) error {
	err := r.Reader.CommitMessages(ctx, messages...)
	if err == nil {
		r.commits++
		r.cancel()
	}
	return err
}

func TestProjectionKafkaIntegration(t *testing.T) {
	uri, dsn, broker := os.Getenv("BC_OUTBOX_TEST_MONGODB_URI"), os.Getenv("BC_BARCODE_TEST_POSTGRES_DSN"), os.Getenv("BC_PROJECTION_TEST_KAFKA")
	if uri == "" || dsn == "" || broker == "" {
		t.Skip("requires isolated MongoDB, PostgreSQL and Kafka")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
	defer cancel()
	id := primitive.NewObjectID().Hex()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatal(err)
	}
	mongoDB := client.Database("bc_projection_test_" + id)
	t.Cleanup(func() {
		clean, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		if err := mongoDB.Drop(clean); err != nil {
			t.Error(err)
		}
		client.Disconnect(clean)
	})
	if err := client.Ping(ctx, nil); err != nil {
		t.Fatal(err)
	}
	admin, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { admin.Close() })
	schema := "bc_projection_test_" + id
	if _, err = admin.ExecContext(ctx, "CREATE SCHEMA "+pq.QuoteIdentifier(schema)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec("DROP SCHEMA " + pq.QuoteIdentifier(schema) + " CASCADE"); err != nil {
			t.Error(err)
		}
	})
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	db, err := sql.Open("postgres", u.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	for _, ddl := range []string{
		"CREATE TABLE product (holding_code text,businesscode text,itemcode text,name0 text NOT NULL,unitcode text,unitname text,balanceqty numeric(20,4) NOT NULL DEFAULT 0,PRIMARY KEY(holding_code,businesscode,itemcode))",
		"CREATE TABLE productbarcode (holding_code text,businesscode text,barcode text,itemcode text,name0 text,checksum text,groupcode text,groupnames text,unitcode text,unitname text,price1 numeric(20,4),barcoderefunitstand numeric(20,4),barcoderefunitdivide numeric(20,4),itemtype integer,balanceqty numeric(20,4) NOT NULL DEFAULT 0,PRIMARY KEY(holding_code,businesscode,barcode))",
	} {
		if _, err := db.ExecContext(ctx, ddl); err != nil {
			t.Fatal(err)
		}
	}
	conn, err := kafkago.DialContext(ctx, "tcp", broker)
	if err != nil {
		t.Fatal(err)
	}
	controller, err := conn.Controller()
	conn.Close()
	if err != nil {
		t.Fatal(err)
	}
	control, err := kafkago.DialContext(ctx, "tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		t.Fatal(err)
	}
	defer control.Close()
	topicName := func(topic string) string { return "bc-projection-" + id + "-" + topic }
	topics := []string{"when-product-created", "when-product-updated", "when-product-deleted", "when-product-barcode-created", "when-product-barcode-updated", "when-product-barcode-deleted", "when-product-barcode-bulk-updated"}
	configs := make([]kafkago.TopicConfig, 0, len(topics))
	for _, topic := range topics {
		configs = append(configs, kafkago.TopicConfig{Topic: topicName(topic), NumPartitions: 1, ReplicationFactor: 1})
	}
	if err := control.CreateTopics(configs...); err != nil {
		t.Fatal(err)
	}
	// The compose broker is dedicated to this suite; no production topic is used.
	producer := microservice.NewProducerWithTimeout(broker, "plaintext", "", "", "", projectionTestLogger{}, 10*time.Second)
	defer producer.Close()
	send := func(topic, key string, payload interface{}) error {
		return producer.SendMessage(topicName(topic), key, payload)
	}
	t.Cleanup(func() {
		c, err := kafkago.DialContext(context.Background(), "tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
		if err != nil {
			t.Error(err)
			return
		}
		defer c.Close()
		names := make([]string, 0, len(topics))
		for _, topic := range topics {
			names = append(names, topicName(topic))
		}
		if err := c.DeleteTopics(names...); err != nil {
			t.Error(err)
		}
	})
	store := outbox.New(microservice.NewPersisterMongoWithDBContext(mongoDB))
	source := mongoProjectionSource{mongoDB}
	signal := func(guid string) MongoProductModel {
		return MongoProductModel{HoldingCode: "H", BusinessCode: "A", GuidFixed: guid, Code: "SKU", UnitCode: "PCS"}
	}
	queue := func(guid, topic, name string, deleted bool) {
		t.Helper()
		p := signal(guid)
		p.Names = append(p.Names, struct {
			Name *string `json:"name" bson:"name"`
		}{&name})
		err := store.Commit(ctx, "H", "A", guid, func(tx context.Context) ([]outbox.Message, error) {
			if deleted {
				_, err := mongoDB.Collection("products").UpdateOne(tx, bson.M{"holdingcode": "H", "businesscode": "A", "guidfixed": guid}, bson.M{"$set": bson.M{"deletedat": time.Now().UTC()}})
				if err != nil {
					return nil, err
				}
			} else {
				_, err := mongoDB.Collection("products").UpdateOne(tx, bson.M{"holdingcode": "H", "businesscode": "A", "guidfixed": guid}, bson.M{"$set": p}, options.Update().SetUpsert(true))
				if err != nil {
					return nil, err
				}
			}
			m, err := outbox.NewMessage(topic, outbox.AggregateKey("H", "A", guid), p)
			return []outbox.Message{m}, err
		})
		if err != nil {
			t.Fatal(err)
		}
		var persisted bson.M
		if err := mongoDB.Collection("products").FindOne(ctx, bson.M{"guidfixed": guid}).Decode(&persisted); err != nil {
			t.Fatal(err)
		}
		if (persisted["deletedat"] != nil) != deleted {
			t.Fatal("MongoDB lifecycle mismatch immediately after mutation")
		}
	}
	dispatch := func() {
		t.Helper()
		if err := store.Dispatch(ctx, send); err != nil {
			t.Fatal(err)
		}
	}
	deliver := func(topic string, handler mykafkaconsumer.ConsumerHandleFunc, wantFailure bool) {
		t.Helper()
		readCtx, stop := context.WithTimeout(ctx, 20*time.Second)
		defer stop()
		reader := &commitObservedReader{Reader: kafkago.NewReader(kafkago.ReaderConfig{Brokers: []string{broker}, Topic: topicName(topic), GroupID: "bc-projection-" + id + "-" + topic, StartOffset: kafkago.FirstOffset, CommitInterval: 0, MinBytes: 1, MaxBytes: 1000000, MaxWait: 100 * time.Millisecond}), cancel: stop}
		defer reader.Close()
		err := mykafkaconsumer.ConsumeProjectionMessages(readCtx, reader, handler)
		if wantFailure {
			if err == nil || reader.commits != 0 {
				t.Fatalf("failed handler committed: %v commits=%d", err, reader.commits)
			}
		} else if reader.commits != 1 || !errors.Is(err, context.Canceled) {
			t.Fatalf("delivery/ack: %v commits=%d", err, reader.commits)
		}
	}
	productHandler := func(msg string) error {
		var p MongoProductModel
		if err := json.Unmarshal([]byte(msg), &p); err != nil {
			return err
		}
		return reconcileProductSignal(ctx, db, source, p)
	}
	checkProduct := func(want string) {
		t.Helper()
		var name string
		err := db.QueryRowContext(ctx, "SELECT name0 FROM product WHERE holding_code='H' AND businesscode='A' AND itemcode='SKU'").Scan(&name)
		if want == "" {
			if !errors.Is(err, sql.ErrNoRows) {
				t.Fatalf("deleted Product resurrected: %q %v", name, err)
			}
		} else if err != nil || name != want {
			t.Fatalf("Product=%q want=%q err=%v", name, want, err)
		}
	}
	t.Run("outbox_cross_topic_delete_before_create_update", func(t *testing.T) {
		queue("old", "when-product-created", "old", false)
		queue("old", "when-product-updated", "newer", false)
		queue("old", "when-product-deleted", "deleted", true)
		dispatch()
		dispatch()
		dispatch()
		count, err := mongoDB.Collection("outboxevents").CountDocuments(ctx, bson.M{"status": "PUBLISHED"})
		if err != nil || count != 3 {
			t.Fatalf("published=%d: %v", count, err)
		}
		// Process the delete first although it was published after create/update.
		deliver("when-product-deleted", productHandler, false)
		deliver("when-product-created", productHandler, false)
		deliver("when-product-updated", productHandler, false)
		checkProduct("")
		var version int64
		var tombstone bool
		if err := db.QueryRowContext(ctx, "SELECT version,deleted FROM product_projection_fences WHERE aggregateuid=$1", outbox.AggregateKey("H", "A", "old")).Scan(&version, &tombstone); err != nil || version != 3 || !tombstone {
			t.Fatalf("fence/tombstone=%d/%v: %v", version, tombstone, err)
		}
	})
	t.Run("legacy_and_recreated_identity", func(t *testing.T) {
		if err := send("when-product-created", "", signal("old")); err != nil {
			t.Fatal(err)
		}
		deliver("when-product-created", productHandler, false)
		checkProduct("")
		queue("new", "when-product-created", "recreated", false)
		dispatch()
		deliver("when-product-created", productHandler, false)
		checkProduct("recreated")
		if err := send("when-product-deleted", "", signal("old")); err != nil {
			t.Fatal(err)
		}
		deliver("when-product-deleted", productHandler, false)
		checkProduct("recreated")
		if _, err := db.ExecContext(ctx, "UPDATE product SET balanceqty='12.3400' WHERE holding_code='H' AND businesscode='A'"); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("database_failure_replays_after_reader_restart", func(t *testing.T) {
		queue("new", "when-product-updated", "retry", false)
		dispatch()
		if _, err := db.ExecContext(ctx, "ALTER TABLE product ADD CONSTRAINT reject_retry CHECK(name0<>'retry')"); err != nil {
			t.Fatal(err)
		}
		deliver("when-product-updated", productHandler, true)
		checkProduct("recreated")
		var version int64
		if err := db.QueryRowContext(ctx, "SELECT version FROM product_projection_fences WHERE aggregateuid=$1", outbox.AggregateKey("H", "A", "new")).Scan(&version); err != nil || version != 1 {
			t.Fatalf("failed write advanced version: %d %v", version, err)
		}
		if _, err := db.ExecContext(ctx, "ALTER TABLE product DROP CONSTRAINT reject_retry"); err != nil {
			t.Fatal(err)
		}
		deliver("when-product-updated", productHandler, false)
		checkProduct("retry")
		var balance string
		if err := db.QueryRowContext(ctx, "SELECT balanceqty::text FROM product WHERE holding_code='H' AND businesscode='A'").Scan(&balance); err != nil || balance != "12.3400" {
			t.Fatalf("metadata write changed stock %q: %v", balance, err)
		}
	})
	t.Run("legacy_barcode_old_snapshot_and_delete", func(t *testing.T) {
		current := models.MongoProductBarcodeModel{HoldingCode: "H", BusinessCode: "A", GuidFixed: "BAR-NEW", Barcode: "BAR", ItemCode: "SKU", ItemUnitCode: "CURRENT", StandValue: 1, DivideValue: 1}
		other := current
		other.BusinessCode = "B"
		other.ItemUnitCode = "OTHER"
		if _, err := mongoDB.Collection("productbarcodes").InsertMany(ctx, []interface{}{current, other}); err != nil {
			t.Fatal(err)
		}
		if count, err := mongoDB.Collection("productbarcodes").CountDocuments(ctx, bson.M{"barcode": "BAR"}); err != nil || count != 2 {
			t.Fatal("MongoDB create check failed")
		}
		stale := current
		stale.ItemUnitCode = "OLD"
		handler := func(msg string) error {
			var p models.MongoProductBarcodeModel
			if err := json.Unmarshal([]byte(msg), &p); err != nil {
				return err
			}
			return reconcileBarcodeSignals(ctx, db, source, []models.MongoProductBarcodeModel{p})
		}
		for _, p := range []models.MongoProductBarcodeModel{stale, other} {
			if err := send("when-product-barcode-created", "", p); err != nil {
				t.Fatal(err)
			}
			deliver("when-product-barcode-created", handler, false)
		}
		check := func(want string) {
			t.Helper()
			var unit string
			err := db.QueryRowContext(ctx, "SELECT unitcode FROM productbarcode WHERE holding_code='H' AND businesscode='A' AND barcode='BAR'").Scan(&unit)
			if want == "" {
				if !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("barcode resurrected %q %v", unit, err)
				}
			} else if err != nil || unit != want {
				t.Fatalf("barcode unit %q want %q: %v", unit, want, err)
			}
		}
		check("CURRENT")
		if _, err := mongoDB.Collection("productbarcodes").UpdateOne(ctx, bson.M{"businesscode": "A", "barcode": "BAR"}, bson.M{"$set": bson.M{"deletedat": time.Now().UTC()}}); err != nil {
			t.Fatal(err)
		}
		active, err := source.Barcodes(ctx, "H", "A", []string{"BAR"})
		if err != nil || len(active) != 0 {
			t.Fatal("MongoDB delete check failed")
		}
		if err := send("when-product-barcode-updated", "", stale); err != nil {
			t.Fatal(err)
		}
		deliver("when-product-barcode-updated", handler, false)
		check("")
		current.GuidFixed = "BAR-RECREATED"
		current.ItemUnitCode = "RECREATED"
		if _, err := mongoDB.Collection("productbarcodes").InsertOne(ctx, current); err != nil {
			t.Fatal(err)
		}
		active, err = source.Barcodes(ctx, "H", "A", []string{"BAR"})
		if err != nil || active["BAR"].GuidFixed != "BAR-RECREATED" {
			t.Fatal("MongoDB recreate check failed")
		}
		if err := send("when-product-barcode-deleted", "", stale); err != nil {
			t.Fatal(err)
		}
		deliver("when-product-barcode-deleted", handler, false)
		check("RECREATED")
		var unit string
		if err := db.QueryRowContext(ctx, "SELECT unitcode FROM productbarcode WHERE holding_code='H' AND businesscode='B' AND barcode='BAR'").Scan(&unit); err != nil || unit != "OTHER" {
			t.Fatal("other company barcode changed")
		}
	})
	t.Log("real Kafka acknowledgements, PostgreSQL 18 fences/rollback and MongoDB source checks passed")
}
