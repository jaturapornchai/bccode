package config

const (
	MQ_TOPIC_CREATED      string = "when-stockadjustment-created"
	MQ_TOPIC_UPDATED      string = "when-stockadjustment-updated"
	MQ_TOPIC_DELETED      string = "when-stockadjustment-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-stockadjustment-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-stockadjustment-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-stockadjustment-bulk-deleted"
)

type StockAdjustmentMessageQueueConfig struct{}

func (StockAdjustmentMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (StockAdjustmentMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (StockAdjustmentMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (StockAdjustmentMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (StockAdjustmentMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (StockAdjustmentMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
