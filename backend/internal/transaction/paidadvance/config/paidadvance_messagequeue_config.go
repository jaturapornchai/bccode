package config

const (
	MQ_TOPIC_CREATED      string = "when-paidadvance-created"
	MQ_TOPIC_UPDATED      string = "when-paidadvance-updated"
	MQ_TOPIC_DELETED      string = "when-paidadvance-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-paidadvance-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-paidadvance-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-paidadvance-bulk-deleted"
)

type PaidAdvanceMessageQueueConfig struct{}

func (PaidAdvanceMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (PaidAdvanceMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (PaidAdvanceMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (PaidAdvanceMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (PaidAdvanceMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (PaidAdvanceMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
