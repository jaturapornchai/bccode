//go:build integration

package services

import (
	"context"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"smlcloudplatform/internal/product/productbarcode/models"
	"smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/pkg/microservice"
)

func TestLegacyBarcodePrimarySourceIntegration(t *testing.T) {
	uri := os.Getenv("BC_OUTBOX_TEST_MONGODB_URI")
	if uri == "" {
		t.Skip("requires isolated MongoDB replica set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	// The isolated replica set has one PRIMARY and no SECONDARY. A lookup
	// inheriting this client preference fails instead of silently passing.
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri).
		SetReadPreference(readpref.Secondary()).SetServerSelectionTimeout(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { client.Disconnect(context.Background()) })
	db := client.Database("bc_barcode_primary_test_" + primitive.NewObjectID().Hex())
	t.Cleanup(func() {
		cleanup, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		if err := db.Drop(cleanup); err != nil {
			t.Error(err)
		}
	})
	id := primitive.NewObjectID()
	coll := db.Collection((models.ProductBarcodeDoc{}).CollectionName())
	if _, err := coll.InsertMany(ctx, []interface{}{
		bson.M{"_id": id, "holdingcode": "H", "businesscode": "A", "barcode": "BAR", "itemcode": "SKU", "itemunitcode": "EA"},
		bson.M{"holdingcode": "H", "businesscode": "B", "barcode": "BAR", "itemcode": "OTHER", "itemunitcode": "EA"},
	}); err != nil {
		t.Fatal(err)
	}

	repo := repositories.NewProductBarcodeRepository(primaryProjectionReads{microservice.NewPersisterMongoWithDBContext(db)}, nil)
	got, err := repo.FindByBarcodeInCompany(ctx, "H", "A", "BAR")
	if err != nil || got.ID != id {
		t.Fatalf("primary lookup failed: %v", err)
	}
	if _, err := coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"deletedat": time.Now().UTC()}}); err != nil {
		t.Fatal(err)
	}
	got, err = repo.FindByBarcodeInCompany(ctx, "H", "A", "BAR")
	if err != nil || got.ID != primitive.NilObjectID {
		t.Fatalf("deleted source was returned: %v", err)
	}
	other, err := repo.FindByBarcodeInCompany(ctx, "H", "B", "BAR")
	if err != nil || other.ID == primitive.NilObjectID {
		t.Fatalf("other company changed: %v", err)
	}
}
