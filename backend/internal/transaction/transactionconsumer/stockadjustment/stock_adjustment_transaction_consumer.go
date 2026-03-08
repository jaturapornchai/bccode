package stockadjustment

import (
	pkgConfig "smlcloudplatform/internal/config"
	"smlcloudplatform/internal/logger"
	"smlcloudplatform/internal/transaction/models"
	stockadjustmentproductconfig "smlcloudplatform/internal/transaction/stockadjustment/config"
	"smlcloudplatform/internal/transaction/transactionconsumer/services"
	"smlcloudplatform/internal/transaction/transactionconsumer/stocktransaction"
	"smlcloudplatform/internal/transaction/transactionconsumer/usecases"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type StockAdjustmentTransactionConsumer struct {
	ms                   *microservice.Microservice
	cfg                  pkgConfig.IConfig
	svc                  IStockAdjustmentTransactionConsumerService
	transactionPhaser    usecases.ITransactionPhaser[models.StockAdjustmentTransactionPG]
	stockPhaser          usecases.IStockTransactionPhaser[models.StockAdjustmentTransactionPG]
	stockConsumerService stocktransaction.IStockTransactionConsumerService
}

func InitStockAdjustmentTransactionConsumer(ms *microservice.Microservice, cfg pkgConfig.IConfig) services.ITransactionDocConsumer {
	persister := ms.Persister(cfg.PersisterConfig())
	producer := ms.Producer(cfg.MQConfig())

	repo := NewStockAdjustmentTransactionPGRepository(persister)
	stockReceiveProductConsumerService := NewStockAdjustmentTransactionConsumerService(repo)
	stockService := stocktransaction.NewStockTransactionConsumerService(persister, producer)

	stockReceiveTransactionConsumer := NewStockAdjustmentTransactionConsumer(
		ms,
		cfg,
		stockReceiveProductConsumerService,
		stockService,
	)
	return stockReceiveTransactionConsumer
}

func NewStockAdjustmentTransactionConsumer(
	ms *microservice.Microservice,
	cfg pkgConfig.IConfig,
	svc IStockAdjustmentTransactionConsumerService,
	stockConsumerService stocktransaction.IStockTransactionConsumerService,
) services.ITransactionDocConsumer {

	stockReceiveTransactionPhaser := StockAdjustmentTransactionPhaser{}
	stockReceiveStockPhaser := StockAdjustmentStockPhaser{}

	return &StockAdjustmentTransactionConsumer{
		ms:                   ms,
		cfg:                  cfg,
		svc:                  svc,
		transactionPhaser:    stockReceiveTransactionPhaser,
		stockPhaser:          stockReceiveStockPhaser,
		stockConsumerService: stockConsumerService,
	}
}

func consumerPanicHandler() {
	r := recover()
	if r != nil {
		logger.GetLogger().Errorf("Consumer panic: %v", r)
	}
}

func (c *StockAdjustmentTransactionConsumer) ConsumeOnCreateOrUpdate(ctx microservice.IContext) error {

	defer consumerPanicHandler()
	input := ctx.ReadInput()
	transaction, err := c.transactionPhaser.PhaseSingleDoc(input)
	if err != nil {
		logger.GetLogger().Errorf("Error phasing transaction: %v", err)
		return err
	}

	err = c.svc.Upsert(transaction.ShopID, transaction.DocNo, *transaction)
	if err != nil {
		logger.GetLogger().Errorf("Error upserting transaction: %v", err)
	}

	stockTransaction, err := c.stockPhaser.PhaseSingleDoc(*transaction)
	if err != nil {
		logger.GetLogger().Errorf("Error phasing stock transaction: %v", err)
		return err
	}

	err = c.stockConsumerService.Upsert(stockTransaction.ShopID, stockTransaction.DocNo, *stockTransaction)
	if err != nil {
		logger.GetLogger().Errorf("Error upserting stock transaction: %v", err)
	}

	return nil
}

func (c *StockAdjustmentTransactionConsumer) ConsumeOnBulkCreateOrUpdate(ctx microservice.IContext) error {

	defer consumerPanicHandler()
	input := ctx.ReadInput()
	transactions, err := c.transactionPhaser.PhaseMultipleDoc(input)
	if err != nil {
		logger.GetLogger().Errorf("Error phasing transactions: %v", err)
		return err
	}

	for _, transaction := range *transactions {

		err = c.svc.Upsert(transaction.ShopID, transaction.DocNo, transaction)
		if err != nil {
			logger.GetLogger().Errorf("Error upserting transaction: %v", err)
		}

		stockTransaction, err := c.stockPhaser.PhaseSingleDoc(transaction)
		if err != nil {
			logger.GetLogger().Errorf("Error phasing stock transaction: %v", err)
			return err
		}

		err = c.stockConsumerService.Upsert(stockTransaction.ShopID, stockTransaction.DocNo, *stockTransaction)
		if err != nil {
			logger.GetLogger().Errorf("Error upserting stock transaction: %v", err)
		}
	}

	return nil
}

func (c *StockAdjustmentTransactionConsumer) ConsumeOnDelete(ctx microservice.IContext) error {

	defer consumerPanicHandler()
	input := ctx.ReadInput()
	transaction, err := c.transactionPhaser.PhaseSingleDoc(input)
	if err != nil {
		logger.GetLogger().Errorf("Error phasing transaction: %v", err)
	}

	err = c.svc.Delete(transaction.ShopID, transaction.DocNo)
	if err != nil {
		logger.GetLogger().Errorf("Error deleting transaction: %v", err)
	}

	err = c.stockConsumerService.Delete(transaction.ShopID, transaction.DocNo)
	if err != nil {
		logger.GetLogger().Errorf("Error deleting stock transaction: %v", err)
		return err
	}

	return nil
}

func (c *StockAdjustmentTransactionConsumer) ConsumeOnBulkDelete(ctx microservice.IContext) error {

	defer consumerPanicHandler()
	input := ctx.ReadInput()

	transactions, err := c.transactionPhaser.PhaseMultipleDoc(input)
	if err != nil {
		logger.GetLogger().Errorf("Error phasing transactions: %v", err)
		return err
	}

	for _, transaction := range *transactions {

		err = c.svc.Delete(transaction.ShopID, transaction.DocNo)
		if err != nil {
			logger.GetLogger().Errorf("Error deleting transaction: %v", err)
		}

		err = c.stockConsumerService.Delete(transaction.ShopID, transaction.DocNo)
		if err != nil {
			logger.GetLogger().Errorf("Error deleting stock transaction: %v", err)
			return err
		}
	}
	return nil
}

func (c *StockAdjustmentTransactionConsumer) RegisterConsumer(ms *microservice.Microservice) {

	trxConsumerGroup := pkgConfig.GetEnv("TRANSACTION_CONSUMER_GROUP", "transaction-consumer-group-01")
	mq := microservice.NewMQ(c.cfg.MQConfig(), ms.Logger)
	stockProductReceiveKafkaConfig := stockadjustmentproductconfig.StockAdjustmentMessageQueueConfig{}

	mq.CreateTopicR(stockProductReceiveKafkaConfig.TopicCreated(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(stockProductReceiveKafkaConfig.TopicUpdated(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(stockProductReceiveKafkaConfig.TopicDeleted(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(stockProductReceiveKafkaConfig.TopicBulkCreated(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(stockProductReceiveKafkaConfig.TopicBulkUpdated(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(stockProductReceiveKafkaConfig.TopicBulkDeleted(), 5, 1, time.Hour*24*7)

	ms.Consume(c.cfg.MQConfig().URI(), stockProductReceiveKafkaConfig.TopicCreated(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnCreateOrUpdate)
	ms.Consume(c.cfg.MQConfig().URI(), stockProductReceiveKafkaConfig.TopicUpdated(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnCreateOrUpdate)
	ms.Consume(c.cfg.MQConfig().URI(), stockProductReceiveKafkaConfig.TopicDeleted(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnDelete)
	ms.Consume(c.cfg.MQConfig().URI(), stockProductReceiveKafkaConfig.TopicBulkCreated(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnBulkCreateOrUpdate)
	ms.Consume(c.cfg.MQConfig().URI(), stockProductReceiveKafkaConfig.TopicBulkUpdated(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnBulkCreateOrUpdate)
	ms.Consume(c.cfg.MQConfig().URI(), stockProductReceiveKafkaConfig.TopicBulkDeleted(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnBulkDelete)

}

func MigrationDatabase(ms *microservice.Microservice, cfg pkgConfig.IConfig) error {
	pst := ms.Persister(cfg.PersisterConfig())
	pst.AutoMigrate(
		models.StockAdjustmentTransactionPG{},
		models.StockAdjustmentTransactionDetailPG{},
	)
	return nil
}
