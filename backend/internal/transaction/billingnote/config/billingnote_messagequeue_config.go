package config

const (
	MQ_TOPIC_CREATED      string = "when-billingnote-created"
	MQ_TOPIC_UPDATED      string = "when-billingnote-updated"
	MQ_TOPIC_DELETED      string = "when-billingnote-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-billingnote-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-billingnote-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-billingnote-bulk-deleted"
)

type MessageQueueConfig struct{}

func (MessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (MessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (MessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (MessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (MessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (MessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
