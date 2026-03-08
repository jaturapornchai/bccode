package creditorpayment

import (
	"smlcloudplatform/internal/config"
	pkgConfig "smlcloudplatform/internal/config"
	models "smlcloudplatform/internal/transaction/models"
	creditorPaymentConfig "smlcloudplatform/internal/transaction/pay/config"
	"smlcloudplatform/internal/transaction/transactionconsumer/services"
	"smlcloudplatform/internal/transaction/transactionconsumer/usecases"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type CreditorPaymentTransactionConsumer struct {
	ms                *microservice.Microservice
	cfg               config.IConfig
	svc               ICreditorPaymentTransactionConsumerService
	transactionPhaser usecases.ITransactionPhaser[models.CreditorPaymentTransactionPG]
}

func InitCreditorPaymentTransactionConsumer(ms *microservice.Microservice, cfg config.IConfig) services.ITransactionDocConsumer {

	persister := ms.Persister(cfg.PersisterConfig())
	creditorPaymentTransactionConsumerService := NewCreditorPaymentTransactionPGRepository(persister)
	creditPaymentConsumerService := NewCreditorPaymentTransactionConsumerService(creditorPaymentTransactionConsumerService)

	creditorPaymentTransactionConsumer := NewCreditorPaymentTransactionConsumer(
		ms,
		cfg,
		creditPaymentConsumerService,
	)
	return creditorPaymentTransactionConsumer
}

func NewCreditorPaymentTransactionConsumer(
	ms *microservice.Microservice,
	cfg config.IConfig,
	svc ICreditorPaymentTransactionConsumerService,
) services.ITransactionDocConsumer {

	creditorPaymentTransactionPhaser := CreditorPaymentTransactionPhaser{}

	return &CreditorPaymentTransactionConsumer{
		ms:                ms,
		cfg:               cfg,
		svc:               svc,
		transactionPhaser: creditorPaymentTransactionPhaser,
	}
}

func (c *CreditorPaymentTransactionConsumer) RegisterConsumer(ms *microservice.Microservice) {

	trxConsumerGroup := pkgConfig.GetEnv("TRANSACTION_CONSUMER_GROUP", "transaction-consumer-group-01")
	mq := microservice.NewMQ(c.cfg.MQConfig(), ms.Logger)

	purchaseKafkaConfig := creditorPaymentConfig.CreditorPaymentMessageQueueConfig{}

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

func (c *CreditorPaymentTransactionConsumer) ConsumeOnCreateOrUpdate(ctx microservice.IContext) error {

	input := ctx.ReadInput()
	transaction, err := c.transactionPhaser.PhaseSingleDoc(input)
	if err != nil {
		return err
	}

	err = c.svc.Upsert(transaction.ShopID, transaction.DocNo, *transaction)
	if err != nil {
		return err
	}

	// produce to creditor process topic

	return nil
}

func (c *CreditorPaymentTransactionConsumer) ConsumeOnBulkCreateOrUpdate(ctx microservice.IContext) error {

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

func (c *CreditorPaymentTransactionConsumer) ConsumeOnDelete(ctx microservice.IContext) error {

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
func (c *CreditorPaymentTransactionConsumer) ConsumeOnBulkDelete(ctx microservice.IContext) error {

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

func MigrationDatabase(ms *microservice.Microservice, cfg config.IConfig) error {
	pst := ms.Persister(cfg.PersisterConfig())
	pst.AutoMigrate(
		models.CreditorPaymentTransactionPG{},
		models.CreditorPaymentTransactionDetailPG{},
	)
	return nil
}
