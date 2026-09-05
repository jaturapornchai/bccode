package microservice

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	kafkago "github.com/segmentio/kafka-go"
	"smlcloudplatform/internal/product/projection"
)

// kafka-go and the librdkafka members that still serve the other legacy topics
// encode JoinGroup member metadata differently, so a shared group would either
// rebalance endlessly or leave these readers without partitions. Barcode readers
// therefore use their own group; reconciliation reads current source, so the
// first start from the earliest offset is idempotent.
func barcodeProjectionGroup(group string) string { return group + "-projection" }

// Barcode metadata handlers must complete before committing an offset. Other
// legacy topics retain their existing transport until their contracts are reviewed.
func (ms *Microservice) consumeBarcodeProjection(ctx context.Context, servers, topic, group string, h ServiceHandleFunc) {
	for ctx.Err() == nil {
		reader := kafkago.NewReader(kafkago.ReaderConfig{
			Brokers: strings.Split(servers, ","), Topic: topic, GroupID: barcodeProjectionGroup(group),
			StartOffset: kafkago.FirstOffset, CommitInterval: 0,
			MinBytes: 1, MaxBytes: 10000000, MaxWait: time.Second,
		})
		err := projection.Consume(ctx, reader, func(message kafkago.Message) (err error) {
			defer func() {
				if recover() != nil {
					err = fmt.Errorf("barcode handler panicked at offset %d", message.Offset)
				}
			}()
			err = h(NewConsumerContext(ms, string(message.Value)))
			if errors.Is(err, projection.ErrRejected) {
				ms.Logger.Warnf("Barcode projection skipped unusable message topic=%s partition=%d offset=%d: %v", topic, message.Partition, message.Offset, err)
			}
			return err
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
