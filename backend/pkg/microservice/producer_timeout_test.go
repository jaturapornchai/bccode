package microservice

import (
	"testing"
	"time"
)

func TestProducerPurePostgresMode(t *testing.T) {
	p := NewProducerWithTimeout("127.0.0.1:1", "plaintext", "", "", "", nil, time.Second)
	if err := p.TestConnect(); err != nil {
		t.Fatalf("test connect: %v", err)
	}
	defer p.Close()

	err := p.SendMessage("bc-outbox-timeout-test", "test-key", map[string]string{"amount": "0.30"})
	if err != nil {
		t.Fatalf("expected nil error in pure postgres mode, got %v", err)
	}
}
