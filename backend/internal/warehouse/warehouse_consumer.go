package warehouse

import (
	pkgConfig "smlcloudplatform/internal/config"
	"smlcloudplatform/internal/logger"
	"smlcloudplatform/internal/warehouse/config"
	"smlcloudplatform/internal/warehouse/repositories"
	"smlcloudplatform/internal/warehouse/services"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type WarehouseConsumer struct {
	ms              *microservice.Microservice
	cfg             pkgConfig.IConfig
	svc             services.IWarehouseConsumerService
	warehousePhaser services.WarehousePhaser
}

func NewWarehouseConsumer(
	ms *microservice.Microservice,
	cfg pkgConfig.IConfig,
	svc services.IWarehouseConsumerService,
	warehousePhaser services.WarehousePhaser,
) *WarehouseConsumer {
	return &WarehouseConsumer{
		ms:              ms,
		cfg:             cfg,
		svc:             svc,
		warehousePhaser: warehousePhaser,
	}
}

func InitWarehouseConsumer(ms *microservice.Microservice, cfg pkgConfig.IConfig) *WarehouseConsumer {
	persister := ms.Persister(cfg.PersisterConfig())
	repo := repositories.NewWarehousePGRepository(persister)
	warehouseConsumerService := services.NewWarehouseConsumerService(repo)
	warehousePhaser := services.WarehousePhaser{}

	warehouseConsumer := NewWarehouseConsumer(ms, cfg, warehouseConsumerService, warehousePhaser)
	return warehouseConsumer
}

func (c *WarehouseConsumer) RegisterConsumer(ms *microservice.Microservice) {

	consumerGroup := pkgConfig.GetEnv("WAREHOUSE_CONSUMER_GROUP", "warehouse-consumer-group-01")
	mq := microservice.NewMQ(c.cfg.MQConfig(), ms.Logger)

	warehouseKafkaConfig := config.WarehouseMessageQueueConfig{}

	mq.CreateTopicR(warehouseKafkaConfig.TopicCreated(), 5, 1, 24*7)
	mq.CreateTopicR(warehouseKafkaConfig.TopicUpdated(), 5, 1, 24*7)
	mq.CreateTopicR(warehouseKafkaConfig.TopicDeleted(), 5, 1, 24*7)
	mq.CreateTopicR(warehouseKafkaConfig.TopicBulkCreated(), 5, 1, 24*7)
	mq.CreateTopicR(warehouseKafkaConfig.TopicBulkUpdated(), 5, 1, 24*7)
	mq.CreateTopicR(warehouseKafkaConfig.TopicBulkDeleted(), 5, 1, 24*7)

	ms.Consume(c.cfg.MQConfig().URI(), warehouseKafkaConfig.TopicCreated(), consumerGroup, time.Duration(-1), c.ConsumeOnCreateOrUpdate)
	ms.Consume(c.cfg.MQConfig().URI(), warehouseKafkaConfig.TopicUpdated(), consumerGroup, time.Duration(-1), c.ConsumeOnCreateOrUpdate)
	ms.Consume(c.cfg.MQConfig().URI(), warehouseKafkaConfig.TopicDeleted(), consumerGroup, time.Duration(-1), c.ConsumeOnDelete)
	ms.Consume(c.cfg.MQConfig().URI(), warehouseKafkaConfig.TopicBulkCreated(), consumerGroup, time.Duration(-1), c.ConsumeOnBulkCreateOrUpdate)
	ms.Consume(c.cfg.MQConfig().URI(), warehouseKafkaConfig.TopicBulkUpdated(), consumerGroup, time.Duration(-1), c.ConsumeOnBulkCreateOrUpdate)
	ms.Consume(c.cfg.MQConfig().URI(), warehouseKafkaConfig.TopicBulkDeleted(), consumerGroup, time.Duration(-1), c.ConsumeOnBulkDelete)

}

func (c *WarehouseConsumer) ConsumeOnCreateOrUpdate(ctx microservice.IContext) error {

	msg := ctx.ReadInput()
	doc, err := c.warehousePhaser.PhaseSingleDoc(msg)

	if err != nil {
		logger.GetLogger().Errorf("Cannot phase Warehouse single doc: %v", err)
		return err
	}

	err = c.svc.Upsert(doc.ShopID, doc.GuidFixed, *doc)

	if err != nil {
		logger.GetLogger().Errorf("Cannot upsert Warehouse doc: %v", err)
		return err
	}

	return nil
}

func (c *WarehouseConsumer) ConsumeOnDelete(ctx microservice.IContext) error {
	msg := ctx.ReadInput()
	doc, err := c.warehousePhaser.PhaseSingleDoc(msg)

	if err != nil {
		logger.GetLogger().Errorf("Cannot phase Warehouse single doc: %v", err)
		return err
	}

	err = c.svc.Delete(doc.ShopID, doc.GuidFixed)

	if err != nil {
		logger.GetLogger().Errorf("Cannot delete Warehouse doc: %v", err)
		return err
	}

	return nil
}

func (c *WarehouseConsumer) ConsumeOnBulkCreateOrUpdate(ctx microservice.IContext) error {
	msg := ctx.ReadInput()
	docs, err := c.warehousePhaser.PhaseMultipleDoc(msg)
	if err != nil {
		logger.GetLogger().Errorf("Cannot phase Warehouse multiple doc: %v", err)
		return err
	}

	for _, doc := range *docs {
		err = c.svc.Upsert(doc.ShopID, doc.GuidFixed, doc)
		if err != nil {
			logger.GetLogger().Errorf("Cannot upsert Warehouse doc: %v", err)
			return err
		}
	}

	return nil
}

func (c *WarehouseConsumer) ConsumeOnBulkDelete(ctx microservice.IContext) error {
	msg := ctx.ReadInput()
	docs, err := c.warehousePhaser.PhaseMultipleDoc(msg)
	if err != nil {
		logger.GetLogger().Errorf("Cannot phase Warehouse multiple doc: %v", err)
		return err
	}

	for _, doc := range *docs {
		err = c.svc.Delete(doc.ShopID, doc.GuidFixed)
		if err != nil {
			logger.GetLogger().Errorf("Cannot delete Warehouse doc: %v", err)
			return err
		}
	}
	return nil
}
