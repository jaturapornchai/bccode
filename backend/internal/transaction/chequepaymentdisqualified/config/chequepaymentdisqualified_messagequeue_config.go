package config

const (
	MQ_TOPIC_CREATED      string = "when-chequepaymentdisqualified-created"
	MQ_TOPIC_UPDATED      string = "when-chequepaymentdisqualified-updated"
	MQ_TOPIC_DELETED      string = "when-chequepaymentdisqualified-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-chequepaymentdisqualified-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-chequepaymentdisqualified-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-chequepaymentdisqualified-bulk-deleted"
)

type ChequePaymentDisqualifiedMessageQueueConfig struct{}

func (ChequePaymentDisqualifiedMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (ChequePaymentDisqualifiedMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (ChequePaymentDisqualifiedMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (ChequePaymentDisqualifiedMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (ChequePaymentDisqualifiedMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (ChequePaymentDisqualifiedMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
