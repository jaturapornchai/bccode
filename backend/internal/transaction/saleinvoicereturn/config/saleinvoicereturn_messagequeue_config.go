package config

const (
	MQ_TOPIC_CREATED      string = "when-saleinvoicereturn-created"
	MQ_TOPIC_UPDATED      string = "when-saleinvoicereturn-updated"
	MQ_TOPIC_DELETED      string = "when-saleinvoicereturn-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-saleinvoicereturn-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-saleinvoicereturn-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-saleinvoicereturn-bulk-deleted"
)

type SaleInvoiceReturnMessageQueueConfig struct{}

func (SaleInvoiceReturnMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (SaleInvoiceReturnMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (SaleInvoiceReturnMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (SaleInvoiceReturnMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (SaleInvoiceReturnMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (SaleInvoiceReturnMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
