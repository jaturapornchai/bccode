package config

const (
	MQ_TOPIC_CREATED      string = "when-depositrecord-created"
	MQ_TOPIC_UPDATED      string = "when-depositrecord-updated"
	MQ_TOPIC_DELETED      string = "when-depositrecord-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-depositrecord-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-depositrecord-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-depositrecord-bulk-deleted"
)

type DepositRecordMessageQueueConfig struct{}

func (DepositRecordMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (DepositRecordMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (DepositRecordMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (DepositRecordMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (DepositRecordMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (DepositRecordMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
