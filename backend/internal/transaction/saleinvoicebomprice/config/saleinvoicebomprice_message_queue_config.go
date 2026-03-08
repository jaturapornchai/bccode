package config

const (
	MQ_TOPIC_CREATED      string = "when-saleinvoice-bom-price-created"
	MQ_TOPIC_UPDATED      string = "when-saleinvoice-bom-price-updated"
	MQ_TOPIC_DELETED      string = "when-saleinvoice-bom-price-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-saleinvoice-bom-price-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-saleinvoice-bom-price-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-saleinvoice-bom-price-bulk-deleted"
)

type SaleInvoiceBOMPriceMessageQueueConfig struct {
}

func (SaleInvoiceBOMPriceMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (SaleInvoiceBOMPriceMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (SaleInvoiceBOMPriceMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (SaleInvoiceBOMPriceMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (SaleInvoiceBOMPriceMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (SaleInvoiceBOMPriceMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
