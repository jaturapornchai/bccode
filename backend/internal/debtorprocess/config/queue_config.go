package config

const (
	MQ_TOPIC_STOCK_PROCESS_CREATED      string = "when-debtor-process-created"
	MQ_TOPIC_STOCK_PROCESS_BULK_CREATED string = "when-debtor-process-bulk-created"
)

type DebtorProcessMessageQueueConfig struct {
}

func (DebtorProcessMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_STOCK_PROCESS_CREATED
}

func (DebtorProcessMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_STOCK_PROCESS_CREATED
}

func (DebtorProcessMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_STOCK_PROCESS_CREATED
}

func (DebtorProcessMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_STOCK_PROCESS_BULK_CREATED
}

func (DebtorProcessMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_STOCK_PROCESS_BULK_CREATED
}

func (DebtorProcessMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_STOCK_PROCESS_BULK_CREATED
}

func (DebtorProcessMessageQueueConfig) ConsumerGroup() string {
	return MQ_TOPIC_STOCK_PROCESS_CREATED
}
