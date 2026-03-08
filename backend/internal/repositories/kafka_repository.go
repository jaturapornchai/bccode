package repositories

import (
	"smlcloudplatform/pkg/microservice"
)

type KafkaConfig interface {
	TopicCreated() string
	TopicUpdated() string
	TopicDeleted() string
	TopicBulkDeleted() string
	TopicBulkCreated() string
	TopicBulkUpdated() string
}

type KafkaRepository[T any] struct {
	topic KafkaConfig
	prod  microservice.IProducer
	mqKey string
}

func NewKafkaRepository[T any](prod microservice.IProducer, topicConfig KafkaConfig, mqKey string) KafkaRepository[T] {
	return KafkaRepository[T]{
		prod:  prod,
		topic: topicConfig,
		mqKey: mqKey,
	}
}

func (repo KafkaRepository[T]) Create(doc T) error {
	err := repo.prod.SendMessage(repo.topic.TopicCreated(), repo.mqKey, doc)

	if err != nil {
		return err
	}

	return nil
}

func (repo KafkaRepository[T]) Update(doc T) error {
	err := repo.prod.SendMessage(repo.topic.TopicUpdated(), repo.mqKey, doc)

	if err != nil {
		return err
	}

	return nil
}

func (repo KafkaRepository[T]) Delete(doc T) error {
	err := repo.prod.SendMessage(repo.topic.TopicDeleted(), repo.mqKey, doc)

	if err != nil {
		return err
	}

	return nil
}

func (repo KafkaRepository[T]) CreateInBatch(docList []T) error {
	err := repo.prod.SendMessage(repo.topic.TopicBulkCreated(), repo.mqKey, docList)

	if err != nil {
		return err
	}

	return nil
}

func (repo KafkaRepository[T]) UpdateInBatch(docList []T) error {
	err := repo.prod.SendMessage(repo.topic.TopicBulkUpdated(), repo.mqKey, docList)

	if err != nil {
		return err
	}

	return nil
}

func (repo KafkaRepository[T]) DeleteInBatch(docList []T) error {
	err := repo.prod.SendMessage(repo.topic.TopicBulkDeleted(), repo.mqKey, docList)

	if err != nil {
		return err
	}

	return nil
}
