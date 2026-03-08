package bom

import (
	"encoding/json"
	bomConfig "smlcloudplatform/internal/product/bom/config"
	"smlcloudplatform/internal/product/bom/models"
	"smlcloudplatform/internal/product/bom/repositories"
	"smlcloudplatform/internal/product/bom/services"
	"smlcloudplatform/pkg/microservice"
	"time"

	pkgConfig "smlcloudplatform/internal/config"
)

type BOMConsumer struct {
	ms  *microservice.Microservice
	cfg pkgConfig.IConfig
	svc services.IBOMConsumerService
}

func InitBOMConsumer(ms *microservice.Microservice, cfg pkgConfig.IConfig) *BOMConsumer {

	repo := repositories.NewBOMPostgresRepository(ms.Persister(cfg.PersisterConfig()))
	bomConsumerService := services.NewBOMConsumerService(repo)

	bomTransactionConsumer := NewBOMConsumer(ms, cfg, bomConsumerService)
	return bomTransactionConsumer
}

func NewBOMConsumer(
	ms *microservice.Microservice,
	cfg pkgConfig.IConfig,
	svc services.IBOMConsumerService,
) *BOMConsumer {

	return &BOMConsumer{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (c *BOMConsumer) RegisterConsumer(ms *microservice.Microservice) {

	trxConsumerGroup := pkgConfig.GetEnv("PRODUCT_BOM_CONSUMER_GROUP", "product-bom-consumer-group-01")
	mq := microservice.NewMQ(c.cfg.MQConfig(), ms.Logger)

	kafkaConfig := bomConfig.BomMessageQueueConfig{}

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

func (c *BOMConsumer) ConsumeOnCreateOrUpdate(ctx microservice.IContext) error {

	input := ctx.ReadInput()

	doc := models.ProductBarcodeBOMViewPG{}
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

func (c *BOMConsumer) ConsumeOnDelete(ctx microservice.IContext) error {

	input := ctx.ReadInput()
	doc := models.ProductBarcodeBOMViewPG{}
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

func (c *BOMConsumer) ConsumeOnBulkCreateOrUpdate(ctx microservice.IContext) error {
	input := ctx.ReadInput()

	docs := []models.ProductBarcodeBOMViewPG{}
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

func (c *BOMConsumer) ConsumeOnBulkDelete(ctx microservice.IContext) error {
	input := ctx.ReadInput()
	docs := []models.ProductBarcodeBOMViewPG{}
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
