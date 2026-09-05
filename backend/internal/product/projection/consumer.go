package projection

import (
	"context"
	"errors"
	"github.com/segmentio/kafka-go"
	"time"
)

// ErrRejected marks a message that can never be applied: unparsable JSON or a
// missing tenant identity. The handler logs the position and reason; Consume
// acknowledges the offset so one unusable message cannot block every later
// message on the partition. Infrastructure or source failures must not use it.
var ErrRejected = errors.New("projection message rejected")

type Reader interface {
	FetchMessage(context.Context) (kafka.Message, error)
	CommitMessages(context.Context, ...kafka.Message) error
}

// Consume never fetches a higher offset after a handler or commit failure.
// Reopening the reader with the same group replays the unacknowledged message.
func Consume(ctx context.Context, reader Reader, handle func(kafka.Message) error) error {
	for {
		fetchCtx, cancel := context.WithTimeout(ctx, time.Second)
		message, err := reader.FetchMessage(fetchCtx)
		cancel()
		if errors.Is(err, context.DeadlineExceeded) && ctx.Err() == nil {
			continue
		}
		if err != nil {
			return err
		}
		if err := handle(message); err != nil && !errors.Is(err, ErrRejected) {
			return err
		}
		commitCtx, stop := context.WithTimeout(ctx, 30*time.Second)
		err = reader.CommitMessages(commitCtx, message)
		stop()
		if err != nil {
			return err
		}
	}
}
func IsBarcodeTopic(topic string) bool {
	switch topic {
	case "when-product-barcode-created", "when-product-barcode-updated", "when-product-barcode-deleted",
		"when-product-barcode-bulk-created", "when-product-barcode-bulk-updated", "when-product-barcode-bulk-deleted":
		return true
	default:
		return false
	}
}
