package config

const (
	MQ_TOPIC_CREATED      string = "when-organization-jobproject-created"
	MQ_TOPIC_UPDATED      string = "when-organization-jobproject-updated"
	MQ_TOPIC_DELETED      string = "when-organization-jobproject-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-organization-jobproject-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-organization-jobproject-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-organization-jobproject-bulk-deleted"
)

type JobProjectMessageQueueConfig struct{}

func (JobProjectMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (JobProjectMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (JobProjectMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (JobProjectMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (JobProjectMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (JobProjectMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
