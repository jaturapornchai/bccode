package config

const (
	MQ_TOPIC_CREATED      string = "when-chartofaccount-created"
	MQ_TOPIC_UPDATED      string = "when-chartofaccount-updated"
	MQ_TOPIC_DELETED      string = "when-chartofaccount-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-chartofaccount-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-chartofaccount-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-chartofaccount-bulk-deleted"
)

type ChartOfAccountMessageQueueConfig struct{}

func (ChartOfAccountMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (ChartOfAccountMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (ChartOfAccountMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (ChartOfAccountMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (ChartOfAccountMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (ChartOfAccountMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
