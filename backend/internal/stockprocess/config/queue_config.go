package stockprocess

const (
	MQ_TOPIC_STOCK_PROCESS_CREATED      string = "when-stock-process-created"
	MQ_TOPIC_STOCK_PROCESS_BULK_CREATED string = "when-stock-process-bulk-created"
)

type StockProcessMessageQueueConfig struct{}

func (StockProcessMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_STOCK_PROCESS_CREATED
}

func (StockProcessMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_STOCK_PROCESS_CREATED
}

func (StockProcessMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_STOCK_PROCESS_CREATED
}

func (StockProcessMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_STOCK_PROCESS_BULK_CREATED
}

func (StockProcessMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_STOCK_PROCESS_BULK_CREATED
}

func (StockProcessMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_STOCK_PROCESS_BULK_CREATED
}

func (StockProcessMessageQueueConfig) ConsumerGroup() string {
	return MQ_TOPIC_STOCK_PROCESS_CREATED
}
