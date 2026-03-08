package config

const (
	MQ_TOPIC_TASK_CHANGED  string = "when-documentimage-task-changed"
	MQ_TOPIC_TASK_REJECTED string = "when-documentimage-task-rejected"
)

type DocumentImageMessageQueueConfig struct{}

func (DocumentImageMessageQueueConfig) TopicTaskChanged() string {
	return MQ_TOPIC_TASK_CHANGED
}

func (DocumentImageMessageQueueConfig) TopicTaskRejected() string {
	return MQ_TOPIC_TASK_REJECTED
}
