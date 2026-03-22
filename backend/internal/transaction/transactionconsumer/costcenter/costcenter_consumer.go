package costcenter

import (
	"encoding/json"
	pkgConfig "smlcloudplatform/internal/config"
	"smlcloudplatform/internal/logger"
	costcenterConfig "smlcloudplatform/internal/organization/costcenter/config"
	"smlcloudplatform/internal/organization/costcenter/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/services"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type CostCenterConsumer struct {
	ms  *microservice.Microservice
	cfg pkgConfig.IConfig
	svc ICostCenterConsumerService
}

func NewCostCenterConsumer(
	ms *microservice.Microservice,
	cfg pkgConfig.IConfig,
	svc ICostCenterConsumerService,
) services.ITransactionDocConsumer {
	return &CostCenterConsumer{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func InitCostCenterConsumer(ms *microservice.Microservice, cfg pkgConfig.IConfig) services.ITransactionDocConsumer {
	pgPersister := ms.Persister(cfg.PersisterConfig())
	pgRepo := NewCostCenterPGRepository(pgPersister)

	clickhouseCfg := cfg.ClickHouseConfig()
	var chRepo ICostCenterCHRepository
	if len(clickhouseCfg.ServerAddress()) > 0 {
		chRepo = NewCostCenterCHRepository(ms.ClickHousePersister(clickhouseCfg))
	}

	svc := NewCostCenterConsumerService(pgRepo, chRepo)
	return NewCostCenterConsumer(ms, cfg, svc)
}

func (c *CostCenterConsumer) RegisterConsumer(ms *microservice.Microservice) {
	trxConsumerGroup := pkgConfig.GetEnv("TRANSACTION_CONSUMER_GROUP", "transaction-consumer-group-01")
	mq := microservice.NewMQ(c.cfg.MQConfig(), ms.Logger)
	mqConfig := costcenterConfig.CostCenterMessageQueueConfig{}

	mq.CreateTopicR(mqConfig.TopicCreated(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(mqConfig.TopicUpdated(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(mqConfig.TopicDeleted(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(mqConfig.TopicBulkCreated(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(mqConfig.TopicBulkUpdated(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(mqConfig.TopicBulkDeleted(), 5, 1, time.Hour*24*7)

	ms.Consume(c.cfg.MQConfig().URI(), mqConfig.TopicCreated(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnCreateOrUpdate)
	ms.Consume(c.cfg.MQConfig().URI(), mqConfig.TopicUpdated(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnCreateOrUpdate)
	ms.Consume(c.cfg.MQConfig().URI(), mqConfig.TopicDeleted(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnDelete)
	ms.Consume(c.cfg.MQConfig().URI(), mqConfig.TopicBulkCreated(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnBulkCreateOrUpdate)
	ms.Consume(c.cfg.MQConfig().URI(), mqConfig.TopicBulkUpdated(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnBulkCreateOrUpdate)
	ms.Consume(c.cfg.MQConfig().URI(), mqConfig.TopicBulkDeleted(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnBulkDelete)
}

func (c *CostCenterConsumer) ConsumeOnCreateOrUpdate(ctx microservice.IContext) error {
	msg := ctx.ReadInput()
	doc := models.CostCenterDoc{}
	err := json.Unmarshal([]byte(msg), &doc)
	if err != nil {
		logger.GetLogger().Errorf("Cannot unmarshal costcenter doc: %v", err.Error())
		return err
	}
	err = c.svc.Upsert(doc.ShopID, doc.GuidFixed, doc)
	if err != nil {
		logger.GetLogger().Errorf("Cannot upsert costcenter pg: %v", err.Error())
		return err
	}
	return nil
}

func (c *CostCenterConsumer) ConsumeOnBulkCreateOrUpdate(ctx microservice.IContext) error {
	msg := ctx.ReadInput()
	docs := []models.CostCenterDoc{}
	err := json.Unmarshal([]byte(msg), &docs)
	if err != nil {
		logger.GetLogger().Errorf("Cannot unmarshal costcenter bulk doc: %v", err.Error())
		return err
	}
	for _, doc := range docs {
		err = c.svc.Upsert(doc.ShopID, doc.GuidFixed, doc)
		if err != nil {
			logger.GetLogger().Errorf("Cannot upsert costcenter pg: %v", err.Error())
			return err
		}
	}
	return nil
}

func (c *CostCenterConsumer) ConsumeOnDelete(ctx microservice.IContext) error {
	msg := ctx.ReadInput()
	doc := models.CostCenterDoc{}
	err := json.Unmarshal([]byte(msg), &doc)
	if err != nil {
		logger.GetLogger().Errorf("Cannot unmarshal costcenter doc: %v", err.Error())
		return err
	}
	err = c.svc.Delete(doc.ShopID, doc.GuidFixed)
	if err != nil {
		logger.GetLogger().Errorf("Cannot delete costcenter pg: %v", err.Error())
		return err
	}
	return nil
}

func (c *CostCenterConsumer) ConsumeOnBulkDelete(ctx microservice.IContext) error {
	msg := ctx.ReadInput()
	docs := []models.CostCenterDoc{}
	err := json.Unmarshal([]byte(msg), &docs)
	if err != nil {
		logger.GetLogger().Errorf("Cannot unmarshal costcenter bulk doc: %v", err.Error())
		return err
	}
	for _, doc := range docs {
		err = c.svc.Delete(doc.ShopID, doc.GuidFixed)
		if err != nil {
			logger.GetLogger().Errorf("Cannot delete costcenter pg: %v", err.Error())
			return err
		}
	}
	return nil
}
