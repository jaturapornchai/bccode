package config

const (
	MQ_TOPIC_CREATED      string = "when-creditor-payment-created"
	MQ_TOPIC_UPDATED      string = "when-creditor-payment-updated"
	MQ_TOPIC_DELETED      string = "when-creditor-payment-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-creditor-payment-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-creditor-payment-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-creditor-payment-bulk-deleted"
)

type CreditorPaymentMessageQueueConfig struct{}

func (CreditorPaymentMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (CreditorPaymentMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (CreditorPaymentMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (CreditorPaymentMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (CreditorPaymentMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (CreditorPaymentMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
