package mykafkaconsumer

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/segmentio/kafka-go"
)

type fakeProjectionReader struct {
	messages  []kafka.Message
	committed []kafka.Message
	fetched   int
	commitErr error
}

func (r *fakeProjectionReader) FetchMessage(context.Context) (kafka.Message, error) {
	r.fetched++
	if len(r.messages) == 0 {
		return kafka.Message{}, io.EOF
	}
	m := r.messages[0]
	r.messages = r.messages[1:]
	return m, nil
}
func (r *fakeProjectionReader) CommitMessages(_ context.Context, messages ...kafka.Message) error {
	if r.commitErr != nil {
		return r.commitErr
	}
	r.committed = append(r.committed, messages...)
	return nil
}

func TestProductProjectionAcknowledgesOnlySuccessfulWork(t *testing.T) {
	for _, mode := range []string{"success", "handler failure", "panic", "missing holding", "commit failure"} {
		t.Run(mode, func(t *testing.T) {
			payload := []byte("{\"holdingcode\":\"HOLDING\",\"businesscode\":\"COMPANY\"}")
			if mode == "missing holding" {
				payload = []byte("{}")
			}
			reader := &fakeProjectionReader{messages: []kafka.Message{{Offset: 7, Value: payload}, {Offset: 8, Value: payload}}}
			if mode == "commit failure" {
				reader.commitErr = errors.New("broker unavailable")
			}
			handled := 0
			err := consumeProductProjection(context.Background(), reader, func(string) error {
				if len(reader.committed) != handled {
					t.Error("committed before handler success")
				}
				handled++
				if mode == "handler failure" {
					return errors.New("database unavailable")
				}
				if mode == "panic" {
					panic("private payload")
				}
				return nil
			})
			if mode == "success" {
				if !errors.Is(err, io.EOF) || len(reader.committed) != 2 || reader.committed[0].Offset != 7 || reader.committed[1].Offset != 8 {
					t.Fatal("successful offsets not acknowledged in order")
				}
			} else if err == nil || len(reader.committed) != 0 || reader.fetched != 1 {
				t.Fatalf("failure skipped/committed offset: err=%v committed=%d fetched=%d", err, len(reader.committed), reader.fetched)
			}
		})
	}
}
func TestProductProjectionTopicScope(t *testing.T) {
	for _, topic := range []string{"when-product-created", "when-product-updated", "when-product-deleted", "when-product-barcode-bulk-updated"} {
		if !isProductProjectionTopic(topic) || projectionCommitInterval(topic) != 0 {
			t.Fatal("projection must commit synchronously")
		}
	}
	if isProductProjectionTopic("when-saleinvoice-created") || projectionCommitInterval("when-saleinvoice-created") == 0 {
		t.Fatal("unrelated consumer contract changed")
	}
}
