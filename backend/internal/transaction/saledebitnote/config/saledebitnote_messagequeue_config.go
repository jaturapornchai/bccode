package config

const (
	MQ_TOPIC_CREATED      string = "when-saledebitnote-created"
	MQ_TOPIC_UPDATED      string = "when-saledebitnote-updated"
	MQ_TOPIC_DELETED      string = "when-saledebitnote-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-saledebitnote-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-saledebitnote-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-saledebitnote-bulk-deleted"
)

type SaleDebitNoteMessageQueueConfig struct{}

func (SaleDebitNoteMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (SaleDebitNoteMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (SaleDebitNoteMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (SaleDebitNoteMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (SaleDebitNoteMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (SaleDebitNoteMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
