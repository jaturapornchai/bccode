package config

const (
	MQ_TOPIC_CREATED      string = "when-creditor-created"
	MQ_TOPIC_UPDATED      string = "when-creditor-updated"
	MQ_TOPIC_DELETED      string = "when-creditor-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-creditor-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-creditor-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-creditor-bulk-deleted"
)

type CreditorMessageQueueConfig struct {
}

func (CreditorMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (CreditorMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (CreditorMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (CreditorMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (CreditorMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (CreditorMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
