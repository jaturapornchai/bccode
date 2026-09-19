package microservice

import (
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/logger"
	"time"
)

// IMQ is interface to manage message topics (No-op in pure PostgreSQL mode)
type IMQ interface {
	CreateTopic(topic string, partitions int, replications int) error
	CreateTopicR(topic string, partitions int, replications int, retentionPeriod time.Duration) error
}

// MQ is message queue placeholder for pure PostgreSQL mode
type MQ struct {
	logger  logger.ILogger
	servers string
}

// NewMQ return new MQ
func NewMQ(mqConfig config.IMQConfig, logger logger.ILogger) *MQ {
	return &MQ{
		servers: mqConfig.URI(),
		logger:  logger,
	}
}

// CreateTopicR create topic with retention period (No-op)
func (q *MQ) CreateTopicR(topic string, partitions int, replications int, retentionPeriod time.Duration) error {
	return nil
}

// CreateTopic create the topic (No-op)
func (q *MQ) CreateTopic(topic string, partitions int, replications int) error {
	return nil
}
