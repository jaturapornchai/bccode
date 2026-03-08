package config

const (
	MQ_TOPIC_CREATED      string = "when-journal-created"
	MQ_TOPIC_UPDATED      string = "when-journal-updated"
	MQ_TOPIC_DELETED      string = "when-journal-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-journal-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-journal-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-journal-bulk-deleted"
)

type JournalMessageQueueConfig struct{}

func (JournalMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (JournalMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (JournalMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (JournalMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (JournalMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (JournalMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
