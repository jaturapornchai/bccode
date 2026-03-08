package config

const (
	MQ_TOPIC_CREATED      string = "when-stocktransfer-created"
	MQ_TOPIC_UPDATED      string = "when-stocktransfer-updated"
	MQ_TOPIC_DELETED      string = "when-stocktransfer-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-stocktransfer-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-stocktransfer-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-stocktransfer-bulk-deleted"
)

type StockTransferMessageQueueConfig struct{}

func (StockTransferMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (StockTransferMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (StockTransferMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (StockTransferMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (StockTransferMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (StockTransferMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
