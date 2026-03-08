package config

const (
	MQ_TOPIC_CREATED      string = "when-stockpickupproduct-created"
	MQ_TOPIC_UPDATED      string = "when-stockpickupproduct-updated"
	MQ_TOPIC_DELETED      string = "when-stockpickupproduct-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-stockpickupproduct-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-stockpickupproduct-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-stockpickupproduct-bulk-deleted"
)

type StockPickupProductMessageQueueConfig struct{}

func (StockPickupProductMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (StockPickupProductMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (StockPickupProductMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (StockPickupProductMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (StockPickupProductMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (StockPickupProductMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
