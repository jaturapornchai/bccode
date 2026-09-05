//go:build integration

package services

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"smlcloudplatform/internal/product/product/models"
	"smlcloudplatform/internal/product/product/outbox"
	"smlcloudplatform/internal/product/product/repositories"
	barcodeModel "smlcloudplatform/internal/product/productbarcode/models"
	barcodeRepo "smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/pkg/microservice"
)

func TestProductServiceOutboxIntegration(t *testing.T) {
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
	db := client.Database("bc_product_test_" + primitive.NewObjectID().Hex())
	t.Cleanup(func() {
		clean, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		if err := db.Drop(clean); err != nil {
			t.Error(err)
		}
		if err := client.Disconnect(clean); err != nil {
			t.Error(err)
		}
	})
	pst := microservice.NewPersisterMongoWithDBContext(db)
	repo := repositories.NewProductRepository(pst)
	store := outbox.New(pst)
	svc := ProductHttpService{repo: repo, eventOutbox: store, repomgProductBarcode: *barcodeRepo.NewProductBarcodeRepository(pst, nil), contextTimeout: 15 * time.Second}
	doc := models.ProductDoc{}
	doc.HoldingCode = "H"
	doc.BusinessCode = "A"
	doc.Code = "SKU"
	doc.UnitCode = "PCS"
	if err := svc.Create(&doc); err != nil {
		t.Fatal(err)
	}
	original, err := repo.FindByGuidInCompany(ctx, "H", "A", doc.GuidFixed)
	if err != nil || original.ID == primitive.NilObjectID || original.Code != "SKU" {
		t.Fatalf("create not persisted: %v", err)
	}
	checkEvents := func(want int64) {
		count, err := db.Collection("outboxevents").CountDocuments(ctx, bson.M{"aggregateuid": outbox.AggregateKey("H", "A", doc.GuidFixed), "status": "PENDING"})
		if err != nil || count != want {
			t.Fatalf("pending count=%d want=%d: %v", count, want, err)
		}
	}
	checkEvents(1)
	if err := svc.Create(&doc); err == nil {
		t.Fatal("duplicate create accepted")
	}
	checkEvents(1)
	// A linked Barcode in another company must survive every operation.
	linked := barcodeModel.ProductBarcodeDoc{}
	linked.GuidFixed = "BAR-A"
	linked.HoldingCode = "H"
	linked.BusinessCode = "A"
	linked.ItemCode = "SKU"
	linked.ItemUnitCode = "PCS"
	linked.Barcode = "885001"
	other := linked
	other.GuidFixed = "BAR-B"
	other.BusinessCode = "B"
	if _, err := db.Collection(linked.CollectionName()).InsertMany(ctx, []interface{}{linked, other}); err != nil {
		t.Fatal(err)
	}
	doc.GroupCode = "UPDATED"
	doc.UnitGuid = "UNIT-NEW"
	updated, err := svc.Update("H", "A", doc.GuidFixed, "tester", &doc)
	if err != nil {
		t.Fatal(err)
	}
	persisted, err := repo.FindByGuidInCompany(ctx, "H", "A", doc.GuidFixed)
	if err != nil || persisted.ID != original.ID || persisted.GroupCode != "UPDATED" || updated.GuidFixed != original.GuidFixed {
		t.Fatal("update lost identity or data")
	}
	checkEvents(2)
	var linkedAfter, otherAfter barcodeModel.ProductBarcodeDoc
	if err := db.Collection(linked.CollectionName()).FindOne(ctx, bson.M{"guidfixed": "BAR-A"}).Decode(&linkedAfter); err != nil {
		t.Fatal(err)
	}
	if err := db.Collection(linked.CollectionName()).FindOne(ctx, bson.M{"guidfixed": "BAR-B"}).Decode(&otherAfter); err != nil {
		t.Fatal(err)
	}
	if linkedAfter.ItemUnitGuid != "UNIT-NEW" || otherAfter.ItemUnitGuid != "" {
		t.Fatal("linked snapshot missing or crossed company boundary")
	}
	var event outbox.Event
	if err := db.Collection("outboxevents").FindOne(ctx, bson.M{"aggregateuid": outbox.AggregateKey("H", "A", doc.GuidFixed), "version": 2}).Decode(&event); err != nil {
		t.Fatal(err)
	}
	var messages []outbox.Message
	if err := json.Unmarshal([]byte(event.Payload), &messages); err != nil || len(messages) != 2 {
		t.Fatal("product/barcode intent not committed together")
	}
	invalid := doc
	invalid.UnitCode = ""
	if _, err := svc.Update("H", "A", doc.GuidFixed, "tester", &invalid); err == nil {
		t.Fatal("invalid update accepted")
	}
	checkEvents(2)
	queued, err := svc.Resync("H", "A")
	if err != nil || queued != 1 {
		t.Fatalf("resync queue count=%d: %v", queued, err)
	}
	checkEvents(3)
	if err := svc.Delete("H", "A", doc.GuidFixed, "tester"); err != nil {
		t.Fatal(err)
	}
	deleted, err := repo.FindByGuidInCompany(ctx, "H", "A", doc.GuidFixed)
	if err != nil || deleted.ID != primitive.NilObjectID {
		t.Fatal("deleted product remains visible")
	}
	checkEvents(4)
	queued, err = svc.Resync("H", "A")
	if err != nil || queued != 0 {
		t.Fatalf("resync queued deleted product: %d %v", queued, err)
	}
	checkEvents(4)
	var raw bson.M
	if err := db.Collection(doc.CollectionName()).FindOne(ctx, bson.M{"_id": original.ID}).Decode(&raw); err != nil || raw["deletedat"] == nil {
		t.Fatal("soft-delete audit not preserved")
	}
	if count, err := db.Collection(linked.CollectionName()).CountDocuments(ctx, bson.M{"guidfixed": "BAR-B", "businesscode": "B"}); err != nil || count != 1 {
		t.Fatal("other company barcode lost")
	}
}
