package config

const (
	MQ_TOPIC_CREATED      string = "when-receivedepositrefund-created"
	MQ_TOPIC_UPDATED      string = "when-receivedepositrefund-updated"
	MQ_TOPIC_DELETED      string = "when-receivedepositrefund-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-receivedepositrefund-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-receivedepositrefund-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-receivedepositrefund-bulk-deleted"
)

type ReceiveDepositRefundMessageQueueConfig struct{}

func (ReceiveDepositRefundMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (ReceiveDepositRefundMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (ReceiveDepositRefundMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (ReceiveDepositRefundMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (ReceiveDepositRefundMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (ReceiveDepositRefundMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
