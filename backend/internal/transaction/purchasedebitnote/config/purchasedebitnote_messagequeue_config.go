package config

const (
	MQ_TOPIC_CREATED      string = "when-purchasedebitnote-created"
	MQ_TOPIC_UPDATED      string = "when-purchasedebitnote-updated"
	MQ_TOPIC_DELETED      string = "when-purchasedebitnote-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-purchasedebitnote-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-purchasedebitnote-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-purchasedebitnote-bulk-deleted"
)

type PurchaseDebitNoteMessageQueueConfig struct{}

func (PurchaseDebitNoteMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (PurchaseDebitNoteMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (PurchaseDebitNoteMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (PurchaseDebitNoteMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (PurchaseDebitNoteMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (PurchaseDebitNoteMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
