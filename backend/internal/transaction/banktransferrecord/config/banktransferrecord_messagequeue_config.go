package config

const (
	MQ_TOPIC_CREATED      string = "when-banktransferrecord-created"
	MQ_TOPIC_UPDATED      string = "when-banktransferrecord-updated"
	MQ_TOPIC_DELETED      string = "when-banktransferrecord-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-banktransferrecord-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-banktransferrecord-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-banktransferrecord-bulk-deleted"
)

type BankTransferRecordMessageQueueConfig struct{}

func (BankTransferRecordMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (BankTransferRecordMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (BankTransferRecordMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (BankTransferRecordMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (BankTransferRecordMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (BankTransferRecordMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
