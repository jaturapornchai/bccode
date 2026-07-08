package config

const (
	MQ_TOPIC_CREATED      string = "when-product-created"
	MQ_TOPIC_UPDATED      string = "when-product-updated"
	MQ_TOPIC_DELETED      string = "when-product-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-product-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-product-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-product-bulk-deleted"
)

type ProductMessageQueueConfig struct{}

func (ProductMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (ProductMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (ProductMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (ProductMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (ProductMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (ProductMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
