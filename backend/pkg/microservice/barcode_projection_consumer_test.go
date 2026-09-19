package microservice

import "testing"

func TestBarcodeConsumersInPurePostgresMode(t *testing.T) {
	ms := &Microservice{}
	topics := []string{"when-product-barcode-created", "when-product-barcode-updated", "when-product-barcode-deleted"}
	for _, topic := range topics {
		if err := ms.Consume("unused:9092", topic, "test-group", 0, func(IContext) error { return nil }); err != nil {
			t.Fatal(err)
		}
	}
}
