package config

const (
	MQ_TOPIC_CREATED      string = "when-debtor-payment-created"
	MQ_TOPIC_UPDATED      string = "when-debtor-payment-updated"
	MQ_TOPIC_DELETED      string = "when-debtor-payment-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-debtor-payment-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-debtor-payment-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-debtor-payment-bulk-deleted"
)

type DebtorPaymentMessageQueueConfig struct{}

func (DebtorPaymentMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (DebtorPaymentMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (DebtorPaymentMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (DebtorPaymentMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (DebtorPaymentMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (DebtorPaymentMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
