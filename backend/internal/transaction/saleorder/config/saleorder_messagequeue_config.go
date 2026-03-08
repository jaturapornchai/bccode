package config

const (
	MQ_TOPIC_CREATED      string = "when-saleorder-created"
	MQ_TOPIC_UPDATED      string = "when-saleorder-updated"
	MQ_TOPIC_DELETED      string = "when-saleorder-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-saleorder-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-saleorder-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-saleorder-bulk-deleted"
)

type SaleOrderMessageQueueConfig struct{}

func (SaleOrderMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (SaleOrderMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (SaleOrderMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (SaleOrderMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (SaleOrderMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (SaleOrderMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
