package config

const (
	MQ_TOPIC_CREATED      string = "when-saleinvoice-created"
	MQ_TOPIC_UPDATED      string = "when-saleinvoice-updated"
	MQ_TOPIC_DELETED      string = "when-saleinvoice-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-saleinvoice-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-saleinvoice-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-saleinvoice-bulk-deleted"
)

type SaleInvoiceMessageQueueConfig struct{}

func (SaleInvoiceMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (SaleInvoiceMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (SaleInvoiceMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (SaleInvoiceMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (SaleInvoiceMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (SaleInvoiceMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
