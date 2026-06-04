package jobproject

import (
	"encoding/json"
	pkgConfig "smlcloudplatform/internal/config"
	"smlcloudplatform/internal/logger"
	jobprojectConfig "smlcloudplatform/internal/organization/jobproject/config"
	"smlcloudplatform/internal/organization/jobproject/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/services"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type JobProjectConsumer struct {
	ms  *microservice.Microservice
	cfg pkgConfig.IConfig
	svc IJobProjectConsumerService
}

func NewJobProjectConsumer(
	ms *microservice.Microservice,
	cfg pkgConfig.IConfig,
	svc IJobProjectConsumerService,
) services.ITransactionDocConsumer {
	return &JobProjectConsumer{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func InitJobProjectConsumer(ms *microservice.Microservice, cfg pkgConfig.IConfig) services.ITransactionDocConsumer {
	pgPersister := ms.Persister(cfg.PersisterConfig())
	pgRepo := NewJobProjectPGRepository(pgPersister)

	clickhouseCfg := cfg.ClickHouseConfig()
	var chRepo IJobProjectCHRepository
	if len(clickhouseCfg.ServerAddress()) > 0 {
		chRepo = NewJobProjectCHRepository(ms.ClickHousePersister(clickhouseCfg))
	}

	svc := NewJobProjectConsumerService(pgRepo, chRepo)
	return NewJobProjectConsumer(ms, cfg, svc)
}

func (c *JobProjectConsumer) RegisterConsumer(ms *microservice.Microservice) {
	trxConsumerGroup := pkgConfig.GetEnv("TRANSACTION_CONSUMER_GROUP", "transaction-consumer-group-01")
	mq := microservice.NewMQ(c.cfg.MQConfig(), ms.Logger)
	mqConfig := jobprojectConfig.JobProjectMessageQueueConfig{}

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

func (c *JobProjectConsumer) ConsumeOnCreateOrUpdate(ctx microservice.IContext) error {
	msg := ctx.ReadInput()
	doc := models.JobProjectDoc{}
	err := json.Unmarshal([]byte(msg), &doc)
	if err != nil {
		logger.GetLogger().Errorf("Cannot unmarshal jobproject doc: %v", err.Error())
		return err
	}
	err = c.svc.Upsert(doc.HoldingCode, doc.GuidFixed, doc)
	if err != nil {
		logger.GetLogger().Errorf("Cannot upsert jobproject pg: %v", err.Error())
		return err
	}
	return nil
}

func (c *JobProjectConsumer) ConsumeOnBulkCreateOrUpdate(ctx microservice.IContext) error {
	msg := ctx.ReadInput()
	docs := []models.JobProjectDoc{}
	err := json.Unmarshal([]byte(msg), &docs)
	if err != nil {
		logger.GetLogger().Errorf("Cannot unmarshal jobproject bulk doc: %v", err.Error())
		return err
	}
	for _, doc := range docs {
		err = c.svc.Upsert(doc.HoldingCode, doc.GuidFixed, doc)
		if err != nil {
			logger.GetLogger().Errorf("Cannot upsert jobproject pg: %v", err.Error())
			return err
		}
	}
	return nil
}

func (c *JobProjectConsumer) ConsumeOnDelete(ctx microservice.IContext) error {
	msg := ctx.ReadInput()
	doc := models.JobProjectDoc{}
	err := json.Unmarshal([]byte(msg), &doc)
	if err != nil {
		logger.GetLogger().Errorf("Cannot unmarshal jobproject doc: %v", err.Error())
		return err
	}
	err = c.svc.Delete(doc.HoldingCode, doc.GuidFixed)
	if err != nil {
		logger.GetLogger().Errorf("Cannot delete jobproject pg: %v", err.Error())
		return err
	}
	return nil
}

func (c *JobProjectConsumer) ConsumeOnBulkDelete(ctx microservice.IContext) error {
	msg := ctx.ReadInput()
	docs := []models.JobProjectDoc{}
	err := json.Unmarshal([]byte(msg), &docs)
	if err != nil {
		logger.GetLogger().Errorf("Cannot unmarshal jobproject bulk doc: %v", err.Error())
		return err
	}
	for _, doc := range docs {
		err = c.svc.Delete(doc.HoldingCode, doc.GuidFixed)
		if err != nil {
			logger.GetLogger().Errorf("Cannot delete jobproject pg: %v", err.Error())
			return err
		}
	}
	return nil
}
