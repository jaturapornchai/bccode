package datamigration

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	shopModel "smlcloudplatform/internal/shop/models"
	adminModels "smlcloudplatform/internal/systemadmin/models"
	accountModel "smlcloudplatform/internal/vfgl/chartofaccount/models"
	journalModels "smlcloudplatform/internal/vfgl/journal/models"
	"smlcloudplatform/pkg/microservice"
)

type IMigrationAPI interface {
	ImportJournal(ctx microservice.IContext) error
	ImportShop(ctx microservice.IContext) error
	ImportChartOfAccount(ctx microservice.IContext) error
	RegisterHttp(ms *microservice.Microservice, pathPrefix string)

	InitCenterAccountGroup(ctx microservice.IContext) error
	InitialChartOfAccountCenter(ctx microservice.IContext) error
	InitCenterJournalBook(ctx microservice.IContext) error
}

type MigrationAPI struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc IMigrationService
}

func NewMigrationAPI(ms *microservice.Microservice, cfg config.IConfig) *MigrationAPI {

	mqPersister := ms.Producer(cfg.MQConfig())

	// ควร move ไปทำใน db migration
	// mq := microservice.NewMQ(cfg.MQConfig(), ms.Logger)

	// chartOfAccountKafkaConfig := chartOfAccountConfig.ChartOfAccountMessageQueueConfig{}
	// mq.CreateTopicR(chartOfAccountKafkaConfig.TopicCreated(), 5, 1, time.Hour*24*7)
	// mq.CreateTopicR(chartOfAccountKafkaConfig.TopicDeleted(), 5, 1, time.Hour*24*7)
	// mq.CreateTopicR(chartOfAccountKafkaConfig.TopicUpdated(), 5, 1, time.Hour*24*7)
	// mq.CreateTopicR(chartOfAccountKafkaConfig.TopicBulkCreated(), 5, 1, time.Hour*24*7)
	// mq.CreateTopicR(chartOfAccountKafkaConfig.TopicBulkDeleted(), 5, 1, time.Hour*24*7)
	// mq.CreateTopicR(chartOfAccountKafkaConfig.TopicBulkUpdated(), 5, 1, time.Hour*24*7)

	// journalKafkaConfig := journalConfig.JournalMessageQueueConfig{}
	// mq.CreateTopicR(journalKafkaConfig.TopicCreated(), 5, 1, time.Hour*24*7)
	// mq.CreateTopicR(journalKafkaConfig.TopicDeleted(), 5, 1, time.Hour*24*7)
	// mq.CreateTopicR(journalKafkaConfig.TopicUpdated(), 5, 1, time.Hour*24*7)
	// mq.CreateTopicR(journalKafkaConfig.TopicBulkCreated(), 5, 1, time.Hour*24*7)
	// mq.CreateTopicR(journalKafkaConfig.TopicBulkDeleted(), 5, 1, time.Hour*24*7)
	// mq.CreateTopicR(journalKafkaConfig.TopicBulkUpdated(), 5, 1, time.Hour*24*7)

	// productBarcodeKafkaConfig := productBarcodeConfig.ProductMessageQueueConfig{}
	// mq.CreateTopicR(productBarcodeKafkaConfig.TopicCreated(), 5, 1, time.Hour*24*7)
	// mq.CreateTopicR(productBarcodeKafkaConfig.TopicDeleted(), 5, 1, time.Hour*24*7)
	// mq.CreateTopicR(productBarcodeKafkaConfig.TopicUpdated(), 5, 1, time.Hour*24*7)
	// mq.CreateTopicR(productBarcodeKafkaConfig.TopicBulkCreated(), 5, 1, time.Hour*24*7)
	// mq.CreateTopicR(productBarcodeKafkaConfig.TopicBulkDeleted(), 5, 1, time.Hour*24*7)
	// mq.CreateTopicR(productBarcodeKafkaConfig.TopicBulkUpdated(), 5, 1, time.Hour*24*7)

	svc := NewMigrationService(ms.Logger, ms.MongoPersister(cfg.MongoPersisterConfig()), mqPersister)
	return &MigrationAPI{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (m *MigrationAPI) RegisterHttp(ms *microservice.Microservice, pathPrefix string) {
	ms.POST(pathPrefix+"/migrationtools/journalimport", m.ImportJournal)
	ms.POST(pathPrefix+"/migrationtools/shopimport", m.ImportShop)
	ms.POST(pathPrefix+"/migrationtools/chartimport", m.ImportChartOfAccount)
	ms.POST(pathPrefix+"/migrationtools/chartresync", m.ChartReSync)
	ms.POST(pathPrefix+"/migrationtools/initcenteraccountgroup", m.InitCenterAccountGroup)
	ms.POST(pathPrefix+"/migrationtools/initcenterchart", m.InitialChartOfAccountCenter)
	ms.POST(pathPrefix+"/migrationtools/initcenterjournalbook", m.InitCenterJournalBook)
}

func (m *MigrationAPI) ImportJournal(ctx microservice.IContext) error {

	input := ctx.ReadInput()

	docReq := &[]journalModels.JournalDoc{}
	err := json.Unmarshal([]byte(input), &docReq)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	err = m.svc.ImportJournal(*docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ResponseSuccess{
		Success: true,
	})
	// ctx.Response(http.StatusOK, common.ApiResponse{
	// 	Success: true,
	// 	Data:    docReq,
	// })
	return nil
}

func (m *MigrationAPI) ImportShop(ctx microservice.IContext) error {

	input := ctx.ReadInput()

	docReq := &[]shopModel.ShopDoc{}
	err := json.Unmarshal([]byte(input), &docReq)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	err = m.svc.ImportShop(*docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ResponseSuccess{
		Success: true,
	})

	// ctx.Response(http.StatusOK, common.ApiResponse{
	// 	Success: true,
	// 	Data:    docReq,
	// })
	return nil
}

func (m *MigrationAPI) ImportChartOfAccount(ctx microservice.IContext) error {

	input := ctx.ReadInput()

	docReq := &[]accountModel.ChartOfAccountDoc{}
	err := json.Unmarshal([]byte(input), &docReq)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	err = m.svc.ImportChartOfAccount(*docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ResponseSuccess{
		Success: true,
	})

	// ctx.Response(http.StatusOK, common.ApiResponse{
	// 	Success: true,
	// 	Data:    docReq,
	// })
	return nil
}

func (m *MigrationAPI) ChartReSync(ctx microservice.IContext) error {

	input := ctx.ReadInput()

	docReq := &[]accountModel.ChartOfAccountDoc{}
	err := json.Unmarshal([]byte(input), &docReq)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	err = m.svc.ResyncChartOfAccount(*docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ResponseSuccess{
		Success: true,
	})

	// ctx.Response(http.StatusOK, common.ApiResponse{
	// 	Success: true,
	// 	Data:    docReq,
	// })
	return nil
}

func (m *MigrationAPI) InitCenterAccountGroup(ctx microservice.IContext) error {

	// input := ctx.ReadInput()

	// docReq := &[]accountGroupModels.AccountGroupDoc{}
	// err := json.Unmarshal([]byte(input), &docReq)
	// if err != nil {
	// 	ctx.ResponseError(http.StatusBadRequest, err.Error())
	// 	return err
	// }

	input := ctx.ReadInput()
	var req adminModels.RequestReSyncTenant

	err := json.Unmarshal([]byte(input), &req)

	if err != nil {
		ctx.Response(http.StatusBadRequest, common.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
		return err
	}

	if req.HoldingCode != "999999999" {
		ctx.ResponseError(http.StatusBadRequest, "HoldingCode Not Valid")
		return err
	}

	err = m.svc.InitCenterAccountGroup()

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ResponseSuccess{
		Success: true,
	})
	return nil
}

func (m *MigrationAPI) InitialChartOfAccountCenter(ctx microservice.IContext) error {

	input := ctx.ReadInput()
	var req adminModels.RequestReSyncTenant

	err := json.Unmarshal([]byte(input), &req)

	if err != nil {
		ctx.Response(http.StatusBadRequest, common.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
		return err
	}

	if req.HoldingCode != "999999999" {
		ctx.ResponseError(http.StatusBadRequest, "HoldingCode Not Valid")
		return err
	}

	err = m.svc.InitialChartOfAccountCenter()

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ResponseSuccess{
		Success: true,
	})
	return nil
}

func (m *MigrationAPI) InitCenterJournalBook(ctx microservice.IContext) error {
	input := ctx.ReadInput()
	var req adminModels.RequestReSyncTenant

	err := json.Unmarshal([]byte(input), &req)

	if err != nil {
		ctx.Response(http.StatusBadRequest, common.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
		return err
	}

	if req.HoldingCode != "999999999" {
		ctx.ResponseError(http.StatusBadRequest, "HoldingCode Not Valid")
		return err
	}

	err = m.svc.InitJournalBookCenter()

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ResponseSuccess{
		Success: true,
	})
	return nil
}
