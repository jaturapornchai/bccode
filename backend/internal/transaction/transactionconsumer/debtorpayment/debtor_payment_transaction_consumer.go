package debtorpayment

import (
	"smlcloudplatform/internal/config"
	models "smlcloudplatform/internal/transaction/models"
	debtorPaymentConfig "smlcloudplatform/internal/transaction/paid/config"
	"smlcloudplatform/internal/transaction/transactionconsumer/services"
	"smlcloudplatform/internal/transaction/transactionconsumer/usecases"
	"smlcloudplatform/pkg/microservice"
	"time"

	pkgConfig "smlcloudplatform/internal/config"
)

type DebtorPaymentTransactionConsumer struct {
	ms                *microservice.Microservice
	cfg               config.IConfig
	svc               IDebtorPaymentConsumerService
	transactionPhaser usecases.ITransactionPhaser[models.DebtorPaymentTransactionPG]
}

func InitDebtorPaymentTransactionConsumer(ms *microservice.Microservice, cfg config.IConfig) services.ITransactionDocConsumer {

	repo := NewDebtorPaymentTransactionPGRepository(ms.Persister(cfg.PersisterConfig()))
	debtorPaymentConsumerService := NewDebtorPaymentConsumerService(repo)

	debtorPaymentTransactionConsumer := NewDebtorPaymentTransactionConsumer(ms, cfg, debtorPaymentConsumerService)
	return debtorPaymentTransactionConsumer
}

func NewDebtorPaymentTransactionConsumer(
	ms *microservice.Microservice,
	cfg config.IConfig,
	svc IDebtorPaymentConsumerService,
) services.ITransactionDocConsumer {
	phaser := DebtorPaymentTransactionPhaser{}

	return &DebtorPaymentTransactionConsumer{
		ms:                ms,
		cfg:               cfg,
		svc:               svc,
		transactionPhaser: phaser,
	}
}

func (c *DebtorPaymentTransactionConsumer) RegisterConsumer(ms *microservice.Microservice) {

	trxConsumerGroup := pkgConfig.GetEnv("TRANSACTION_CONSUMER_GROUP", "transaction-consumer-group-01")
	mq := microservice.NewMQ(c.cfg.MQConfig(), ms.Logger)

	purchaseKafkaConfig := debtorPaymentConfig.DebtorPaymentMessageQueueConfig{}

	mq.CreateTopicR(purchaseKafkaConfig.TopicCreated(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(purchaseKafkaConfig.TopicUpdated(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(purchaseKafkaConfig.TopicDeleted(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(purchaseKafkaConfig.TopicBulkCreated(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(purchaseKafkaConfig.TopicBulkUpdated(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(purchaseKafkaConfig.TopicBulkDeleted(), 5, 1, time.Hour*24*7)

	ms.Consume(c.cfg.MQConfig().URI(), purchaseKafkaConfig.TopicCreated(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnCreateOrUpdate)
	ms.Consume(c.cfg.MQConfig().URI(), purchaseKafkaConfig.TopicUpdated(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnCreateOrUpdate)
	ms.Consume(c.cfg.MQConfig().URI(), purchaseKafkaConfig.TopicDeleted(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnDelete)
	ms.Consume(c.cfg.MQConfig().URI(), purchaseKafkaConfig.TopicBulkCreated(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnBulkCreateOrUpdate)
	ms.Consume(c.cfg.MQConfig().URI(), purchaseKafkaConfig.TopicBulkUpdated(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnBulkCreateOrUpdate)
	ms.Consume(c.cfg.MQConfig().URI(), purchaseKafkaConfig.TopicBulkDeleted(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnBulkDelete)

}

func (c *DebtorPaymentTransactionConsumer) ConsumeOnCreateOrUpdate(ctx microservice.IContext) error {

	input := ctx.ReadInput()
	transaction, err := c.transactionPhaser.PhaseSingleDoc(input)
	if err != nil {
		return err
	}

	err = c.svc.Upsert(transaction.ShopID, transaction.DocNo, *transaction)
	if err != nil {
		return err
	}

	return nil
}

func (c *DebtorPaymentTransactionConsumer) ConsumeOnDelete(ctx microservice.IContext) error {

	input := ctx.ReadInput()
	transaction, err := c.transactionPhaser.PhaseSingleDoc(input)
	if err != nil {
		return err
	}

	err = c.svc.Delete(transaction.ShopID, transaction.DocNo)
	if err != nil {
		return err
	}

	return nil
}

func (c *DebtorPaymentTransactionConsumer) ConsumeOnBulkCreateOrUpdate(ctx microservice.IContext) error {
	input := ctx.ReadInput()
	transactions, err := c.transactionPhaser.PhaseMultipleDoc(input)
	if err != nil {
		return err
	}

	for _, transaction := range *transactions {
		err = c.svc.Upsert(transaction.ShopID, transaction.DocNo, transaction)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *DebtorPaymentTransactionConsumer) ConsumeOnBulkDelete(ctx microservice.IContext) error {
	input := ctx.ReadInput()
	transactions, err := c.transactionPhaser.PhaseMultipleDoc(input)
	if err != nil {
		return err
	}

	for _, transaction := range *transactions {
		err = c.svc.Delete(transaction.ShopID, transaction.DocNo)
		if err != nil {
			return err
		}
	}
	return nil
}

func MigrationDatabase(ms *microservice.Microservice, cfg config.IConfig) {
	pst := ms.Persister(cfg.PersisterConfig())
	pst.AutoMigrate(
		models.DebtorPaymentTransactionPG{},
		models.DebtorPaymentTransactionDetailPG{},
	)
}
