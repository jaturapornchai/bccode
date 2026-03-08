package config

const (
	MQ_TOPIC_CREATED      string = "when-stockbalancedetail-created"
	MQ_TOPIC_UPDATED      string = "when-stockbalancedetail-updated"
	MQ_TOPIC_DELETED      string = "when-stockbalancedetail-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-stockbalancedetail-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-stockbalancedetail-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-stockbalancedetail-bulk-deleted"
)

type StockBalanceDetailMessageQueueConfig struct{}

func (StockBalanceDetailMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (StockBalanceDetailMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (StockBalanceDetailMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (StockBalanceDetailMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (StockBalanceDetailMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (StockBalanceDetailMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
