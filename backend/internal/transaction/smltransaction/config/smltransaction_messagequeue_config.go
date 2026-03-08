package config

const (
	MQ_TOPIC_SAVED      string = "when-smltransaction-saved"
	MQ_TOPIC_BULK_SAVED string = "when-smltransaction-bulk-saved"
	MQ_TOPIC_DELETED    string = "when-smltransaction-deleted"
)

type SMLTransactionMessageQueueConfig struct{}

func (SMLTransactionMessageQueueConfig) TopicSaved() string {
	return MQ_TOPIC_SAVED
}

func (SMLTransactionMessageQueueConfig) TopicBulkSaved() string {
	return MQ_TOPIC_SAVED
}

func (SMLTransactionMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}
