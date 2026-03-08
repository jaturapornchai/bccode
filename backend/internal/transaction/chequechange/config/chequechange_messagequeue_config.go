package config

const (
	MQ_TOPIC_CREATED      string = "when-chequechange-created"
	MQ_TOPIC_UPDATED      string = "when-chequechange-updated"
	MQ_TOPIC_DELETED      string = "when-chequechange-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-chequechange-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-chequechange-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-chequechange-bulk-deleted"
)

type ChequeChangeMessageQueueConfig struct{}

func (ChequeChangeMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (ChequeChangeMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (ChequeChangeMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (ChequeChangeMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (ChequeChangeMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (ChequeChangeMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
