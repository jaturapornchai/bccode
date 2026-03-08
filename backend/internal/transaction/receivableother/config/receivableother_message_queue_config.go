package config

const (
	MQ_TOPIC_CREATED      string = "when-debtor-receivableother-created"
	MQ_TOPIC_UPDATED      string = "when-debtor-receivableother-updated"
	MQ_TOPIC_DELETED      string = "when-debtor-receivableother-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-debtor-receivableother-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-debtor-receivableother-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-debtor-receivableother-bulk-deleted"
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
