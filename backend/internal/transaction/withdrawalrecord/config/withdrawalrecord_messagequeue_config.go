package config

const (
	MQ_TOPIC_CREATED      string = "when-withdrawalrecord-created"
	MQ_TOPIC_UPDATED      string = "when-withdrawalrecord-updated"
	MQ_TOPIC_DELETED      string = "when-withdrawalrecord-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-withdrawalrecord-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-withdrawalrecord-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-withdrawalrecord-bulk-deleted"
)

type WithdrawalRecordMessageQueueConfig struct{}

func (WithdrawalRecordMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (WithdrawalRecordMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (WithdrawalRecordMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (WithdrawalRecordMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (WithdrawalRecordMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (WithdrawalRecordMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
