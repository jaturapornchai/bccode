package config

const (
	MQ_TOPIC_CREATED      string = "when-stockreturnproduct-created"
	MQ_TOPIC_UPDATED      string = "when-stockreturnproduct-updated"
	MQ_TOPIC_DELETED      string = "when-stockreturnproduct-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-stockreturnproduct-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-stockreturnproduct-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-stockreturnproduct-bulk-deleted"
)

type StockReturnProductMessageQueueConfig struct{}

func (StockReturnProductMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (StockReturnProductMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (StockReturnProductMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (StockReturnProductMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (StockReturnProductMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (StockReturnProductMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
