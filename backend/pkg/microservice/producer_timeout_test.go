package microservice

import (
	"smlcloudplatform/internal/logger"
	"strings"
	"testing"
	"time"
)

type silentProducerLogger struct{ logger.ILogger }

func (silentProducerLogger) Error(...interface{}) {}
func (silentProducerLogger) Debug(...interface{}) {}

func TestProducerDeliveryTimeout(t *testing.T) {
	defaultProducer := NewProducer("127.0.0.1:1", "plaintext", "", "", "", silentProducerLogger{})
	if defaultProducer.messageTimeoutMs != 43200000 {
		t.Fatal("existing callers must retain their configured timeout")
	}
	p := NewProducerWithTimeout("127.0.0.1:1", "plaintext", "", "", "", silentProducerLogger{}, time.Second)
	// Construct the real librdkafka producer, so an invalid timeout configuration
	// cannot masquerade as a successful bounded-delivery test.
	if err := p.TestConnect(); err != nil {
		t.Fatalf("initialize producer: %v", err)
	}
	defer p.Close()

	start := time.Now()
	err := p.SendMessage("bc-outbox-timeout-test", "test-key", map[string]string{"amount": "0.30"})
	elapsed := time.Since(start)
	if err == nil || !strings.Contains(err.Error(), "delivery failed") {
		t.Fatalf("expected broker delivery failure, got %v", err)
	}
	if elapsed > 5*time.Second {
		t.Fatalf("unavailable broker blocked delivery for %s", elapsed)
	}
	t.Logf("unavailable broker returned delivery failure in %s", elapsed)
}
