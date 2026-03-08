package config

const (
	MQ_TOPIC_CREATED      string = "when-chequepass-created"
	MQ_TOPIC_UPDATED      string = "when-chequepass-updated"
	MQ_TOPIC_DELETED      string = "when-chequepass-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-chequepass-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-chequepass-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-chequepass-bulk-deleted"
)

type ChequePassMessageQueueConfig struct{}

func (ChequePassMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (ChequePassMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (ChequePassMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (ChequePassMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (ChequePassMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (ChequePassMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
