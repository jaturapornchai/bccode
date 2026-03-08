package api

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	shopModel "smlcloudplatform/internal/shop/models"
	accountModel "smlcloudplatform/internal/vfgl/chartofaccount/models"
	journalModels "smlcloudplatform/internal/vfgl/journal/models"
	"smlcloudplatform/pkg/microservice"
)

type IMigrationAPI interface {
	ImportJournal(ctx microservice.IContext) error
	ImportShop(ctx microservice.IContext) error
	ImportChartOfAccount(ctx microservice.IContext) error
	RegisterHttp(ms *microservice.Microservice)
}

type MigrationAPI struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc IMigrationService
}

func NewMigrationAPI(ms *microservice.Microservice, cfg config.IConfig) *MigrationAPI {

	mqPersister := ms.Producer(cfg.MQConfig())
	svc := NewMigrationService(ms.Logger, ms.MongoPersister(cfg.MongoPersisterConfig()), mqPersister)
	return &MigrationAPI{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (m *MigrationAPI) RegisterHttp() {
	m.ms.POST("/migrationtools/journalimport", m.ImportJournal)
	m.ms.POST("/migrationtools/shopimport", m.ImportShop)
	m.ms.POST("/migrationtools/chartimport", m.ImportChartOfAccount)
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
