package config

const (
	MQ_TOPIC_CREATED      string = "when-journalbook-created"
	MQ_TOPIC_UPDATED      string = "when-journalbook-updated"
	MQ_TOPIC_DELETED      string = "when-journalbook-deleted"
	MQ_TOPIC_BULK_CREATED string = "when-journalbook-bulk-created"
	MQ_TOPIC_BULK_UPDATED string = "when-journalbook-bulk-updated"
	MQ_TOPIC_BULK_DELETED string = "when-journalbook-bulk-deleted"
)

type JournalBookMessageQueueConfig struct{}

func (JournalBookMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}

func (JournalBookMessageQueueConfig) TopicUpdated() string {
	return MQ_TOPIC_UPDATED
}

func (JournalBookMessageQueueConfig) TopicDeleted() string {
	return MQ_TOPIC_DELETED
}

func (JournalBookMessageQueueConfig) TopicBulkCreated() string {
	return MQ_TOPIC_BULK_CREATED
}

func (JournalBookMessageQueueConfig) TopicBulkUpdated() string {
	return MQ_TOPIC_BULK_UPDATED
}

func (JournalBookMessageQueueConfig) TopicBulkDeleted() string {
	return MQ_TOPIC_BULK_DELETED
}
