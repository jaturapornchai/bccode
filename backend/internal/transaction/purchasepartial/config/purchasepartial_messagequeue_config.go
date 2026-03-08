package config

const (
	MQ_TOPIC_CREATED      string = "when-purchasepartial-created"
	MQ_TOPIC_UPDATED      string = "when-purchasepartial-updated"
	MQ_TOPIC_DELETED      string = "when-purchasepartial-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-purchasepartial-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-purchasepartial-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-purchasepartial-bulk-deleted"
)

type PurchasepartialMessageQueueConfig struct{}

func (PurchasepartialMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (PurchasepartialMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (PurchasepartialMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (PurchasepartialMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (PurchasepartialMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (PurchasepartialMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
