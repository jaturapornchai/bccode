package config

const (
	MQ_TOPIC_CREATED      string = "when-receivedeposit-created"
	MQ_TOPIC_UPDATED      string = "when-receivedeposit-updated"
	MQ_TOPIC_DELETED      string = "when-receivedeposit-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-receivedeposit-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-receivedeposit-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-receivedeposit-bulk-deleted"
)

type ReceiveDepositMessageQueueConfig struct{}

func (ReceiveDepositMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (ReceiveDepositMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (ReceiveDepositMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (ReceiveDepositMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (ReceiveDepositMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (ReceiveDepositMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
