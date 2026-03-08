package config

const (
	MQ_TOPIC_CREATED      string = "when-creditcardwithdrawal-created"
	MQ_TOPIC_UPDATED      string = "when-creditcardwithdrawal-updated"
	MQ_TOPIC_DELETED      string = "when-creditcardwithdrawal-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-creditcardwithdrawal-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-creditcardwithdrawal-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-creditcardwithdrawal-bulk-deleted"
)

type CreditCardWithdrawalMessageQueueConfig struct{}

func (CreditCardWithdrawalMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (CreditCardWithdrawalMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (CreditCardWithdrawalMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (CreditCardWithdrawalMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (CreditCardWithdrawalMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (CreditCardWithdrawalMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
