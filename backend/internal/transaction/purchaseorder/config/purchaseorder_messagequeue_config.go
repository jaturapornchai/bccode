package config

const (
	MQ_TOPIC_CREATED      string = "when-purchaseorder-created"
	MQ_TOPIC_UPDATED      string = "when-purchaseorder-updated"
	MQ_TOPIC_DELETED      string = "when-purchaseorder-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-purchaseorder-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-purchaseorder-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-purchaseorder-bulk-deleted"
)

type PurchaseOrderMessageQueueConfig struct{}

func (PurchaseOrderMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (PurchaseOrderMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (PurchaseOrderMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (PurchaseOrderMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (PurchaseOrderMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (PurchaseOrderMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
