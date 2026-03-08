package config

const (
	MQ_TOPIC_CREATED      string = "when-stockbalance-created"
	MQ_TOPIC_UPDATED      string = "when-stockbalance-updated"
	MQ_TOPIC_DELETED      string = "when-stockbalance-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-stockbalance-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-stockbalance-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-stockbalance-bulk-deleted"
)

type StockBalanceMessageQueueConfig struct{}

func (StockBalanceMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (StockBalanceMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (StockBalanceMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (StockBalanceMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (StockBalanceMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (StockBalanceMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
