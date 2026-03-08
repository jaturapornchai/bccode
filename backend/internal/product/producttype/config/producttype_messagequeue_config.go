package config

const (
	MQ_TOPIC_CREATED      string = "when-product-type-created"
	MQ_TOPIC_UPDATED      string = "when-product-type-updated"
	MQ_TOPIC_DELETED      string = "when-product-type-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-product-type-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-product-type-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-product-type-bulk-deleted"
)

type ProductTypeMessageQueueConfig struct{}

func (ProductTypeMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (ProductTypeMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (ProductTypeMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (ProductTypeMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (ProductTypeMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (ProductTypeMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
