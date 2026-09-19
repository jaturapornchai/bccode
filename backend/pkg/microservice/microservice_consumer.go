package microservice

import (
	"time"
)

type IMicroserviceConsumer interface {
	RegisterConsumer(*Microservice)
}

// Consume registers consumer endpoint (No-op in pure PostgreSQL mode)
func (ms *Microservice) Consume(servers string, topic string, groupID string, readTimeout time.Duration, h ServiceHandleFunc) error {
	if ms != nil && ms.Logger != nil {
		ms.Logger.Debugf("[SQL-MODE] Consumer registered as no-op for topic: %s", topic)
	}
	return nil
}

// ConsumeFromBegining registers consumer endpoint (No-op in pure PostgreSQL mode)
func (ms *Microservice) ConsumeFromBegining(servers string, topic string, readTimeout time.Duration, h ServiceHandleFunc) error {
	if ms != nil && ms.Logger != nil {
		ms.Logger.Debugf("[SQL-MODE] Consumer registered as no-op for topic: %s", topic)
	}
	return nil
}

func (ms *Microservice) RegisterConsumer(consumer IMicroserviceConsumer) {
	defer ms.consumerRecover()
	consumer.RegisterConsumer(ms)
}

func (ms *Microservice) consumerRecover() {
	if r := recover(); r != nil {
		ms.Logger.Errorf("Recovered from panic: %v", r)
	}
}
