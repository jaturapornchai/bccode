package models

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestProductAggregateBarcodesAreNotPersisted(t *testing.T) {
	payload, err := bson.Marshal(Product{Barcodes: []Barcodes{{Barcode: "8850000000001"}}})
	if err != nil {
		t.Fatalf("marshal product: %v", err)
	}
	var document bson.M
	if err := bson.Unmarshal(payload, &document); err != nil {
		t.Fatalf("unmarshal product: %v", err)
	}
	if _, exists := document["barcodes"]; exists {
		t.Fatal("read-time barcode aggregation leaked into persisted Product BSON")
	}
}

func TestProductVideoPosterUsesLowercaseBSONField(t *testing.T) {
	payload, err := bson.Marshal(ProductVideo{XOrder: 1, URI: "video.mp4", PosterURI: "poster.jpg"})
	if err != nil {
		t.Fatalf("marshal product video: %v", err)
	}
	var document bson.M
	if err := bson.Unmarshal(payload, &document); err != nil {
		t.Fatalf("unmarshal product video: %v", err)
	}
	if document["posteruri"] != "poster.jpg" {
		t.Fatalf("posteruri missing from ProductVideo BSON: %#v", document)
	}
}
