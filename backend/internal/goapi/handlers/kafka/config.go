package kafka

import (
	"os"
)

// KafkaConfig holds configuration for Kafka consumers
type KafkaConfig struct {
	ServerURL      string
	BatchSize      int
	ConsumerGroups map[string]string
}

// NewKafkaConfig creates a new Kafka configuration
func NewKafkaConfig() *KafkaConfig {
	return &KafkaConfig{
		ServerURL: os.Getenv("KAFKA_SERVER_URL"),
		BatchSize: BATCH_SIZE,
		ConsumerGroups: map[string]string{
			"sale_invoice":   CONSUMER_GROUP_SALE_INVOICE,
			"sale_return":    CONSUMER_GROUP_SALE_RETURN,
			"inventory":      CONSUMER_GROUP_INVENTORY,
			"inventory_bulk": CONSUMER_GROUP_INVENTORY_BULK,
			"warehouse":      CONSUMER_GROUP_WAREHOUSE,
		},
	}
}

// IsValid checks if the configuration is valid
func (c *KafkaConfig) IsValid() bool {
	return c.ServerURL != ""
}
