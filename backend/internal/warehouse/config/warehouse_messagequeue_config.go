package config

const (
	MQ_TOPIC_CREATED      string = "when-warehouse-created"
	MQ_TOPIC_UPDATED      string = "when-warehouse-updated"
	MQ_TOPIC_DELETED      string = "when-warehouse-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-warehouse-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-warehouse-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-warehouse-bulk-deleted"
)

type WarehouseMessageQueueConfig struct{}

func (WarehouseMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (WarehouseMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (WarehouseMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (WarehouseMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (WarehouseMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (WarehouseMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
