//go:build integration

package projection

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type rebalanceCommitReader struct {
	*kafka.Reader
	committed chan<- kafka.Message
}

func (r rebalanceCommitReader) CommitMessages(ctx context.Context, messages ...kafka.Message) error {
	if err := r.Reader.CommitMessages(ctx, messages...); err != nil {
		return err
	}
	for _, message := range messages {
		select {
		case r.committed <- message:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func TestProjectionRebalanceIntegration(t *testing.T) {
	broker := os.Getenv("BC_PROJECTION_TEST_KAFKA")
	if broker == "" {
		t.Skip("requires isolated Kafka")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	topic := "bc-rebalance-" + primitive.NewObjectID().Hex()
	conn, err := kafka.DialContext(ctx, "tcp", broker)
	if err != nil {
		t.Fatal(err)
	}
	controller, err := conn.Controller()
	conn.Close()
	if err != nil {
		t.Fatal(err)
	}
	control, err := kafka.DialContext(ctx, "tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		t.Fatal(err)
	}
	if err := control.CreateTopics(kafka.TopicConfig{Topic: topic, NumPartitions: 2, ReplicationFactor: 1}); err != nil {
		control.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		control.SetDeadline(time.Now().Add(10 * time.Second))
		if err := control.DeleteTopics(topic); err != nil {
			t.Error(err)
		}
		control.Close()
	})
	for partition := 0; partition < 2; partition++ {
		leader, err := kafka.DialLeader(ctx, "tcp", broker, topic, partition)
		if err != nil {
			t.Fatal(err)
		}
		leader.SetWriteDeadline(time.Now().Add(10 * time.Second))
		messages := []kafka.Message{}
		for i := 0; i < 6; i++ {
			messages = append(messages, kafka.Message{Value: []byte(fmt.Sprintf("%d:%d", partition, i))})
		}
		_, err = leader.WriteMessages(messages...)
		leader.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
	newReader := func() *kafka.Reader {
		return kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{broker}, Topic: topic, GroupID: topic + "-group",
			MinBytes: 1, MaxBytes: 1e6, CommitInterval: 0, StartOffset: kafka.FirstOffset,
			MaxWait: time.Second, SessionTimeout: 6 * time.Second, HeartbeatInterval: time.Second,
		})
	}
	readerA := newReader()
	t.Cleanup(func() { readerA.Close() })
	failure := errors.New("write rejected")
	var uncommitted kafka.Message
	if err := Consume(ctx, readerA, func(message kafka.Message) error { uncommitted = message; return failure }); !errors.Is(err, failure) {
		t.Fatalf("failed handler advanced: %v", err)
	}
	// Reader A remains a live group member but its handler did not commit.
	// Reader B joins the same group, requiring a rebalance. Closing A after B
	// commits forces another assignment change and replay of A's failed offset.
	progress := make(chan kafka.Message, 24)
	bCtx, stopB := context.WithCancel(ctx)
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		for bCtx.Err() == nil {
			r := newReader()
			_ = Consume(bCtx, rebalanceCommitReader{r, progress}, func(kafka.Message) error { return nil })
			r.Close()
		}
	}()
	defer func() { stopB(); <-finished }()
	seen := map[string]bool{}
	aClosed := false
	for len(seen) < 12 {
		select {
		case message := <-progress:
			seen[fmt.Sprintf("%d:%d", message.Partition, message.Offset)] = true
			if !aClosed {
				if err := readerA.Close(); err != nil {
					t.Fatal(err)
				}
				aClosed = true
			}
		case <-ctx.Done():
			t.Fatalf("partition reassignment lost messages: committed=%d: %v", len(seen), ctx.Err())
		}
	}
	if !seen[fmt.Sprintf("%d:%d", uncommitted.Partition, uncommitted.Offset)] {
		t.Fatal("failed offset was skipped after reassignment")
	}
	t.Log("two live members, two partitions, failed handler, member departure and committed replay passed")
}
