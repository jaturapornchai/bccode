package config

const (
	MQ_TOPIC_CREATED      string = "when-chequepaymentreturn-created"
	MQ_TOPIC_UPDATED      string = "when-chequepaymentreturn-updated"
	MQ_TOPIC_DELETED      string = "when-chequepaymentreturn-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-chequepaymentreturn-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-chequepaymentreturn-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-chequepaymentreturn-bulk-deleted"
)

type ChequePaymentReturnMessageQueueConfig struct{}

func (ChequePaymentReturnMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (ChequePaymentReturnMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (ChequePaymentReturnMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (ChequePaymentReturnMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (ChequePaymentReturnMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (ChequePaymentReturnMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
