package config

const (
	MQ_TOPIC_CREATED      string = "when-pickandpack-created"
	MQ_TOPIC_UPDATED      string = "when-pickandpack-updated"
	MQ_TOPIC_DELETED      string = "when-pickandpack-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-pickandpack-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-pickandpack-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-pickandpack-bulk-deleted"
)

type PickandpackMessageQueueConfig struct{}

func (PickandpackMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (PickandpackMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (PickandpackMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (PickandpackMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (PickandpackMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (PickandpackMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
