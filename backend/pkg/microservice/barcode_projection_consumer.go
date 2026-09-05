package microservice

import (
	"context"
	"fmt"
	"strings"
	"time"

	kafkago "github.com/segmentio/kafka-go"
	"smlcloudplatform/internal/product/projection"
)

// Barcode metadata handlers must complete before committing an offset. Other
// legacy topics retain their existing transport until their contracts are reviewed.
func (ms *Microservice) consumeBarcodeProjection(ctx context.Context, servers, topic, group string, h ServiceHandleFunc) {
	for ctx.Err() == nil {
		reader := kafkago.NewReader(kafkago.ReaderConfig{
			Brokers: strings.Split(servers, ","), Topic: topic, GroupID: group,
			StartOffset: kafkago.FirstOffset, CommitInterval: 0,
			MinBytes: 1, MaxBytes: 10000000, MaxWait: time.Second,
		})
		err := projection.Consume(ctx, reader, func(message kafkago.Message) (err error) {
			defer func() {
				if recover() != nil {
					err = fmt.Errorf("barcode handler panicked at offset %d", message.Offset)
				}
			}()
			return h(NewConsumerContext(ms, string(message.Value)))
		})
		reader.Close()
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			ms.Logger.Warnf("Barcode projection consumer requires retry for topic %s", topic)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
		}
	}
}
