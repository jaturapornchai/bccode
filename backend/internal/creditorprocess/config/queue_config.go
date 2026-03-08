package config

const (
	MQ_TOPIC_STOCK_PROCESS_CREATED      string = "when-creditor-process-created"
	MQ_TOPIC_STOCK_PROCESS_BULK_CREATED string = "when-creditor-process-bulk-created"
)

type CreditorProcessMessageQueueConfig struct {
}

func (CreditorProcessMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_STOCK_PROCESS_CREATED
}

func (CreditorProcessMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_STOCK_PROCESS_CREATED
}

func (CreditorProcessMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_STOCK_PROCESS_CREATED
}

func (CreditorProcessMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_STOCK_PROCESS_BULK_CREATED
}

func (CreditorProcessMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_STOCK_PROCESS_BULK_CREATED
}

func (CreditorProcessMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_STOCK_PROCESS_BULK_CREATED
}

func (CreditorProcessMessageQueueConfig) ConsumerGroup() string {
	return MQ_TOPIC_STOCK_PROCESS_CREATED
}
