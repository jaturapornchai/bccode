package config

const (
	MQ_TOPIC_CREATED      string = "when-depositrefund-created"
	MQ_TOPIC_UPDATED      string = "when-depositrefund-updated"
	MQ_TOPIC_DELETED      string = "when-depositrefund-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-depositrefund-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-depositrefund-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-depositrefund-bulk-deleted"
)

type DepositRefundMessageQueueConfig struct{}

func (DepositRefundMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (DepositRefundMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (DepositRefundMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (DepositRefundMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (DepositRefundMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (DepositRefundMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
