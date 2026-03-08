package config

const (
	MQ_TOPIC_CREATED      string = "when-purchasereturn-created"
	MQ_TOPIC_UPDATED      string = "when-purchasereturn-updated"
	MQ_TOPIC_DELETED      string = "when-purchasereturn-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-purchasereturn-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-purchasereturn-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-purchasereturn-bulk-deleted"
)

type PurchaseReturnMessageQueueConfig struct{}

func (PurchaseReturnMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (PurchaseReturnMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (PurchaseReturnMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (PurchaseReturnMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (PurchaseReturnMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (PurchaseReturnMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
