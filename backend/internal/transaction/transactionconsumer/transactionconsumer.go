package transactionconsumer

import (
	msConfig "smlcloudplatform/internal/config"
	pkgConfig "smlcloudplatform/internal/config"
	saleInvoiceReturnConfig "smlcloudplatform/internal/transaction/saleinvoicereturn/config"
	"smlcloudplatform/internal/transaction/transactionconsumer/creditortransaction"
	"smlcloudplatform/internal/transaction/transactionconsumer/debtortransaction"
	transaction_payment_consume "smlcloudplatform/internal/transaction/transactionconsumer/payment"
	"smlcloudplatform/internal/transaction/transactionconsumer/purchase"
	"smlcloudplatform/internal/transaction/transactionconsumer/purchasereturn"
	"smlcloudplatform/internal/transaction/transactionconsumer/saleinvoice"
	"smlcloudplatform/internal/transaction/transactionconsumer/saleinvoicereturn"
	"smlcloudplatform/internal/transaction/transactionconsumer/services"
	"smlcloudplatform/internal/transaction/transactionconsumer/stocktransaction"
	"smlcloudplatform/pkg/microservice"
	"time"

	productbom_repositories "smlcloudplatform/internal/product/bom/repositories"
	productbom_services "smlcloudplatform/internal/product/bom/services"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	saleinvoicebomprice_repositories "smlcloudplatform/internal/transaction/saleinvoicebomprice/repositories"
	saleinvoicebomprice_services "smlcloudplatform/internal/transaction/saleinvoicebomprice/services"
)

type ITransactionConsumer interface {
	RegisterConsumer(ms *microservice.Microservice)

	// ConsumeOnPurchaseDocCreatedOrUpdate(ctx microservice.IContext) error
	// ConsumeOnPurchaseDocDeleted(ctx microservice.IContext) error
	// ConsumeOnPurchaseDocBulkCreatedOrUpdate(ctx microservice.IContext) error

	// ConsumeOnSaleInvoiceDocCreatedOrUpdated(ctx microservice.IContext) error
	// ConsumeOnSaleInvoiceDocDeleted(ctx microservice.IContext) error

	// ConsumeOnSaleInvoiceReturnDocOrUpdated(ctx microservice.IContext) error
	// ConsumeOnSaleInvoiceReturnDocDeleted(ctx microservice.IContext) error
}

type TransactionConsumer struct {
	cfg                       pkgConfig.IConfig
	ms                        *microservice.Microservice
	purchaseConsumer          services.ITransactionDocConsumer
	purchaseReturnConsumer    services.ITransactionDocConsumer
	saleInvoiceConsumer       services.ITransactionDocConsumer
	saleInvoiceReturnConsumer services.ITransactionDocConsumer

	// svc                       services.ITransactionConsumerService
	// phaser                    usecases.ITransactionPhaser
}

func NewTransactionConsumer(ms *microservice.Microservice, cfg pkgConfig.IConfig) ITransactionConsumer {

	persisterMongo := ms.MongoPersister(cfg.MongoPersisterConfig())
	persister := ms.Persister(cfg.PersisterConfig())
	producer := ms.Producer(cfg.MQConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	stockService := stocktransaction.NewStockTransactionConsumerService(persister, producer)
	creditorService := creditortransaction.NewCreditorTransactionConsumerService(persister, producer)
	debtorService := debtortransaction.NewDebtorTransactionService(persister, producer)

	transPaymentConsumeUsecase := transaction_payment_consume.InitPayment(persister)

	purchaseConsumerService := purchase.NewPurchaseTransactionService(purchase.NewPurchaseTransactionPGRepository(persister))
	purchaseTransactionConsumer := purchase.NewPurchaseTransactionConsumer(ms, cfg, purchaseConsumerService, stockService, creditorService, transPaymentConsumeUsecase)

	purchaseReturnConsumerService := purchasereturn.NewPurchaseReturnTransactionService(purchasereturn.NewPurchaseReturnTransactionPGRepository(persister))
	purchaseReturnStockConsumer := purchasereturn.NewTransactionPurchaseReturnConsumer(ms, cfg, purchaseReturnConsumerService, stockService, creditorService, transPaymentConsumeUsecase)

	productBarcodeRepository := productbarcode_repositories.NewProductBarcodeRepository(persisterMongo, cache)
	saleinvoiceBomPriceRepository := saleinvoicebomprice_repositories.NewSaleInvoiceBomPriceRepository(persisterMongo)
	saleinvoiceBomPriceRepo := saleinvoicebomprice_repositories.NewSaleInvoiceBOMPriceMessageQueueRepository(producer)
	saleinvoiceBomPriceService := saleinvoicebomprice_services.NewSaleInvoiceBomPriceService(saleinvoiceBomPriceRepo, saleinvoiceBomPriceRepository)
	productBomRepository := productbom_repositories.NewBomRepository(persisterMongo)
	productBomMqRepository := productbom_repositories.NewBomMessageQueueRepository(producer)
	productBomService := productbom_services.NewBOMHttpService(productBomRepository, productBomMqRepository, productBarcodeRepository, saleinvoiceBomPriceService)

	saleInvoiceConsumerService := saleinvoice.NewSaleInvoiceTransactionConsumerService(saleinvoice.NewSaleInvoiceTransactionPGRepository(persister))
	saleInvoiceStockConsumer := saleinvoice.NewSaleInvoiceTransactionConsumer(ms, cfg, saleInvoiceConsumerService, stockService, debtorService, productBomService, transPaymentConsumeUsecase)

	saleInvoiceReturnConsumerService := saleinvoicereturn.NewSaleInvoiceReturnTransactionConsumerService(saleinvoicereturn.NewSaleInvoiceReturnTransactionPGRepository(persister))
	saleInvoiceReturnStockConsumer := saleinvoicereturn.NewSaleInvoiceReturnTransactionConsumer(ms, cfg, saleInvoiceReturnConsumerService, stockService, debtorService, transPaymentConsumeUsecase)

	return &TransactionConsumer{
		ms:                        ms,
		cfg:                       cfg,
		purchaseConsumer:          purchaseTransactionConsumer,
		purchaseReturnConsumer:    purchaseReturnStockConsumer,
		saleInvoiceConsumer:       saleInvoiceStockConsumer,
		saleInvoiceReturnConsumer: saleInvoiceReturnStockConsumer,
	}
}

func (pbc *TransactionConsumer) RegisterConsumer(ms *microservice.Microservice) {

	trxConsumerGroup := msConfig.GetEnv("TRANSACTION_CONSUMER_GROUP", "transaction-consumer-group-01")
	mq := microservice.NewMQ(pbc.cfg.MQConfig(), ms.Logger)

	// purchaseKafkaConfig := purchaseConfig.PurchaseMessageQueueConfig{}

	// mq.CreateTopicR(purchaseKafkaConfig.TopicCreated(), 5, 1, time.Hour*24*7)
	// mq.CreateTopicR(purchaseKafkaConfig.TopicUpdated(), 5, 1, time.Hour*24*7)
	// mq.CreateTopicR(purchaseKafkaConfig.TopicDeleted(), 5, 1, time.Hour*24*7)
	// mq.CreateTopicR(purchaseKafkaConfig.TopicBulkCreated(), 5, 1, time.Hour*24*7)
	// mq.CreateTopicR(purchaseKafkaConfig.TopicBulkUpdated(), 5, 1, time.Hour*24*7)
	// mq.CreateTopicR(purchaseKafkaConfig.TopicBulkDeleted(), 5, 1, time.Hour*24*7)

	// ms.Consume(pbc.cfg.MQConfig().URI(), purchaseKafkaConfig.TopicCreated(), trxConsumerGroup, time.Duration(-1), pbc.purchaseConsumer.ConsumeOnCreateOrUpdate)
	// ms.Consume(pbc.cfg.MQConfig().URI(), purchaseKafkaConfig.TopicUpdated(), trxConsumerGroup, time.Duration(-1), pbc.purchaseConsumer.ConsumeOnCreateOrUpdate)
	// ms.Consume(pbc.cfg.MQConfig().URI(), purchaseKafkaConfig.TopicDeleted(), trxConsumerGroup, time.Duration(-1), pbc.purchaseConsumer.ConsumeOnDelete)
	// ms.Consume(pbc.cfg.MQConfig().URI(), purchaseKafkaConfig.TopicBulkCreated(), trxConsumerGroup, time.Duration(-1), pbc.purchaseConsumer.ConsumeOnBulkCreateOrUpdate)
	// ms.Consume(pbc.cfg.MQConfig().URI(), purchaseKafkaConfig.TopicBulkUpdated(), trxConsumerGroup, time.Duration(-1), pbc.purchaseConsumer.ConsumeOnBulkCreateOrUpdate)
	// ms.Consume(pbc.cfg.MQConfig().URI(), purchaseKafkaConfig.TopicBulkDeleted(), trxConsumerGroup, time.Duration(-1), pbc.purchaseConsumer.ConsumeOnBulkDelete)

	//purchaseReturnKafkaConfig := purchaseReturnConfig.PurchaseReturnMessageQueueConfig{}
	// mq.CreateTopicR(purchaseReturnKafkaConfig.TopicCreated(), 5, 1, time.Hour*24*7)
	// mq.CreateTopicR(purchaseReturnKafkaConfig.TopicUpdated(), 5, 1, time.Hour*24*7)
	// mq.CreateTopicR(purchaseReturnKafkaConfig.TopicDeleted(), 5, 1, time.Hour*24*7)
	// mq.CreateTopicR(purchaseReturnKafkaConfig.TopicBulkCreated(), 5, 1, time.Hour*24*7)
	// mq.CreateTopicR(purchaseReturnKafkaConfig.TopicBulkUpdated(), 5, 1, time.Hour*24*7)
	// mq.CreateTopicR(purchaseReturnKafkaConfig.TopicBulkDeleted(), 5, 1, time.Hour*24*7)

	// ms.Consume(pbc.cfg.MQConfig().URI(), purchaseReturnKafkaConfig.TopicCreated(), trxConsumerGroup, time.Duration(-1), pbc.purchaseReturnConsumer.ConsumeOnCreateOrUpdate)
	// ms.Consume(pbc.cfg.MQConfig().URI(), purchaseReturnKafkaConfig.TopicUpdated(), trxConsumerGroup, time.Duration(-1), pbc.purchaseReturnConsumer.ConsumeOnCreateOrUpdate)
	// ms.Consume(pbc.cfg.MQConfig().URI(), purchaseReturnKafkaConfig.TopicDeleted(), trxConsumerGroup, time.Duration(-1), pbc.purchaseReturnConsumer.ConsumeOnDelete)
	// ms.Consume(pbc.cfg.MQConfig().URI(), purchaseReturnKafkaConfig.TopicBulkCreated(), trxConsumerGroup, time.Duration(-1), pbc.purchaseReturnConsumer.ConsumeOnBulkCreateOrUpdate)
	// ms.Consume(pbc.cfg.MQConfig().URI(), purchaseReturnKafkaConfig.TopicBulkUpdated(), trxConsumerGroup, time.Duration(-1), pbc.purchaseReturnConsumer.ConsumeOnBulkCreateOrUpdate)
	// ms.Consume(pbc.cfg.MQConfig().URI(), purchaseReturnKafkaConfig.TopicBulkDeleted(), trxConsumerGroup, time.Duration(-1), pbc.purchaseReturnConsumer.ConsumeOnBulkDelete)

	//saleInvoiceKafkaConfig := saleInvoiceConfig.SaleInvoiceMessageQueueConfig{}
	// mq.CreateTopicR(saleInvoiceKafkaConfig.TopicCreated(), 5, 1, time.Hour*24*7)
	// mq.CreateTopicR(saleInvoiceKafkaConfig.TopicUpdated(), 5, 1, time.Hour*24*7)
	// mq.CreateTopicR(saleInvoiceKafkaConfig.TopicDeleted(), 5, 1, time.Hour*24*7)
	// mq.CreateTopicR(saleInvoiceKafkaConfig.TopicBulkCreated(), 5, 1, time.Hour*24*7)
	// mq.CreateTopicR(saleInvoiceKafkaConfig.TopicBulkUpdated(), 5, 1, time.Hour*24*7)
	// mq.CreateTopicR(saleInvoiceKafkaConfig.TopicBulkDeleted(), 5, 1, time.Hour*24*7)

	// ms.Consume(pbc.cfg.MQConfig().URI(), saleInvoiceKafkaConfig.TopicCreated(), trxConsumerGroup, time.Duration(-1), pbc.saleInvoiceConsumer.ConsumeOnCreateOrUpdate)
	// ms.Consume(pbc.cfg.MQConfig().URI(), saleInvoiceKafkaConfig.TopicUpdated(), trxConsumerGroup, time.Duration(-1), pbc.saleInvoiceConsumer.ConsumeOnCreateOrUpdate)
	// ms.Consume(pbc.cfg.MQConfig().URI(), saleInvoiceKafkaConfig.TopicDeleted(), trxConsumerGroup, time.Duration(-1), pbc.saleInvoiceConsumer.ConsumeOnDelete)
	// ms.Consume(pbc.cfg.MQConfig().URI(), saleInvoiceKafkaConfig.TopicBulkCreated(), trxConsumerGroup, time.Duration(-1), pbc.saleInvoiceConsumer.ConsumeOnBulkCreateOrUpdate)
	// ms.Consume(pbc.cfg.MQConfig().URI(), saleInvoiceKafkaConfig.TopicBulkUpdated(), trxConsumerGroup, time.Duration(-1), pbc.saleInvoiceConsumer.ConsumeOnBulkCreateOrUpdate)
	// ms.Consume(pbc.cfg.MQConfig().URI(), saleInvoiceKafkaConfig.TopicBulkDeleted(), trxConsumerGroup, time.Duration(-1), pbc.saleInvoiceConsumer.ConsumeOnBulkDelete)

	saleInvoiceReturnKafkaConfig := saleInvoiceReturnConfig.SaleInvoiceReturnMessageQueueConfig{}
	mq.CreateTopicR(saleInvoiceReturnKafkaConfig.TopicCreated(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(saleInvoiceReturnKafkaConfig.TopicUpdated(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(saleInvoiceReturnKafkaConfig.TopicDeleted(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(saleInvoiceReturnKafkaConfig.TopicBulkCreated(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(saleInvoiceReturnKafkaConfig.TopicBulkUpdated(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(saleInvoiceReturnKafkaConfig.TopicBulkDeleted(), 5, 1, time.Hour*24*7)

	ms.Consume(pbc.cfg.MQConfig().URI(), saleInvoiceReturnKafkaConfig.TopicCreated(), trxConsumerGroup, time.Duration(-1), pbc.saleInvoiceReturnConsumer.ConsumeOnCreateOrUpdate)
	ms.Consume(pbc.cfg.MQConfig().URI(), saleInvoiceReturnKafkaConfig.TopicUpdated(), trxConsumerGroup, time.Duration(-1), pbc.saleInvoiceReturnConsumer.ConsumeOnCreateOrUpdate)
	ms.Consume(pbc.cfg.MQConfig().URI(), saleInvoiceReturnKafkaConfig.TopicDeleted(), trxConsumerGroup, time.Duration(-1), pbc.saleInvoiceReturnConsumer.ConsumeOnDelete)
	ms.Consume(pbc.cfg.MQConfig().URI(), saleInvoiceReturnKafkaConfig.TopicBulkCreated(), trxConsumerGroup, time.Duration(-1), pbc.saleInvoiceReturnConsumer.ConsumeOnBulkCreateOrUpdate)
	ms.Consume(pbc.cfg.MQConfig().URI(), saleInvoiceReturnKafkaConfig.TopicBulkUpdated(), trxConsumerGroup, time.Duration(-1), pbc.saleInvoiceReturnConsumer.ConsumeOnBulkCreateOrUpdate)
	ms.Consume(pbc.cfg.MQConfig().URI(), saleInvoiceReturnKafkaConfig.TopicBulkDeleted(), trxConsumerGroup, time.Duration(-1), pbc.saleInvoiceReturnConsumer.ConsumeOnBulkDelete)

}
