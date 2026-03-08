package shift

import (
	"encoding/json"
	shiftConfig "smlcloudplatform/internal/pos/shift/config"
	"smlcloudplatform/internal/pos/shift/models"
	"smlcloudplatform/internal/pos/shift/repositories"
	"smlcloudplatform/internal/pos/shift/services"
	"smlcloudplatform/pkg/microservice"
	"time"

	pkgConfig "smlcloudplatform/internal/config"
)

type ShiftConsumer struct {
	ms  *microservice.Microservice
	cfg pkgConfig.IConfig
	svc services.IShiftConsumerService
}

func InitShiftConsumer(ms *microservice.Microservice, cfg pkgConfig.IConfig) *ShiftConsumer {

	repo := repositories.NewShiftPostgresRepository(ms.Persister(cfg.PersisterConfig()))
	shiftConsumerService := services.NewShiftConsumerService(repo)

	shiftTransactionConsumer := NewShiftConsumer(ms, cfg, shiftConsumerService)
	return shiftTransactionConsumer
}

func NewShiftConsumer(
	ms *microservice.Microservice,
	cfg pkgConfig.IConfig,
	svc services.IShiftConsumerService,
) *ShiftConsumer {

	return &ShiftConsumer{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (c *ShiftConsumer) RegisterConsumer(ms *microservice.Microservice) {

	trxConsumerGroup := pkgConfig.GetEnv("TRANSACTION_CONSUMER_GROUP", "transaction-consumer-group-01")
	mq := microservice.NewMQ(c.cfg.MQConfig(), ms.Logger)

	kafkaConfig := shiftConfig.ShiftMessageQueueConfig{}

	mq.CreateTopicR(kafkaConfig.TopicCreated(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(kafkaConfig.TopicUpdated(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(kafkaConfig.TopicDeleted(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(kafkaConfig.TopicBulkCreated(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(kafkaConfig.TopicBulkUpdated(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(kafkaConfig.TopicBulkDeleted(), 5, 1, time.Hour*24*7)

	ms.Consume(c.cfg.MQConfig().URI(), kafkaConfig.TopicCreated(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnCreateOrUpdate)
	ms.Consume(c.cfg.MQConfig().URI(), kafkaConfig.TopicUpdated(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnCreateOrUpdate)
	ms.Consume(c.cfg.MQConfig().URI(), kafkaConfig.TopicDeleted(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnDelete)
	ms.Consume(c.cfg.MQConfig().URI(), kafkaConfig.TopicBulkCreated(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnBulkCreateOrUpdate)
	ms.Consume(c.cfg.MQConfig().URI(), kafkaConfig.TopicBulkUpdated(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnBulkCreateOrUpdate)
	ms.Consume(c.cfg.MQConfig().URI(), kafkaConfig.TopicBulkDeleted(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnBulkDelete)

}

func (c *ShiftConsumer) ConsumeOnCreateOrUpdate(ctx microservice.IContext) error {

	input := ctx.ReadInput()

	doc := models.ShiftPG{}
	err := json.Unmarshal([]byte(input), &doc)
	if err != nil {
		return err
	}

	err = c.svc.Upsert(doc.ShopID, doc.GuidFixed, doc)
	if err != nil {
		return err
	}

	return nil
}

func (c *ShiftConsumer) ConsumeOnDelete(ctx microservice.IContext) error {

	input := ctx.ReadInput()
	doc := models.ShiftPG{}
	err := json.Unmarshal([]byte(input), &doc)
	if err != nil {
		return err
	}

	err = c.svc.Delete(doc.ShopID, doc.GuidFixed)
	if err != nil {
		return err
	}

	return nil
}

func (c *ShiftConsumer) ConsumeOnBulkCreateOrUpdate(ctx microservice.IContext) error {
	input := ctx.ReadInput()

	docs := []models.ShiftPG{}
	err := json.Unmarshal([]byte(input), &docs)
	if err != nil {
		return err
	}

	for _, doc := range docs {
		err = c.svc.Upsert(doc.ShopID, doc.GuidFixed, doc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *ShiftConsumer) ConsumeOnBulkDelete(ctx microservice.IContext) error {
	input := ctx.ReadInput()
	docs := []models.ShiftPG{}
	err := json.Unmarshal([]byte(input), &docs)
	if err != nil {
		return err
	}

	for _, doc := range docs {
		err = c.svc.Delete(doc.ShopID, doc.GuidFixed)
		if err != nil {
			return err
		}
	}
	return nil
}
