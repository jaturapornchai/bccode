package config

const (
	MQ_TOPIC_CREATED      string = "when-product-unit-created"
	MQ_TOPIC_UPDATED      string = "when-product-unit-updated"
	MQ_TOPIC_DELETED      string = "when-product-unit-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-product-unit-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-product-unit-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-product-unit-bulk-deleted"
)

type UnitMessageQueueConfig struct{}

func (UnitMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (UnitMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (UnitMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (UnitMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (UnitMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (UnitMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
