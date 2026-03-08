package config

const (
	MQ_TOPIC_CREATED      string = "when-stockreceiveproduct-created"
	MQ_TOPIC_UPDATED      string = "when-stockreceiveproduct-updated"
	MQ_TOPIC_DELETED      string = "when-stockreceiveproduct-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-stockreceiveproduct-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-stockreceiveproduct-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-stockreceiveproduct-bulk-deleted"
)

type StockReceiveProductMessageQueueConfig struct{}

func (StockReceiveProductMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (StockReceiveProductMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (StockReceiveProductMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (StockReceiveProductMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (StockReceiveProductMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (StockReceiveProductMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
