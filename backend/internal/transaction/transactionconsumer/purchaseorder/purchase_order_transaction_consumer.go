package purchaseorder

import (
	pkgConfig "smlcloudplatform/internal/config"
	"smlcloudplatform/internal/logger"
	"smlcloudplatform/internal/transaction/models"
	purchaserOrderConfig "smlcloudplatform/internal/transaction/purchaseorder/config"
	"smlcloudplatform/internal/transaction/transactionconsumer/services"
	"smlcloudplatform/internal/transaction/transactionconsumer/usecases"
	"smlcloudplatform/pkg/microservice"
	"time"

	trans_models "smlcloudplatform/internal/transaction/models"
)

type PurchaseOrderTransactionConsumer struct {
	ms        *microservice.Microservice
	cfg       pkgConfig.IConfig
	svc       IPurchaseOrderTransactionConsumerService
	txnPhaser usecases.ITransactionPhaser[models.PurchaseOrderTransactionPG]
}

func NewPurchaseOrderTransactionConsumer(
	ms *microservice.Microservice,
	cfg pkgConfig.IConfig,
	svc IPurchaseOrderTransactionConsumerService,
	txnPhaser usecases.ITransactionPhaser[models.PurchaseOrderTransactionPG],
) services.ITransactionDocConsumer {

	return &PurchaseOrderTransactionConsumer{
		ms:        ms,
		cfg:       cfg,
		svc:       svc,
		txnPhaser: txnPhaser,
	}
}

func InitPurchaseOrderTransactionConsumer(ms *microservice.Microservice, cfg pkgConfig.IConfig) services.ITransactionDocConsumer {
	persister := ms.Persister(cfg.PersisterConfig())

	repo := NewPurchaseOrderTransactionRepository(persister)
	purchaseOrderConsumerService := NewPurchaseOrderTransactionService(repo)
	purchaseOrderPhaser := PurchaseOrderTransactionPhaser{}

	purchaseOrderTransactionConsumer := NewPurchaseOrderTransactionConsumer(
		ms,
		cfg,
		purchaseOrderConsumerService,
		purchaseOrderPhaser,
	)
	return purchaseOrderTransactionConsumer
}

func (c *PurchaseOrderTransactionConsumer) RegisterConsumer(ms *microservice.Microservice) {
	trxConsumerGroup := pkgConfig.GetEnv("TRANSACTION_CONSUMER_GROUP", "transaction-consumer-group-01")
	mq := microservice.NewMQ(c.cfg.MQConfig(), ms.Logger)

	mqConfig := purchaserOrderConfig.PurchaseOrderMessageQueueConfig{}

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

func (c *PurchaseOrderTransactionConsumer) ConsumeOnCreateOrUpdate(ctx microservice.IContext) error {

	msg := ctx.ReadInput()

	transaction, err := c.txnPhaser.PhaseSingleDoc(msg)
	if err != nil {
		logger.GetLogger().Errorf("Cannot phase purchase order doc : %v", err.Error())
		return err
	}

	err = c.svc.Upsert(transaction.HoldingCode, transaction.DocNo, *transaction)
	if err != nil {
		logger.GetLogger().Errorf("Cannot insert purchase order transaction pg: %v", err.Error())
		return err
	}

	return nil
}

func (c *PurchaseOrderTransactionConsumer) ConsumeOnBulkCreateOrUpdate(ctx microservice.IContext) error {

	msg := ctx.ReadInput()
	transactions, err := c.txnPhaser.PhaseMultipleDoc(msg)
	if err != nil {
		c.ms.Logger.Errorf("Cannot phase purchase order doc: %v", err.Error())
		return err
	}

	for _, transaction := range *transactions {

		err = c.svc.Upsert(transaction.HoldingCode, transaction.DocNo, transaction)
		if err != nil {
			logger.GetLogger().Errorf("Cannot insert purchase order transaction  pg: %v", err.Error())
			return err
		}
	}
	return nil
}

func (c *PurchaseOrderTransactionConsumer) ConsumeOnDelete(ctx microservice.IContext) error {

	msg := ctx.ReadInput()

	transaction, err := c.txnPhaser.PhaseSingleDoc(msg)
	if err != nil {
		logger.GetLogger().Errorf("Cannot phase purchase order doc : %v", err.Error())
		return err
	}

	err = c.svc.Delete(transaction.HoldingCode, transaction.DocNo)
	if err != nil {
		logger.GetLogger().Errorf("Cannot insert purchase order transaction pg: %v", err.Error())
		return err
	}

	return nil
}

func (c *PurchaseOrderTransactionConsumer) ConsumeOnBulkDelete(ctx microservice.IContext) error {
	msg := ctx.ReadInput()
	transactions, err := c.txnPhaser.PhaseMultipleDoc(msg)
	if err != nil {
		c.ms.Logger.Errorf("Cannot phase purchase order doc: %v", err.Error())
		return err
	}

	for _, transaction := range *transactions {

		err = c.svc.Delete(transaction.HoldingCode, transaction.DocNo)
		if err != nil {
			logger.GetLogger().Errorf("Cannot insert purchase order transaction  pg: %v", err.Error())
			return err
		}
	}
	return nil
}

func MigrationDatabase(ms *microservice.Microservice, cfg pkgConfig.IConfig) error {
	pst := ms.Persister(cfg.PersisterConfig())
	pst.AutoMigrate(
		trans_models.PurchaseOrderTransactionPG{},
		trans_models.PurchaseOrderDetailTransactionPG{},
	)
	return nil
}
