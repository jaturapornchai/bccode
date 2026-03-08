package config

const (
	MQ_TOPIC_CREATED      string = "when-chequepaymentchange-created"
	MQ_TOPIC_UPDATED      string = "when-chequepaymentchange-updated"
	MQ_TOPIC_DELETED      string = "when-chequepaymentchange-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-chequepaymentchange-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-chequepaymentchange-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-chequepaymentchange-bulk-deleted"
)

type ChequePaymentChangeMessageQueueConfig struct{}

func (ChequePaymentChangeMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (ChequePaymentChangeMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (ChequePaymentChangeMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (ChequePaymentChangeMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (ChequePaymentChangeMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (ChequePaymentChangeMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
