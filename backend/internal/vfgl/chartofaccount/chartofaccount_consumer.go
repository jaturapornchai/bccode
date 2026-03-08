package chartofaccount

import (
	"encoding/json"
	msConfig "smlcloudplatform/internal/config"
	"smlcloudplatform/internal/vfgl/chartofaccount/config"
	"smlcloudplatform/internal/vfgl/chartofaccount/models"
	"smlcloudplatform/internal/vfgl/chartofaccount/repositories"
	"smlcloudplatform/internal/vfgl/chartofaccount/services"
	"smlcloudplatform/pkg/microservice"
	"time"
)

func MigrationChartOfAccountTable(ms *microservice.Microservice, cfg msConfig.IConfig) error {
	pst := ms.Persister(cfg.PersisterConfig())
	pst.AutoMigrate(
		models.ChartOfAccountPG{},
	)
	return nil
}

func StartChartOfAccountConsumerCreated(ms *microservice.Microservice, cfg msConfig.IConfig, groupID string) {

	topicCreated := config.MQ_TOPIC_CREATED
	timeout := time.Duration(-1)
	mqConfig := cfg.MQConfig()
	mq := microservice.NewMQ(mqConfig, ms.Logger)
	mq.CreateTopicR(topicCreated, 5, 1, time.Hour*24*7)
	ms.Consume(mqConfig.URI(), topicCreated, groupID, timeout, func(ctx microservice.IContext) error {
		moduleName := "comsume chartofaccount created"

		pst := ms.Persister(cfg.PersisterConfig())
		msg := ctx.ReadInput()

		ms.Logger.Debugf("Consume Journal Created : %v", msg)
		doc := models.ChartOfAccountDoc{}
		err := json.Unmarshal([]byte(msg), &doc)

		if err != nil {
			ms.Logger.Errorf(moduleName, err.Error())
		}

		repo := repositories.NewChartOfAccountPgRepository(pst)
		svc := services.NewChartOfAccountConsumeService(repo)

		_, err = svc.Upsert(doc.ShopID, doc)

		if err != nil {
			ms.Logger.Errorf(moduleName, err.Error())
		}
		return nil
	})
}

func StartChartOfAccountConsumerUpdated(ms *microservice.Microservice, cfg msConfig.IConfig, groupID string) {

	topicUpdated := config.MQ_TOPIC_UPDATED
	timeout := time.Duration(-1)
	mqConfig := cfg.MQConfig()
	mq := microservice.NewMQ(mqConfig, ms.Logger)
	mq.CreateTopicR(topicUpdated, 5, 1, time.Hour*24*7)
	ms.Consume(mqConfig.URI(), topicUpdated, groupID, timeout, func(ctx microservice.IContext) error {
		moduleName := "comsume chartofaccount updated"

		pst := ms.Persister(cfg.PersisterConfig())
		msg := ctx.ReadInput()

		ms.Logger.Debugf("Consume Journal Updated : %v", msg)
		doc := models.ChartOfAccountDoc{}
		err := json.Unmarshal([]byte(msg), &doc)

		if err != nil {
			ms.Logger.Errorf(moduleName, err.Error())
		}

		repo := repositories.NewChartOfAccountPgRepository(pst)
		svc := services.NewChartOfAccountConsumeService(repo)

		_, err = svc.Upsert(doc.ShopID, doc)

		if err != nil {
			ms.Logger.Errorf(moduleName, err.Error())
		}
		return nil
	})
}

func StartChartOfAccountConsumerDeleted(ms *microservice.Microservice, cfg msConfig.IConfig, groupID string) {

	topicDeleted := config.MQ_TOPIC_DELETED
	timeout := time.Duration(-1)
	mqConfig := cfg.MQConfig()
	mq := microservice.NewMQ(mqConfig, ms.Logger)
	mq.CreateTopicR(topicDeleted, 5, 1, time.Hour*24*7)
	ms.Consume(mqConfig.URI(), topicDeleted, groupID, timeout, func(ctx microservice.IContext) error {
		moduleName := "comsume chartofaccount deleted"

		pst := ms.Persister(cfg.PersisterConfig())
		msg := ctx.ReadInput()

		ms.Logger.Debugf("Consume Journal Deleted : %v", msg)
		doc := models.ChartOfAccountDoc{}
		err := json.Unmarshal([]byte(msg), &doc)

		if err != nil {
			ms.Logger.Errorf(moduleName, err.Error())
		}

		repo := repositories.NewChartOfAccountPgRepository(pst)
		svc := services.NewChartOfAccountConsumeService(repo)

		err = svc.Delete(doc.ShopID, doc.AccountCode)

		if err != nil {
			ms.Logger.Errorf(moduleName, err.Error())
		}
		return nil
	})
}

func StartChartOfAccountConsumerBlukCreated(ms *microservice.Microservice, cfg msConfig.IConfig, groupID string) {

	topicCreated := config.MQ_TOPIC_BULK_CREATED
	timeout := time.Duration(-1)
	mqConfig := cfg.MQConfig()
	mq := microservice.NewMQ(mqConfig, ms.Logger)
	mq.CreateTopicR(topicCreated, 5, 1, time.Hour*24*7)
	ms.Consume(mqConfig.URI(), topicCreated, groupID, timeout, func(ctx microservice.IContext) error {
		moduleName := "comsume chartofaccount bulk created"

		pst := ms.Persister(cfg.PersisterConfig())
		msg := ctx.ReadInput()
		ms.Logger.Debugf("Consume Journal Bulk Created : %v", msg)

		docList := []models.ChartOfAccountDoc{}
		err := json.Unmarshal([]byte(msg), &docList)
		if err != nil {
			ms.Logger.Errorf(moduleName, err.Error())
		}

		repo := repositories.NewChartOfAccountPgRepository(pst)
		svc := services.NewChartOfAccountConsumeService(repo)

		err = svc.SaveInBatch(docList)

		if err != nil {
			ms.Logger.Errorf(moduleName, err.Error())
		}
		return nil
	})

}
