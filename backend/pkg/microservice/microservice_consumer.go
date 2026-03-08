package microservice

import (
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type IMicroserviceConsumer interface {
	RegisterConsumer(*Microservice)
}

func (ms *Microservice) consumeSingle(servers string, topic string, groupID string, readTimeout time.Duration, h ServiceHandleFunc) {
	ms.Logger.Debugf("Consumer Kafka on topic: %s ", topic)
	c, err := ms.newKafkaConsumer(servers, groupID)
	if err != nil {
		return
	}

	defer c.Close()

	c.Subscribe(topic, nil)

	for {
		if readTimeout <= 0 {
			// readtimeout -1 indicates no timeout
			readTimeout = -1
		}

		msg, err := c.ReadMessage(readTimeout)
		if err != nil {
			kafkaErr, ok := err.(kafka.Error)
			if ok {
				if kafkaErr.Code() == kafka.ErrTimedOut {
					if readTimeout == -1 {
						// No timeout just continue to read message again
						continue
					}
				}
			}
			ms.Log("Consumer", err.Error())
			ms.Stop()
			return
		}

		// Execute Handler
		h(NewConsumerContext(ms, string(msg.Value)))
	}
}

// Consume register service endpoint for Consumer service
func (ms *Microservice) Consume(servers string, topic string, groupID string, readTimeout time.Duration, h ServiceHandleFunc) error {
	go ms.consumeSingle(servers, topic, groupID, readTimeout, h)
	return nil
}

func (ms *Microservice) consumeWithoutGroupFromBeginig(servers string, topic string, readTimeout time.Duration, h ServiceHandleFunc) {
	ms.Logger.Debugf("Consumer Kafka on topic: %s ", topic)
	c, err := ms.newKafkaComsuperStartFromBeginning(servers)
	if err != nil {
		return
	}

	defer c.Close()

	c.Subscribe(topic, nil)

	for {
		if readTimeout <= 0 {
			// readtimeout -1 indicates no timeout
			readTimeout = -1
		}

		msg, err := c.ReadMessage(readTimeout)
		if err != nil {
			kafkaErr, ok := err.(kafka.Error)
			if ok {
				if kafkaErr.Code() == kafka.ErrTimedOut {
					if readTimeout == -1 {
						// No timeout just continue to read message again
						continue
					}
				}
			}
			ms.Log("Consumer", err.Error())
			ms.Stop()
			return
		}

		// Execute Handler
		h(NewConsumerContext(ms, string(msg.Value)))
	}
}

func (ms *Microservice) ConsumeFromBegining(servers string, topic string, readTimeout time.Duration, h ServiceHandleFunc) error {
	go ms.consumeWithoutGroupFromBeginig(servers, topic, readTimeout, h)
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
