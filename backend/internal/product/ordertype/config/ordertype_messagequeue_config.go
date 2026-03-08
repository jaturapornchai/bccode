package config

const (
	MQ_TOPIC_CREATED      string = "when-product-order-type-created"
	MQ_TOPIC_UPDATED      string = "when-product-order-type-updated"
	MQ_TOPIC_DELETED      string = "when-product-order-type-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-product-order-type-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-product-order-type-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-product-order-type-bulk-deleted"
)

type OrderTypeMessageQueueConfig struct{}

func (OrderTypeMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (OrderTypeMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (OrderTypeMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (OrderTypeMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (OrderTypeMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (OrderTypeMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
