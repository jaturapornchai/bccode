//go:build integration

package services

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"smlcloudplatform/internal/product/product/outbox"
	productrepo "smlcloudplatform/internal/product/product/repositories"
	"smlcloudplatform/internal/product/productbarcode/models"
	"smlcloudplatform/internal/product/productbarcode/repositories"
	unitrepo "smlcloudplatform/internal/product/unit/repositories"
	"smlcloudplatform/pkg/microservice"
)

func TestBarcodeServiceOutboxIntegration(t *testing.T) {
	uri := os.Getenv("BC_OUTBOX_TEST_MONGODB_URI")
	if uri == "" {
		t.Skip("requires isolated MongoDB replica set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatal(err)
	}
	db := client.Database("bc_barcode_outbox_test_" + primitive.NewObjectID().Hex())
	t.Cleanup(func() {
		clean, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		if err := db.Drop(clean); err != nil {
			t.Error(err)
		}
		client.Disconnect(clean)
	})
	pst := microservice.NewPersisterMongoWithDBContext(db)
	repo := repositories.NewProductBarcodeRepository(pst, nil)
	store := outbox.New(pst)
	svc := ProductBarcodeHttpService{repo: repo, repoMaster: productrepo.NewProductRepository(pst),
		repoUnit: unitrepo.NewUnitRepository(pst), eventOutbox: store, contextTimeout: 15 * time.Second}
	request := models.ProductBarcodeRequest{ProductBarcodeBase: models.ProductBarcodeBase{
		ItemCode: "SKU", Barcode: "4006381333931", ItemUnitCode: "PCS", Description: "before",
	}}
	guid, err := svc.CreateProductBarcodeInCompany("H", "A", "tester", request)
	if err != nil {
		t.Fatal(err)
	}
	read := func() models.ProductBarcodeDoc {
		t.Helper()
		var doc models.ProductBarcodeDoc
		if err := db.Collection("productbarcodes").FindOne(ctx, bson.M{"holdingcode": "H", "businesscode": "A", "guidfixed": guid}).Decode(&doc); err != nil {
			t.Fatal(err)
		}
		return doc
	}
	created := read()
	if created.ID == primitive.NilObjectID || created.Description != "before" {
		t.Fatal("create was not persisted")
	}
	key := outbox.BarcodeAggregateKey("H", "A", guid)
	count := func(collection string, filter bson.M, want int64) {
		t.Helper()
		n, err := db.Collection(collection).CountDocuments(ctx, filter)
		if err != nil || n != want {
			t.Fatalf("%s count=%d want=%d: %v", collection, n, want, err)
		}
	}
	count("units", bson.M{"holdingcode": "H", "unitcode": "PCS", "guidfixed": created.ItemUnitGuid, "unitname1": "PCS"}, 1)
	count("products", bson.M{"holdingcode": "H", "businesscode": "A", "code": "SKU"}, 1)
	count("outboxevents", bson.M{"aggregateuid": key, "status": "PENDING"}, 1)
	// No broker call occurs inside CRUD. A subsequent delivery failure retains intent.
	if err := store.Dispatch(ctx, func(string, string, interface{}) error { return errors.New("broker unavailable") }); err == nil {
		t.Fatal("expected delivery failure")
	}
	count("outboxevents", bson.M{"aggregateuid": key, "status": "PENDING"}, 1)
	if _, err := svc.CreateProductBarcodeInCompany("H", "A", "tester", request); err == nil {
		t.Fatal("duplicate accepted")
	}
	count("outboxevents", bson.M{"aggregateuid": key}, 1)

	request.Description = "after <ทดสอบ>"
	if err := svc.UpdateProductBarcodeInCompany("H", "A", guid, "tester", request); err != nil {
		t.Fatal(err)
	}
	updated := read()
	if updated.ID != created.ID || updated.Description != request.Description {
		t.Fatal("update lost identity or value")
	}
	count("outboxevents", bson.M{"aggregateuid": key}, 2)
	invalid := request
	invalid.ItemCode = "CHANGED"
	if err := svc.UpdateProductBarcodeInCompany("H", "A", guid, "tester", invalid); err == nil {
		t.Fatal("immutable identity changed")
	}
	if read().ItemCode != "SKU" {
		t.Fatal("invalid update persisted")
	}

	other := created
	other.ID = primitive.NewObjectID()
	other.GuidFixed = "OTHER"
	other.BusinessCode = "B"
	// Preserve the existing holding-wide stock guard until stock keys include company.
	if _, err := db.Collection("productbarcodes").InsertOne(ctx, other); !mongo.IsDuplicateKeyError(err) {
		t.Fatalf("holding barcode guard changed: %v", err)
	}
	other.Barcode = "OTHER-BAR"
	if _, err := db.Collection("productbarcodes").InsertOne(ctx, other); err != nil {
		t.Fatal(err)
	}
	if err := svc.DeleteProductBarcodeInCompany("H", "A", guid, "tester"); err != nil {
		t.Fatal(err)
	}
	count("productbarcodes", bson.M{"_id": created.ID, "deletedat": bson.M{"$exists": true}, "deletedby": "tester"}, 1)
	count("productbarcodes", bson.M{"_id": other.ID, "deletedat": bson.M{"$exists": false}}, 1)
	count("outboxevents", bson.M{"aggregateuid": key}, 3)

	// Simulate an outbox insert failure after Unit, Product and Barcode writes.
	// All source documents must roll back, not leave orphan masters.
	if err := db.RunCommand(ctx, bson.D{{Key: "collMod", Value: "outboxevents"}, {Key: "validator", Value: bson.M{"status": "REJECT"}}, {Key: "validationLevel", Value: "strict"}}).Err(); err != nil {
		t.Fatal(err)
	}
	failing := request
	failing.ItemCode = "ROLLBACK"
	failing.Barcode = "ROLLBACK-BAR"
	failing.ItemUnitCode = "ROLLBACK-UNIT"
	if _, err := svc.CreateProductBarcodeInCompany("H", "A", "tester", failing); err == nil {
		t.Fatal("outbox insert failure was hidden")
	}
	count("units", bson.M{"holdingcode": "H", "unitcode": "ROLLBACK-UNIT"}, 0)
	count("products", bson.M{"holdingcode": "H", "businesscode": "A", "code": "ROLLBACK"}, 0)
	count("productbarcodes", bson.M{"holdingcode": "H", "businesscode": "A", "barcode": "ROLLBACK-BAR"}, 0)
	if err := db.RunCommand(ctx, bson.D{{Key: "collMod", Value: "outboxevents"}, {Key: "validator", Value: bson.M{}}}).Err(); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Collection("outboxevents").UpdateMany(ctx, bson.M{"aggregateuid": key}, bson.M{"$unset": bson.M{"leaseuntil": ""}}); err != nil {
		t.Fatal(err)
	}

	productSignals, barcodeSignals, unitSignals := 0, 0, 0
	for i := 0; i < 3; i++ {
		if err := store.Dispatch(ctx, func(topic, gotKey string, payload interface{}) error {
			if gotKey != key {
				t.Fatal("wrong aggregate key")
			}
			raw, err := json.Marshal(payload)
			if err != nil {
				return err
			}
			var message map[string]json.RawMessage
			if err := json.Unmarshal(raw, &message); err != nil {
				return err
			}
			if _, found := message["_projection"]; found {
				t.Fatal("Barcode sequence was assigned to Product fence")
			}
			if topic == "when-product-unit-created" {
				unitSignals++
			} else if topic == "when-product-created" {
				productSignals++
			} else {
				barcodeSignals++
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	if productSignals != 1 || barcodeSignals != 3 || unitSignals != 1 {
		t.Fatalf("signals: product=%d barcode=%d unit=%d", productSignals, barcodeSignals, unitSignals)
	}
	count("outboxevents", bson.M{"aggregateuid": key, "status": "PUBLISHED"}, 3)
}
