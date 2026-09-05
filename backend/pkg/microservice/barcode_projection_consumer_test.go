package microservice

import "testing"

func TestBarcodeConsumersUseManagedAcknowledgedWorkers(t *testing.T) {
	ms := &Microservice{}
	topics := []string{"when-product-barcode-created", "when-product-barcode-updated", "when-product-barcode-deleted", "when-product-barcode-bulk-created", "when-product-barcode-bulk-updated", "when-product-barcode-bulk-deleted"}
	for _, topic := range topics {
		if err := ms.Consume("unused:9092", topic, "test-group", 0, func(IContext) error { return nil }); err != nil {
			t.Fatal(err)
		}
	}
	if len(ms.backgroundWorkers) != len(topics) {
		t.Fatalf("registered workers=%d", len(ms.backgroundWorkers))
	}
}

// The kafka-go Barcode readers must never join the librdkafka group that still
// serves the other legacy topics; the derived name is stable per configured group.
func TestBarcodeProjectionGroupIsSeparateAndStable(t *testing.T) {
	if got := barcodeProjectionGroup("consumer-productbarcode-group-01"); got != "consumer-productbarcode-group-01-projection" {
		t.Fatalf("unexpected group %q", got)
	}
	if barcodeProjectionGroup("a") == "a" || barcodeProjectionGroup("a") == barcodeProjectionGroup("b") {
		t.Fatal("projection group must differ from the legacy group and between groups")
	}
}
