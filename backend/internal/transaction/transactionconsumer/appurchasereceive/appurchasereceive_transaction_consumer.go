package appurchasereceive

import (
	pkgConfig "smlcloudplatform/internal/config"
	"smlcloudplatform/internal/logger"
	accrualReceiveConfig "smlcloudplatform/internal/transaction/accrualreceive/config"
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/creditortransaction"
	"smlcloudplatform/internal/transaction/transactionconsumer/usecases"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type APPurchaseReceiveTransactionConsumer struct {
	ms                      *microservice.Microservice
	cfg                     pkgConfig.IConfig
	svc                     IAPPurchaseReceiveTransactionConsumerService
	txnPhaser               usecases.ITransactionPhaser[models.APPurchaseReceivePG]
	creditorPhaser          usecases.ICreditorTransactionPhaser[models.APPurchaseReceivePG]
	creditorConsumerService creditortransaction.ICreditorTransactionConsumerService
}

func NewAPPurchaseReceiveTransactionConsumer(
	ms *microservice.Microservice,
	cfg pkgConfig.IConfig,
	svc IAPPurchaseReceiveTransactionConsumerService,
	creditorConsumerService creditortransaction.ICreditorTransactionConsumerService,
) *APPurchaseReceiveTransactionConsumer {

	txnPhaser := APPurchaseReceiveTransactionPhaser{}
	creditorPhaser := APPurchaseReceiveCreditorTransactionPhaser{}

	return &APPurchaseReceiveTransactionConsumer{
		ms:             ms,
		cfg:            cfg,
		svc:            svc,
		txnPhaser:      txnPhaser,
		creditorPhaser: creditorPhaser,
	}
}

func InitAPPurchaseReceiveTransactionConsumer(
	ms *microservice.Microservice,
	cfg pkgConfig.IConfig,
) *APPurchaseReceiveTransactionConsumer {

	persister := ms.Persister(cfg.PersisterConfig())
	producer := ms.Producer(cfg.MQConfig())

	creditorService := creditortransaction.NewCreditorTransactionConsumerService(persister, producer)

	accrualReceiveConsumerService := NewAPPurchaseReceiveTransactionConsumerService(NewAPPurchaseReceiveTransactionPostgresRepository(persister))
	consumer := NewAPPurchaseReceiveTransactionConsumer(
		ms,
		cfg,
		accrualReceiveConsumerService,
		creditorService,
	)
	return consumer
}

func (c *APPurchaseReceiveTransactionConsumer) RegisterConsumer(ms *microservice.Microservice) {

	trxConsumerGroup := pkgConfig.GetEnv("TRANSACTION_CONSUMER_GROUP", "transaction-consumer-group-01")
	mq := microservice.NewMQ(c.cfg.MQConfig(), ms.Logger)

	accrualReceiveKafkaConfig := accrualReceiveConfig.AccrualreceiveMessageQueueConfig{}

	mq.CreateTopicR(accrualReceiveKafkaConfig.TopicCreated(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(accrualReceiveKafkaConfig.TopicUpdated(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(accrualReceiveKafkaConfig.TopicDeleted(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(accrualReceiveKafkaConfig.TopicBulkCreated(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(accrualReceiveKafkaConfig.TopicBulkUpdated(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(accrualReceiveKafkaConfig.TopicBulkDeleted(), 5, 1, time.Hour*24*7)

	ms.Consume(c.cfg.MQConfig().URI(), accrualReceiveKafkaConfig.TopicCreated(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnCreateOrUpdate)
	ms.Consume(c.cfg.MQConfig().URI(), accrualReceiveKafkaConfig.TopicUpdated(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnCreateOrUpdate)
	ms.Consume(c.cfg.MQConfig().URI(), accrualReceiveKafkaConfig.TopicDeleted(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnDelete)
	ms.Consume(c.cfg.MQConfig().URI(), accrualReceiveKafkaConfig.TopicBulkCreated(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnBulkCreateOrUpdate)
	ms.Consume(c.cfg.MQConfig().URI(), accrualReceiveKafkaConfig.TopicBulkUpdated(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnBulkCreateOrUpdate)
	ms.Consume(c.cfg.MQConfig().URI(), accrualReceiveKafkaConfig.TopicBulkDeleted(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnBulkDelete)

}

func (c *APPurchaseReceiveTransactionConsumer) ConsumeOnCreateOrUpdate(ctx microservice.IContext) error {

	msg := ctx.ReadInput()

	transaction, err := c.txnPhaser.PhaseSingleDoc(msg)
	if err != nil {
		logger.GetLogger().Errorf("Cannot Phase PurchaseDoc to Purchase Transaction : %v", err.Error())
		return err
	}

	err = c.svc.Upsert(transaction.ShopID, transaction.DocNo, *transaction)
	if err != nil {
		logger.GetLogger().Errorf("Cannot Insert Purchase Transaction : %v", err.Error())
		return err
	}

	creditor, err := c.creditorPhaser.PhaseSingleDoc(*transaction)
	if err != nil {
		logger.GetLogger().Errorf("Cannot Phase PurchaseDoc to Creditor Transaction : %v", err.Error())
		return err
	}
	err = c.creditorConsumerService.Upsert(transaction.ShopID, transaction.DocNo, *creditor)
	if err != nil {
		logger.GetLogger().Errorf("Cannot Insert CreditorTransaction : %v", err.Error())
		return err
	}

	return nil
}

func (c *APPurchaseReceiveTransactionConsumer) ConsumeOnDelete(ctx microservice.IContext) error {

	msg := ctx.ReadInput()

	transaction, err := c.txnPhaser.PhaseSingleDoc(msg)
	if err != nil {
		c.ms.Logger.Errorf("Cannot Phase PurchaseDoc to Purchase Transaction : %v", err.Error())
		return err
	}

	err = c.svc.Delete(transaction.ShopID, transaction.DocNo)
	if err != nil {
		c.ms.Logger.Errorf("Cannot Insert Purchase Transaction : %v", err.Error())
		return err
	}

	// delete creditor transaction
	err = c.creditorConsumerService.Delete(transaction.ShopID, transaction.DocNo)
	if err != nil {
		c.ms.Logger.Errorf("Cannot Delete Creditor Transaction : %v", err.Error())
		return err
	}

	return nil
}

func (c *APPurchaseReceiveTransactionConsumer) ConsumeOnBulkCreateOrUpdate(ctx microservice.IContext) error {

	msg := ctx.ReadInput()
	transactions, err := c.txnPhaser.PhaseMultipleDoc(msg)
	if err != nil {
		c.ms.Logger.Errorf("Cannot Phase PurchaseDoc to Purchase Transaction : %v", err.Error())
		return err
	}

	for _, transaction := range *transactions {
		err = c.svc.Upsert(transaction.ShopID, transaction.DocNo, transaction)
		if err != nil {
			c.ms.Logger.Errorf("Cannot Insert Purchase Transaction : %v", err.Error())
			return err
		}

		creditor, err := c.creditorPhaser.PhaseSingleDoc(transaction)
		if err != nil {
			logger.GetLogger().Errorf("Cannot Phase PurchaseDoc to Creditor Transaction : %v", err.Error())
			return err
		}
		err = c.creditorConsumerService.Upsert(transaction.ShopID, transaction.DocNo, *creditor)
		if err != nil {
			logger.GetLogger().Errorf("Cannot Insert CreditorTransaction : %v", err.Error())
			return err
		}
	}
	return nil
}

func (c *APPurchaseReceiveTransactionConsumer) ConsumeOnBulkDelete(ctx microservice.IContext) error {

	msg := ctx.ReadInput()

	transactions, err := c.txnPhaser.PhaseMultipleDoc(msg)
	if err != nil {
		c.ms.Logger.Errorf("Cannot Phase PurchaseDoc to Purchase Transaction : %v", err.Error())
		return err
	}

	for _, transaction := range *transactions {
		err = c.svc.Delete(transaction.ShopID, transaction.DocNo)
		if err != nil {
			c.ms.Logger.Errorf("Cannot Insert StockTransaction : %v", err.Error())
			return err
		}

		err = c.creditorConsumerService.Delete(transaction.ShopID, transaction.DocNo)
		if err != nil {
			c.ms.Logger.Errorf("Cannot Delete Creditor Transaction : %v", err.Error())
			return err
		}
	}
	return nil
}

func MigrationDatabase(ms *microservice.Microservice, cfg pkgConfig.IConfig) error {
	pst := ms.Persister(cfg.PersisterConfig())
	pst.AutoMigrate(
		models.APPurchaseReceivePG{},
		models.APPurchaseReceiveDetailPG{},
	)
	return nil
}
