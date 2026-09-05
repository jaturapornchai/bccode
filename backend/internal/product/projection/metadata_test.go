package projection

import (
	"encoding/json"
	"testing"
)

func TestMetadataPreservesPayloadAndExactVersion(t *testing.T) {
	raw := json.RawMessage(`{"amount":"9007199254740993.30","quantity":"0.3000","_projection":{"version":"999"}}`)
	meta := Metadata{AggregateUID: AggregateKey("H", "A", "SKU"), EventUID: "event", Version: 9007199254740993}
	encoded, err := WithMetadata(raw, meta)
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		Amount     string
		Quantity   string
		Projection Metadata `json:"_projection"`
	}
	if err := json.Unmarshal(encoded, &result); err != nil {
		t.Fatal(err)
	}
	if result.Amount != "9007199254740993.30" || result.Quantity != "0.3000" || result.Projection != meta {
		t.Fatalf("payload/metadata changed: %s", encoded)
	}
	for _, bad := range []json.RawMessage{json.RawMessage("null"), json.RawMessage("[]"), json.RawMessage("{")} {
		if _, err := WithMetadata(bad, meta); err == nil {
			t.Fatal("invalid object accepted")
		}
	}
}

func TestProjectionLocksSeparateCompaniesAndResources(t *testing.T) {
	keys := []int64{LockKey("product", "H", "A", "SKU"), LockKey("product", "H", "B", "SKU"), LockKey("product", "H", "A", ""), LockKey("barcode", "H", "A", "SKU")}
	seen := map[int64]bool{}
	for _, key := range keys {
		if seen[key] {
			t.Fatal("distinct lock scopes collided")
		}
		seen[key] = true
	}
}
