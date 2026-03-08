package config

const (
	MQ_TOPIC_CREATED      string = "when-bom-created"
	MQ_TOPIC_UPDATED      string = "when-bom-updated"
	MQ_TOPIC_DELETED      string = "when-bom-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-bom-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-bom-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-bom-bulk-deleted"
)

type BomMessageQueueConfig struct {
}

func (BomMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (BomMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (BomMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (BomMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (BomMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (BomMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
