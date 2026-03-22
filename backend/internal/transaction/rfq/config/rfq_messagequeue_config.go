package config

const (
	MQ_TOPIC_CREATED      string = "when-rfq-created"
	MQ_TOPIC_UPDATED      string = "when-rfq-updated"
	MQ_TOPIC_DELETED      string = "when-rfq-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-rfq-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-rfq-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-rfq-bulk-deleted"
)

type RFQMessageQueueConfig struct{}

func (RFQMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (RFQMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (RFQMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (RFQMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (RFQMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (RFQMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
