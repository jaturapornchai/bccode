package config

const (
	MQ_TOPIC_CREATED      string = "when-debtor-created"
	MQ_TOPIC_UPDATED      string = "when-debtor-updated"
	MQ_TOPIC_DELETED      string = "when-debtor-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-debtor-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-debtor-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-debtor-bulk-deleted"
)

type DebtorMessageQueueConfig struct{}

func (DebtorMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (DebtorMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (DebtorMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (DebtorMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (DebtorMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (DebtorMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
