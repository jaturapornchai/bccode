package rfq

import (
	pkgConfig "smlcloudplatform/internal/config"
	"smlcloudplatform/internal/logger"
	"smlcloudplatform/internal/transaction/models"
	rfqConfig "smlcloudplatform/internal/transaction/rfq/config"
	"smlcloudplatform/internal/transaction/transactionconsumer/services"
	"smlcloudplatform/internal/transaction/transactionconsumer/usecases"
	"smlcloudplatform/pkg/microservice"
	"time"

	trans_models "smlcloudplatform/internal/transaction/models"
)

type RFQTransactionConsumer struct {
	ms        *microservice.Microservice
	cfg       pkgConfig.IConfig
	svc       IRFQTransactionConsumerService
	txnPhaser usecases.ITransactionPhaser[models.RFQTransactionPG]
}

func NewRFQTransactionConsumer(
	ms *microservice.Microservice,
	cfg pkgConfig.IConfig,
	svc IRFQTransactionConsumerService,
	txnPhaser usecases.ITransactionPhaser[models.RFQTransactionPG],
) services.ITransactionDocConsumer {
	return &RFQTransactionConsumer{
		ms:        ms,
		cfg:       cfg,
		svc:       svc,
		txnPhaser: txnPhaser,
	}
}

func InitRFQTransactionConsumer(ms *microservice.Microservice, cfg pkgConfig.IConfig) services.ITransactionDocConsumer {
	persister := ms.Persister(cfg.PersisterConfig())
	repo := NewRFQTransactionRepository(persister)
	consumerService := NewRFQTransactionService(repo)
	phaser := RFQTransactionPhaser{}
	consumer := NewRFQTransactionConsumer(ms, cfg, consumerService, phaser)
	return consumer
}

func (c *RFQTransactionConsumer) RegisterConsumer(ms *microservice.Microservice) {
	trxConsumerGroup := pkgConfig.GetEnv("TRANSACTION_CONSUMER_GROUP", "transaction-consumer-group-01")
	mq := microservice.NewMQ(c.cfg.MQConfig(), ms.Logger)
	mqConfig := rfqConfig.RFQMessageQueueConfig{}

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

func (c *RFQTransactionConsumer) ConsumeOnCreateOrUpdate(ctx microservice.IContext) error {
	msg := ctx.ReadInput()
	transaction, err := c.txnPhaser.PhaseSingleDoc(msg)
	if err != nil {
		logger.GetLogger().Errorf("Cannot phase RFQ doc: %v", err.Error())
		return err
	}
	err = c.svc.Upsert(transaction.ShopID, transaction.DocNo, *transaction)
	if err != nil {
		logger.GetLogger().Errorf("Cannot insert RFQ transaction pg: %v", err.Error())
		return err
	}
	return nil
}

func (c *RFQTransactionConsumer) ConsumeOnBulkCreateOrUpdate(ctx microservice.IContext) error {
	msg := ctx.ReadInput()
	transactions, err := c.txnPhaser.PhaseMultipleDoc(msg)
	if err != nil {
		c.ms.Logger.Errorf("Cannot phase RFQ doc: %v", err.Error())
		return err
	}
	for _, transaction := range *transactions {
		err = c.svc.Upsert(transaction.ShopID, transaction.DocNo, transaction)
		if err != nil {
			logger.GetLogger().Errorf("Cannot insert RFQ transaction pg: %v", err.Error())
			return err
		}
	}
	return nil
}

func (c *RFQTransactionConsumer) ConsumeOnDelete(ctx microservice.IContext) error {
	msg := ctx.ReadInput()
	transaction, err := c.txnPhaser.PhaseSingleDoc(msg)
	if err != nil {
		logger.GetLogger().Errorf("Cannot phase RFQ doc: %v", err.Error())
		return err
	}
	err = c.svc.Delete(transaction.ShopID, transaction.DocNo)
	if err != nil {
		logger.GetLogger().Errorf("Cannot delete RFQ transaction pg: %v", err.Error())
		return err
	}
	return nil
}

func (c *RFQTransactionConsumer) ConsumeOnBulkDelete(ctx microservice.IContext) error {
	msg := ctx.ReadInput()
	transactions, err := c.txnPhaser.PhaseMultipleDoc(msg)
	if err != nil {
		c.ms.Logger.Errorf("Cannot phase RFQ doc: %v", err.Error())
		return err
	}
	for _, transaction := range *transactions {
		err = c.svc.Delete(transaction.ShopID, transaction.DocNo)
		if err != nil {
			logger.GetLogger().Errorf("Cannot delete RFQ transaction pg: %v", err.Error())
			return err
		}
	}
	return nil
}

func MigrationDatabase(ms *microservice.Microservice, cfg pkgConfig.IConfig) error {
	pst := ms.Persister(cfg.PersisterConfig())
	pst.AutoMigrate(
		trans_models.RFQTransactionPG{},
		trans_models.RFQDetailTransactionPG{},
	)
	return nil
}
