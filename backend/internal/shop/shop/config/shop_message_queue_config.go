package config

const (
	MQ_TOPIC_CREATED string = "when-shop-created"
)

type ShopMessageQueueConfig struct{}

func (ShopMessageQueueConfig) TopicCreated() string {
	return MQ_TOPIC_CREATED
}
