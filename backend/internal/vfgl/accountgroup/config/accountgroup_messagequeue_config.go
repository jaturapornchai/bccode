package config

const (
	MQ_TOPIC_CREATED      string = "when-accountgroup-created"
	MQ_TOPIC_UPDATED      string = "when-accountgroup-updated"
	MQ_TOPIC_DELETED      string = "when-accountgroup-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-accountgroup-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-accountgroup-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-accountgroup-bulk-deleted"
)

type AccountGroupMessageQueueConfig struct{}

func (AccountGroupMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (AccountGroupMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (AccountGroupMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (AccountGroupMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (AccountGroupMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (AccountGroupMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
