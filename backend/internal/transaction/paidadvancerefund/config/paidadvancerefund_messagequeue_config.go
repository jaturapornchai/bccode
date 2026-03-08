package config

const (
	MQ_TOPIC_CREATED      string = "when-paidadvancerefund-created"
	MQ_TOPIC_UPDATED      string = "when-paidadvancerefund-updated"
	MQ_TOPIC_DELETED      string = "when-paidadvancerefund-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-paidadvancerefund-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-paidadvancerefund-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-paidadvancerefund-bulk-deleted"
)

type PaidAdvanceRefundMessageQueueConfig struct{}

func (PaidAdvanceRefundMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (PaidAdvanceRefundMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (PaidAdvanceRefundMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (PaidAdvanceRefundMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (PaidAdvanceRefundMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (PaidAdvanceRefundMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
