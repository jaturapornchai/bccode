package creditor

import (
	"encoding/json"
	creditorConfig "smlcloudplatform/internal/debtaccount/creditor/config"
	"smlcloudplatform/internal/debtaccount/creditor/models"
	"smlcloudplatform/internal/debtaccount/creditor/repositories"
	"smlcloudplatform/internal/debtaccount/creditor/services"
	"smlcloudplatform/pkg/microservice"
	"time"

	pkgConfig "smlcloudplatform/internal/config"
)

type CreditorConsumer struct {
	ms  *microservice.Microservice
	cfg pkgConfig.IConfig
	svc services.ICreditorConsumerService
}

func InitCreditorConsumer(ms *microservice.Microservice, cfg pkgConfig.IConfig) *CreditorConsumer {

	repo := repositories.NewCreditorPostgresRepository(ms.Persister(cfg.PersisterConfig()))
	creditorConsumerService := services.NewCreditorConsumerService(repo)

	creditorTransactionConsumer := NewCreditorConsumer(ms, cfg, creditorConsumerService)
	return creditorTransactionConsumer
}

func NewCreditorConsumer(
	ms *microservice.Microservice,
	cfg pkgConfig.IConfig,
	svc services.ICreditorConsumerService,
) *CreditorConsumer {

	return &CreditorConsumer{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (c *CreditorConsumer) RegisterConsumer(ms *microservice.Microservice) {

	trxConsumerGroup := pkgConfig.GetEnv("DEBT_ACCOUNT_CONSUMER_GROUP", "debt-account-consumer-group-01")
	mq := microservice.NewMQ(c.cfg.MQConfig(), ms.Logger)

	kafkaConfig := creditorConfig.CreditorMessageQueueConfig{}

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

func (c *CreditorConsumer) ConsumeOnCreateOrUpdate(ctx microservice.IContext) error {

	input := ctx.ReadInput()

	doc := models.CreditorPG{}
	err := json.Unmarshal([]byte(input), &doc)
	if err != nil {
		return err
	}

	err = c.svc.Upsert(doc.ShopID, doc.Code, doc)
	if err != nil {
		return err
	}

	return nil
}

func (c *CreditorConsumer) ConsumeOnDelete(ctx microservice.IContext) error {

	input := ctx.ReadInput()
	doc := models.CreditorPG{}
	err := json.Unmarshal([]byte(input), &doc)
	if err != nil {
		return err
	}

	err = c.svc.Delete(doc.ShopID, doc.Code)
	if err != nil {
		return err
	}

	return nil
}

func (c *CreditorConsumer) ConsumeOnBulkCreateOrUpdate(ctx microservice.IContext) error {
	input := ctx.ReadInput()

	docs := []models.CreditorPG{}
	err := json.Unmarshal([]byte(input), &docs)
	if err != nil {
		return err
	}

	for _, doc := range docs {
		err = c.svc.Upsert(doc.ShopID, doc.Code, doc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *CreditorConsumer) ConsumeOnBulkDelete(ctx microservice.IContext) error {
	input := ctx.ReadInput()
	docs := []models.CreditorPG{}
	err := json.Unmarshal([]byte(input), &docs)
	if err != nil {
		return err
	}

	for _, doc := range docs {
		err = c.svc.Delete(doc.ShopID, doc.Code)
		if err != nil {
			return err
		}
	}
	return nil
}
