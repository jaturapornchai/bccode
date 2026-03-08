package config

const (
	MQ_TOPIC_CREATED      string = "when-product-group-created"
	MQ_TOPIC_UPDATED      string = "when-product-group-updated"
	MQ_TOPIC_DELETED      string = "when-product-group-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-product-group-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-product-group-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-product-group-bulk-deleted"
)

type ProductGroupMessageQueueConfig struct{}

func (ProductGroupMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (ProductGroupMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (ProductGroupMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (ProductGroupMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (ProductGroupMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (ProductGroupMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
