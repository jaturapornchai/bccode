package journal

import (
	"encoding/json"
	sysConfig "smlcloudplatform/internal/config"
	"smlcloudplatform/internal/vfgl/journal/config"
	"smlcloudplatform/internal/vfgl/journal/models"
	"smlcloudplatform/internal/vfgl/journal/repositories"
	"smlcloudplatform/internal/vfgl/journal/services"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IJournalTransactionConsumer interface {
	RegisterConsumer(ms *microservice.Microservice)
	ConsumeOnCreateOrUpdate(ctx microservice.IContext) error
	ConsumeOnDelete(ctx microservice.IContext) error
	ConsumeOnBulkCreateOrUpdate(ctx microservice.IContext) error
	ConsumeOnBulkDelete(ctx microservice.IContext) error
}

type JournalTransactionConsumer struct {
	ms  *microservice.Microservice
	cfg sysConfig.IConfig
	svc services.IJournalConsumeService
}

func InitJournalTransactionConsumer(ms *microservice.Microservice, cfg sysConfig.IConfig) IJournalTransactionConsumer {

	persister := ms.Persister(cfg.PersisterConfig())

	repo := repositories.NewJournalPgRepository(persister)
	svc := services.NewJournalConsumeService(repo)

	return NewJournslTransactionConsumer(ms, cfg, svc)
}

func NewJournslTransactionConsumer(ms *microservice.Microservice, cfg sysConfig.IConfig, svc services.IJournalConsumeService) IJournalTransactionConsumer {

	return &JournalTransactionConsumer{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}

}
func MigrationJournalTable(ms *microservice.Microservice, cfg sysConfig.IConfig) error {
	pst := ms.Persister(cfg.PersisterConfig())
	pst.AutoMigrate(
		models.JournalPg{},
		models.JournalDetailPg{},
		models.JournalVatPg{},
		models.JournalTaxPg{},
	)
	return nil
}

func (c *JournalTransactionConsumer) RegisterConsumer(ms *microservice.Microservice) {

	trxConsumerGroup := sysConfig.GetEnv("CONSUMER_GROUP_NAME", "03")
	mq := microservice.NewMQ(c.cfg.MQConfig(), ms.Logger)
	journalKafkaConfig := config.JournalMessageQueueConfig{}

	mq.CreateTopicR(journalKafkaConfig.TopicCreated(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(journalKafkaConfig.TopicUpdated(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(journalKafkaConfig.TopicDeleted(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(journalKafkaConfig.TopicBulkCreated(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(journalKafkaConfig.TopicBulkUpdated(), 5, 1, time.Hour*24*7)
	mq.CreateTopicR(journalKafkaConfig.TopicBulkDeleted(), 5, 1, time.Hour*24*7)

	ms.Consume(c.cfg.MQConfig().URI(), journalKafkaConfig.TopicCreated(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnCreateOrUpdate)
	ms.Consume(c.cfg.MQConfig().URI(), journalKafkaConfig.TopicUpdated(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnCreateOrUpdate)
	ms.Consume(c.cfg.MQConfig().URI(), journalKafkaConfig.TopicDeleted(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnDelete)
	ms.Consume(c.cfg.MQConfig().URI(), journalKafkaConfig.TopicBulkCreated(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnBulkCreateOrUpdate)
	ms.Consume(c.cfg.MQConfig().URI(), journalKafkaConfig.TopicBulkUpdated(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnBulkCreateOrUpdate)
	ms.Consume(c.cfg.MQConfig().URI(), journalKafkaConfig.TopicBulkDeleted(), trxConsumerGroup, time.Duration(-1), c.ConsumeOnBulkDelete)

}

func (c *JournalTransactionConsumer) ConsumeOnCreateOrUpdate(ctx microservice.IContext) error {

	moduleName := "comsume journal created"
	msg := ctx.ReadInput()

	doc := models.JournalDoc{}
	err := json.Unmarshal([]byte(msg), &doc)
	if err != nil {
		c.ms.Logger.Errorf(moduleName, err.Error())
		return err
	}

	c.ms.Logger.Debugf("Journal created : %v, %v", doc.HoldingCode, doc.DocNo)
	_, err = c.svc.UpSert(doc.HoldingCode, doc.DocNo, doc)

	if err != nil {
		c.ms.Logger.Errorf(moduleName, err.Error())
		return err
	}
	return nil
}

func (c *JournalTransactionConsumer) ConsumeOnDelete(ctx microservice.IContext) error {
	moduleName := "comsume journal delete"

	msg := ctx.ReadInput()

	doc := models.JournalDoc{}
	err := json.Unmarshal([]byte(msg), &doc)

	if err != nil {
		c.ms.Logger.Errorf(moduleName, err.Error())
	}

	c.ms.Logger.Debugf("Journal delete : %v, %v", doc.HoldingCode, doc.DocNo)
	err = c.svc.Delete(doc.HoldingCode, doc.DocNo)

	if err != nil {
		c.ms.Logger.Errorf(moduleName, err.Error())
	}
	return nil
}

func (c *JournalTransactionConsumer) ConsumeOnBulkCreateOrUpdate(ctx microservice.IContext) error {
	moduleName := "comsume journal bulk created"

	msg := ctx.ReadInput()

	docList := []models.JournalDoc{}
	err := json.Unmarshal([]byte(msg), &docList)

	if err != nil {
		c.ms.Logger.Errorf(moduleName, err.Error())
	}

	for _, transaction := range docList {
		_, err = c.svc.UpSert(transaction.HoldingCode, transaction.DocNo, transaction)
		if err != nil {
			c.ms.Logger.Errorf(moduleName, err.Error())
		}
	}
	return nil
}

func (c *JournalTransactionConsumer) ConsumeOnBulkDelete(ctx microservice.IContext) error {
	moduleName := "comsume journal bulk deleted"

	msg := ctx.ReadInput()

	docList := []models.JournalDoc{}
	err := json.Unmarshal([]byte(msg), &docList)

	if err != nil {
		c.ms.Logger.Errorf(moduleName, err.Error())
	}

	for _, transaction := range docList {
		err := c.svc.Delete(transaction.HoldingCode, transaction.DocNo)

		if err != nil {
			c.ms.Logger.Errorf(moduleName, err.Error())
		}
	}
	return nil
}

// func StartJournalComsumeBlukCreated(ms *microservice.Microservice, cfg sysConfig.IConfig, groupID string) {

// 	topicCreated := config.MQ_TOPIC_BULK_CREATED
// 	timeout := time.Duration(-1)

// 	mqConfig := cfg.MQConfig()

// 	mq := microservice.NewMQ(mqConfig, ms.Logger)

// 	mq.CreateTopicR(topicCreated, 5, 1, time.Hour*24*7)

// 	ms.Consume(mqConfig.URI(), topicCreated, groupID, timeout, func(ctx microservice.IContext) error {
// 		moduleName := "comsume journal created"

// 		pst := ms.Persister(cfg.PersisterConfig())
// 		msg := ctx.ReadInput()

// 		ms.Logger.Debugf("Consume Journal Created : %v", msg)
// 		docList := []models.JournalDoc{}
// 		err := json.Unmarshal([]byte(msg), &docList)

// 		if err != nil {
// 			ms.Logger.Errorf(moduleName, err.Error())
// 		}

// 		repo := repositories.NewJournalPgRepository(pst)
// 		svc := services.NewJournalConsumeService(repo)

// 		err = svc.SaveInBatch(docList)

// 		if err != nil {
// 			ms.Logger.Errorf(moduleName, err.Error())
// 		}
// 		return nil
// 	})

// }
