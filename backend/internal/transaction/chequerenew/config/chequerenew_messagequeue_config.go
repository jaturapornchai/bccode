package config

const (
	MQ_TOPIC_CREATED      string = "when-chequerenew-created"
	MQ_TOPIC_UPDATED      string = "when-chequerenew-updated"
	MQ_TOPIC_DELETED      string = "when-chequerenew-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-chequerenew-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-chequerenew-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-chequerenew-bulk-deleted"
)

type ChequeRenewMessageQueueConfig struct{}

func (ChequeRenewMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (ChequeRenewMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (ChequeRenewMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (ChequeRenewMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (ChequeRenewMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (ChequeRenewMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
