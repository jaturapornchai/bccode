package config

const (
	MQ_TOPIC_CREATED      string = "when-chequedeposit-created"
	MQ_TOPIC_UPDATED      string = "when-chequedeposit-updated"
	MQ_TOPIC_DELETED      string = "when-chequedeposit-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-chequedeposit-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-chequedeposit-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-chequedeposit-bulk-deleted"
)

type ChequeDepositMessageQueueConfig struct{}

func (ChequeDepositMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (ChequeDepositMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (ChequeDepositMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (ChequeDepositMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (ChequeDepositMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (ChequeDepositMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
