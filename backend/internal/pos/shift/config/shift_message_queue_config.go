package config

const (
	MQ_TOPIC_CREATED      string = "when-shift-created"
	MQ_TOPIC_UPDATED      string = "when-shift-updated"
	MQ_TOPIC_DELETED      string = "when-shift-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-shift-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-shift-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-shift-bulk-deleted"
)

type ShiftMessageQueueConfig struct {
}

func (ShiftMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (ShiftMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (ShiftMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (ShiftMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (ShiftMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (ShiftMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
