package purchaserequisition

import (
	pkgConfig "smlcloudplatform/internal/config"
	"smlcloudplatform/internal/logger"
	"smlcloudplatform/internal/transaction/models"
	prConfig "smlcloudplatform/internal/transaction/purchaserequisition/config"
	"smlcloudplatform/internal/transaction/transactionconsumer/services"
	"smlcloudplatform/internal/transaction/transactionconsumer/usecases"
	"smlcloudplatform/pkg/microservice"
	"time"

	trans_models "smlcloudplatform/internal/transaction/models"
)

type PurchaseRequisitionTransactionConsumer struct {
	ms        *microservice.Microservice
	cfg       pkgConfig.IConfig
	svc       IPurchaseRequisitionTransactionConsumerService
	txnPhaser usecases.ITransactionPhaser[models.PurchaseRequisitionTransactionPG]
}

func NewPurchaseRequisitionTransactionConsumer(
	ms *microservice.Microservice,
	cfg pkgConfig.IConfig,
	svc IPurchaseRequisitionTransactionConsumerService,
	txnPhaser usecases.ITransactionPhaser[models.PurchaseRequisitionTransactionPG],
) services.ITransactionDocConsumer {
	return &PurchaseRequisitionTransactionConsumer{
		ms:        ms,
		cfg:       cfg,
		svc:       svc,
		txnPhaser: txnPhaser,
	}
}

func InitPurchaseRequisitionTransactionConsumer(ms *microservice.Microservice, cfg pkgConfig.IConfig) services.ITransactionDocConsumer {
	persister := ms.Persister(cfg.PersisterConfig())
	repo := NewPurchaseRequisitionTransactionRepository(persister)
	consumerService := NewPurchaseRequisitionTransactionService(repo)
	phaser := PurchaseRequisitionTransactionPhaser{}
	consumer := NewPurchaseRequisitionTransactionConsumer(ms, cfg, consumerService, phaser)
	return consumer
}

func (c *PurchaseRequisitionTransactionConsumer) RegisterConsumer(ms *microservice.Microservice) {
	trxConsumerGroup := pkgConfig.GetEnv("TRANSACTION_CONSUMER_GROUP", "transaction-consumer-group-01")
	mq := microservice.NewMQ(c.cfg.MQConfig(), ms.Logger)
	mqConfig := prConfig.PurchaseRequisitionMessageQueueConfig{}

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

func (c *PurchaseRequisitionTransactionConsumer) ConsumeOnCreateOrUpdate(ctx microservice.IContext) error {
	msg := ctx.ReadInput()
	transaction, err := c.txnPhaser.PhaseSingleDoc(msg)
	if err != nil {
		logger.GetLogger().Errorf("Cannot phase purchase requisition doc: %v", err.Error())
		return err
	}
	err = c.svc.Upsert(transaction.ShopID, transaction.DocNo, *transaction)
	if err != nil {
		logger.GetLogger().Errorf("Cannot insert purchase requisition transaction pg: %v", err.Error())
		return err
	}
	return nil
}

func (c *PurchaseRequisitionTransactionConsumer) ConsumeOnBulkCreateOrUpdate(ctx microservice.IContext) error {
	msg := ctx.ReadInput()
	transactions, err := c.txnPhaser.PhaseMultipleDoc(msg)
	if err != nil {
		c.ms.Logger.Errorf("Cannot phase purchase requisition doc: %v", err.Error())
		return err
	}
	for _, transaction := range *transactions {
		err = c.svc.Upsert(transaction.ShopID, transaction.DocNo, transaction)
		if err != nil {
			logger.GetLogger().Errorf("Cannot insert purchase requisition transaction pg: %v", err.Error())
			return err
		}
	}
	return nil
}

func (c *PurchaseRequisitionTransactionConsumer) ConsumeOnDelete(ctx microservice.IContext) error {
	msg := ctx.ReadInput()
	transaction, err := c.txnPhaser.PhaseSingleDoc(msg)
	if err != nil {
		logger.GetLogger().Errorf("Cannot phase purchase requisition doc: %v", err.Error())
		return err
	}
	err = c.svc.Delete(transaction.ShopID, transaction.DocNo)
	if err != nil {
		logger.GetLogger().Errorf("Cannot delete purchase requisition transaction pg: %v", err.Error())
		return err
	}
	return nil
}

func (c *PurchaseRequisitionTransactionConsumer) ConsumeOnBulkDelete(ctx microservice.IContext) error {
	msg := ctx.ReadInput()
	transactions, err := c.txnPhaser.PhaseMultipleDoc(msg)
	if err != nil {
		c.ms.Logger.Errorf("Cannot phase purchase requisition doc: %v", err.Error())
		return err
	}
	for _, transaction := range *transactions {
		err = c.svc.Delete(transaction.ShopID, transaction.DocNo)
		if err != nil {
			logger.GetLogger().Errorf("Cannot delete purchase requisition transaction pg: %v", err.Error())
			return err
		}
	}
	return nil
}

func MigrationDatabase(ms *microservice.Microservice, cfg pkgConfig.IConfig) error {
	pst := ms.Persister(cfg.PersisterConfig())
	pst.AutoMigrate(
		trans_models.PurchaseRequisitionTransactionPG{},
		trans_models.PurchaseRequisitionDetailTransactionPG{},
	)
	return nil
}
