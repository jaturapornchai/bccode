package debtor

import (
	"encoding/json"
	debtorConfig "smlcloudplatform/internal/debtaccount/debtor/config"
	"smlcloudplatform/internal/debtaccount/debtor/models"
	"smlcloudplatform/internal/debtaccount/debtor/repositories"
	"smlcloudplatform/internal/debtaccount/debtor/services"
	"smlcloudplatform/pkg/microservice"
	"time"

	pkgConfig "smlcloudplatform/internal/config"
)

type DebtorConsumer struct {
	ms  *microservice.Microservice
	cfg pkgConfig.IConfig
	svc services.IDebtorConsumerService
}

func InitDebtorConsumer(ms *microservice.Microservice, cfg pkgConfig.IConfig) *DebtorConsumer {

	repo := repositories.NewDebtorPostgresRepository(ms.Persister(cfg.PersisterConfig()))
	debtorConsumerService := services.NewDebtorConsumerService(repo)

	debtorTransactionConsumer := NewDebtorConsumer(ms, cfg, debtorConsumerService)
	return debtorTransactionConsumer
}

func NewDebtorConsumer(
	ms *microservice.Microservice,
	cfg pkgConfig.IConfig,
	svc services.IDebtorConsumerService,
) *DebtorConsumer {

	return &DebtorConsumer{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (c *DebtorConsumer) RegisterConsumer(ms *microservice.Microservice) {

	trxConsumerGroup := pkgConfig.GetEnv("DEBT_ACCOUNT_CONSUMER_GROUP", "debt-account-consumer-group-01")
	mq := microservice.NewMQ(c.cfg.MQConfig(), ms.Logger)

	kafkaConfig := debtorConfig.DebtorMessageQueueConfig{}

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

func (c *DebtorConsumer) ConsumeOnCreateOrUpdate(ctx microservice.IContext) error {

	input := ctx.ReadInput()

	doc := models.DebtorPG{}
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

func (c *DebtorConsumer) ConsumeOnDelete(ctx microservice.IContext) error {

	input := ctx.ReadInput()
	doc := models.DebtorPG{}
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

func (c *DebtorConsumer) ConsumeOnBulkCreateOrUpdate(ctx microservice.IContext) error {
	input := ctx.ReadInput()

	docs := []models.DebtorPG{}
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

func (c *DebtorConsumer) ConsumeOnBulkDelete(ctx microservice.IContext) error {
	input := ctx.ReadInput()
	docs := []models.DebtorPG{}
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
