package repositories

import (
	"testing"

	"smlcloudplatform/internal/product/product/models"
)

type capturingProductProducer struct {
	message any
}

func (producer *capturingProductProducer) SendMessage(_ string, _ string, message interface{}) error {
	producer.message = message
	return nil
}

func (*capturingProductProducer) Close() error       { return nil }
func (*capturingProductProducer) TestConnect() error { return nil }

func TestProductMessageQueueExcludesReadOnlyBarcodes(t *testing.T) {
	tests := []struct {
		name string
		send func(ProductMessageQueueRepository, models.ProductDoc) error
	}{
		{name: "create", send: func(repo ProductMessageQueueRepository, doc models.ProductDoc) error { return repo.Create(doc) }},
		{name: "update", send: func(repo ProductMessageQueueRepository, doc models.ProductDoc) error { return repo.Update(doc) }},
		{name: "delete", send: func(repo ProductMessageQueueRepository, doc models.ProductDoc) error { return repo.Delete(doc) }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			producer := &capturingProductProducer{}
			repo := NewProductMessageQueueRepository(producer)
			doc := models.ProductDoc{}
			doc.Barcodes = []models.Barcodes{{Barcode: "8850000000001"}}

			if err := test.send(repo, doc); err != nil {
				t.Fatalf("send failed: %v", err)
			}
			message, ok := producer.message.(models.ProductDoc)
			if !ok {
				t.Fatalf("unexpected message type %T", producer.message)
			}
			if len(message.Barcodes) != 0 {
				t.Fatal("read-only barcodes leaked into product event")
			}
			if len(doc.Barcodes) != 1 {
				t.Fatal("event sanitization mutated the response document")
			}
		})
	}
}
