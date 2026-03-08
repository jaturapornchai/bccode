package config

const (
	MQ_TOPIC_CREATED      string = "when-advancepaymentrefund-created"
	MQ_TOPIC_UPDATED      string = "when-advancepaymentrefund-updated"
	MQ_TOPIC_DELETED      string = "when-advancepaymentrefund-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-advancepaymentrefund-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-advancepaymentrefund-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-advancepaymentrefund-bulk-deleted"
)

type AdvancePaymentRefundMessageQueueConfig struct{}

func (AdvancePaymentRefundMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (AdvancePaymentRefundMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (AdvancePaymentRefundMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (AdvancePaymentRefundMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (AdvancePaymentRefundMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (AdvancePaymentRefundMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
