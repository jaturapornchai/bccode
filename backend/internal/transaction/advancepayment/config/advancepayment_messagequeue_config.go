package config

const (
	MQ_TOPIC_CREATED      string = "when-advancepayment-created"
	MQ_TOPIC_UPDATED      string = "when-advancepayment-updated"
	MQ_TOPIC_DELETED      string = "when-advancepayment-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-advancepayment-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-advancepayment-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-advancepayment-bulk-deleted"
)

type AdvancePaymentMessageQueueConfig struct{}

func (AdvancePaymentMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (AdvancePaymentMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (AdvancePaymentMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (AdvancePaymentMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (AdvancePaymentMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (AdvancePaymentMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
