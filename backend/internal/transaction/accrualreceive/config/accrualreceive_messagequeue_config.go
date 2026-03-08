package config

const (
	MQ_TOPIC_CREATED      string = "when-accrualreceive-created"
	MQ_TOPIC_UPDATED      string = "when-accrualreceive-updated"
	MQ_TOPIC_DELETED      string = "when-accrualreceive-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-accrualreceive-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-accrualreceive-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-accrualreceive-bulk-deleted"
)

type AccrualreceiveMessageQueueConfig struct{}

func (AccrualreceiveMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (AccrualreceiveMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (AccrualreceiveMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (AccrualreceiveMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (AccrualreceiveMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (AccrualreceiveMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
