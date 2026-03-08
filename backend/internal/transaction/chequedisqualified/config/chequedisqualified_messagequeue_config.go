package config

const (
	MQ_TOPIC_CREATED      string = "when-chequedisqualified-created"
	MQ_TOPIC_UPDATED      string = "when-chequedisqualified-updated"
	MQ_TOPIC_DELETED      string = "when-chequedisqualified-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-chequedisqualified-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-chequedisqualified-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-chequedisqualified-bulk-deleted"
)

type ChequeDisqualifiedMessageQueueConfig struct{}

func (ChequeDisqualifiedMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (ChequeDisqualifiedMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (ChequeDisqualifiedMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (ChequeDisqualifiedMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (ChequeDisqualifiedMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (ChequeDisqualifiedMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
