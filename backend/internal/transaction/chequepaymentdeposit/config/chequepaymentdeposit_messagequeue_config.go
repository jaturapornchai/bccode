package config

const (
	MQ_TOPIC_CREATED      string = "when-chequepaymentdeposit-created"
	MQ_TOPIC_UPDATED      string = "when-chequepaymentdeposit-updated"
	MQ_TOPIC_DELETED      string = "when-chequepaymentdeposit-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-chequepaymentdeposit-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-chequepaymentdeposit-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-chequepaymentdeposit-bulk-deleted"
)

type ChequePaymentDepositMessageQueueConfig struct{}

func (ChequePaymentDepositMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (ChequePaymentDepositMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (ChequePaymentDepositMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (ChequePaymentDepositMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (ChequePaymentDepositMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (ChequePaymentDepositMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
