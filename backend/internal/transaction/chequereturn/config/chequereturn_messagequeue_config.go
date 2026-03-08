package config

const (
	MQ_TOPIC_CREATED      string = "when-chequereturn-created"
	MQ_TOPIC_UPDATED      string = "when-chequereturn-updated"
	MQ_TOPIC_DELETED      string = "when-chequereturn-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-chequereturn-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-chequereturn-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-chequereturn-bulk-deleted"
)

type ChequeReturnMessageQueueConfig struct{}

func (ChequeReturnMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (ChequeReturnMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (ChequeReturnMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (ChequeReturnMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (ChequeReturnMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (ChequeReturnMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
