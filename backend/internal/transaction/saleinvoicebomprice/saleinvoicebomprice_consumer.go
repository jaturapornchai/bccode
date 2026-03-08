package saleinvoicebomprice

import (
	"encoding/json"
	"smlcloudplatform/internal/transaction/saleinvoicebomprice/config"
	"smlcloudplatform/internal/transaction/saleinvoicebomprice/models"
	"smlcloudplatform/internal/transaction/saleinvoicebomprice/repositories"
	"smlcloudplatform/internal/transaction/saleinvoicebomprice/services"
	"smlcloudplatform/pkg/microservice"
	"time"

	pkgConfig "smlcloudplatform/internal/config"
)

type SaleInvoiceBomPriceConsumer struct {
	ms  *microservice.Microservice
	cfg pkgConfig.IConfig
	svc services.ISaleInvoiceBomPriceConsumerService
}

func InitSaleInvoiceBomPriceConsumer(ms *microservice.Microservice, cfg pkgConfig.IConfig) *SaleInvoiceBomPriceConsumer {

	repo := repositories.NewSaleInvoiceBomPricePostgresRepository(ms.Persister(cfg.PersisterConfig()))
	bomConsumerService := services.NewSaleInvoiceBomPriceConsumerService(repo)

	bomTransactionConsumer := NewSaleInvoiceBomPriceConsumer(ms, cfg, bomConsumerService)
	return bomTransactionConsumer
}

func NewSaleInvoiceBomPriceConsumer(
	ms *microservice.Microservice,
	cfg pkgConfig.IConfig,
	svc services.ISaleInvoiceBomPriceConsumerService,
) *SaleInvoiceBomPriceConsumer {

	return &SaleInvoiceBomPriceConsumer{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (c *SaleInvoiceBomPriceConsumer) RegisterConsumer(ms *microservice.Microservice) {

	trxConsumerGroup := pkgConfig.GetEnv("PRODUCT_SaleInvoiceBomPrice_CONSUMER_GROUP", "product-bom-consumer-group-01")
	mq := microservice.NewMQ(c.cfg.MQConfig(), ms.Logger)

	kafkaConfig := config.SaleInvoiceBOMPriceMessageQueueConfig{}

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

func (c *SaleInvoiceBomPriceConsumer) ConsumeOnCreateOrUpdate(ctx microservice.IContext) error {

	input := ctx.ReadInput()

	doc := models.SaleInvoiceBomPricePg{}
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

func (c *SaleInvoiceBomPriceConsumer) ConsumeOnDelete(ctx microservice.IContext) error {

	input := ctx.ReadInput()
	doc := models.SaleInvoiceBomPricePg{}
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

func (c *SaleInvoiceBomPriceConsumer) ConsumeOnBulkCreateOrUpdate(ctx microservice.IContext) error {
	input := ctx.ReadInput()

	docs := []models.SaleInvoiceBomPricePg{}
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

func (c *SaleInvoiceBomPriceConsumer) ConsumeOnBulkDelete(ctx microservice.IContext) error {
	input := ctx.ReadInput()
	docs := []models.SaleInvoiceBomPricePg{}
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
