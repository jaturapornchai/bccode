package mykafkaconsumer

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/product/projection"
	"time"

	"github.com/segmentio/kafka-go"
)

func isProductProjectionTopic(topic string) bool {
	switch topic {
	case "when-product-created", "when-product-updated", "when-product-deleted",
		"when-product-barcode-created", "when-product-barcode-updated", "when-product-barcode-deleted",
		"when-product-barcode-bulk-created", "when-product-barcode-bulk-updated", "when-product-barcode-bulk-deleted":
		return true
	default:
		return false
	}
}

func projectionCommitInterval(topic string) time.Duration {
	if isProductProjectionTopic(topic) {
		return 0
	} // Broker acknowledgement, not a queued async commit.
	return 5 * time.Second
}

type projectionReader = projection.Reader

// A failed message must close/rejoin the reader without committing or fetching
// a higher offset. Detached shop jobs cannot provide that acknowledgement.
func consumeProductProjection(ctx context.Context, reader projectionReader, handler ConsumerHandleFunc) error {
	return projection.Consume(ctx, reader, func(message kafka.Message) error { return handleProductProjection(message, handler) })
}

func handleProductProjection(message kafka.Message, handler ConsumerHandleFunc) (err error) {
	defer func() {
		if recover() != nil {
			err = fmt.Errorf("product projection handler panicked at offset %d", message.Offset)
		}
	}()
	if extractHoldingCodeFromMessage(string(message.Value)) == "" {
		return fmt.Errorf("product projection has no holding identity at offset %d", message.Offset)
	}
	if handler(string(message.Value)) != nil {
		// The handler already owns diagnostic logging. Do not echo payload or raw
		// connection errors into the retry loop.
		return fmt.Errorf("product projection handler failed at offset %d", message.Offset)
	}
	return nil
}

// ConsumeProjectionMessages processes a reader using acknowledged handler completion.
// The reader interface also permits commit observation in real-broker tests.
func ConsumeProjectionMessages(ctx context.Context, reader projectionReader, handler ConsumerHandleFunc) error {
	return consumeProductProjection(ctx, reader, handler)
}
