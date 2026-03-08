package config

const (
	MQ_TOPIC_CREATED      string = "when-deposit-created"
	MQ_TOPIC_UPDATED      string = "when-deposit-updated"
	MQ_TOPIC_DELETED      string = "when-deposit-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-deposit-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-deposit-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-deposit-bulk-deleted"
)

type DepositMessageQueueConfig struct{}

func (DepositMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (DepositMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (DepositMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (DepositMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (DepositMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (DepositMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
