package config

const (
	MQ_TOPIC_CREATED      string = "when-organization-costcenter-created"
	MQ_TOPIC_UPDATED      string = "when-organization-costcenter-updated"
	MQ_TOPIC_DELETED      string = "when-organization-costcenter-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-organization-costcenter-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-organization-costcenter-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-organization-costcenter-bulk-deleted"
)

type CostCenterMessageQueueConfig struct{}

func (CostCenterMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (CostCenterMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (CostCenterMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (CostCenterMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (CostCenterMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (CostCenterMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
