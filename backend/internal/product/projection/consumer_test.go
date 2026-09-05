package projection

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/segmentio/kafka-go"
)

type fakeReader struct {
	messages  []kafka.Message
	committed []kafka.Message
}

func (r *fakeReader) FetchMessage(context.Context) (kafka.Message, error) {
	if len(r.messages) == 0 {
		return kafka.Message{}, io.EOF
	}
	m := r.messages[0]
	r.messages = r.messages[1:]
	return m, nil
}

func (r *fakeReader) CommitMessages(_ context.Context, messages ...kafka.Message) error {
	r.committed = append(r.committed, messages...)
	return nil
}

// A rejected message is acknowledged so the partition keeps moving; any other
// handler failure still holds the offset and stops the loop.
func TestConsumeAcknowledgesRejectedMessagesOnly(t *testing.T) {
	reader := &fakeReader{messages: []kafka.Message{{Offset: 1}, {Offset: 2}, {Offset: 3}, {Offset: 4}}}
	err := Consume(context.Background(), reader, func(m kafka.Message) error {
		switch m.Offset {
		case 2:
			return fmt.Errorf("%w: holdingcode is required", ErrRejected)
		case 4:
			return errors.New("postgresql unavailable")
		}
		return nil
	})
	if err == nil || errors.Is(err, ErrRejected) {
		t.Fatalf("infrastructure failure must surface: %v", err)
	}
	if len(reader.committed) != 3 || reader.committed[0].Offset != 1 || reader.committed[1].Offset != 2 || reader.committed[2].Offset != 3 {
		t.Fatalf("expected offsets 1,2,3 acknowledged, got %v", reader.committed)
	}
	if len(reader.messages) != 0 {
		t.Fatal("failed offset 4 must be the last message fetched")
	}
}
