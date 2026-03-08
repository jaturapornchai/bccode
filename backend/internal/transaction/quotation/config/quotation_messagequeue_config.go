package config

const (
	MQ_TOPIC_CREATED      string = "when-quotation-created"
	MQ_TOPIC_UPDATED      string = "when-quotation-updated"
	MQ_TOPIC_DELETED      string = "when-quotation-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-quotation-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-quotation-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-quotation-bulk-deleted"
)

type QuotationMessageQueueConfig struct{}

func (QuotationMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (QuotationMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (QuotationMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (QuotationMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (QuotationMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (QuotationMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
