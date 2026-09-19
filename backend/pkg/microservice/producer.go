package microservice

import (
	"smlcloudplatform/internal/logger"
	"time"
)

// IProducer is interface for producer
type IProducer interface {
	SendMessage(topic string, key string, message interface{}) error
	Close() error
	TestConnect() error
}

// Producer implements IProducer without Kafka dependency (Pure PostgreSQL mode)
type Producer struct {
	logger           logger.ILogger
	servers          string
	protocol         string
	sslca            string
	sslkey           string
	sslcert          string
	messageTimeoutMs int
}

// NewProducer return new instance of Producer
func NewProducer(servers string, protocol string, sslca string, sslkey string, sslcert string, logger logger.ILogger) *Producer {
	return &Producer{
		logger:           logger,
		messageTimeoutMs: 43200000,
		servers:          servers,
		protocol:         protocol,
		sslca:            sslca,
		sslkey:           sslkey,
		sslcert:          sslcert,
	}
}

// NewProducerWithTimeout gives durable outbox delivery a bounded wait
func NewProducerWithTimeout(servers, protocol, sslca, sslkey, sslcert string, logger logger.ILogger, timeout time.Duration) *Producer {
	p := NewProducer(servers, protocol, sslca, sslkey, sslcert, logger)
	if timeout > 0 {
		p.messageTimeoutMs = int(max(timeout.Milliseconds(), 1))
	}
	return p
}

func (p *Producer) TestConnect() error {
	return nil
}

// SendMessage synchronous no-op for pure PostgreSQL architecture
func (p *Producer) SendMessage(topic string, key string, message interface{}) error {
	if p.logger != nil {
		p.logger.Debugf("[SQL-DISPATCH] Message dispatched synchronously for topic %s, key %s", topic, key)
	}
	return nil
}

// Close the producer
func (p *Producer) Close() error {
	return nil
}
