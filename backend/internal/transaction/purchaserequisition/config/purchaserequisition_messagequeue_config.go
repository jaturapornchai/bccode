package config

const (
	MQ_TOPIC_CREATED      string = "when-purchaserequisition-created"
	MQ_TOPIC_UPDATED      string = "when-purchaserequisition-updated"
	MQ_TOPIC_DELETED      string = "when-purchaserequisition-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-purchaserequisition-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-purchaserequisition-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-purchaserequisition-bulk-deleted"
)

type PurchaseRequisitionMessageQueueConfig struct{}

func (PurchaseRequisitionMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (PurchaseRequisitionMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (PurchaseRequisitionMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (PurchaseRequisitionMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (PurchaseRequisitionMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (PurchaseRequisitionMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
