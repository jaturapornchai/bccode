package api

import (
	"smlcloudplatform/internal/logger"
	"smlcloudplatform/internal/shop"
	shopModel "smlcloudplatform/internal/shop/models"
	chartOfAccountModels "smlcloudplatform/internal/vfgl/chartofaccount/models"
	journalModels "smlcloudplatform/internal/vfgl/journal/models"
	"smlcloudplatform/pkg/microservice"
)

type IMigrationService interface {
	ImportShop(shops []shopModel.ShopDoc) error
	ImportJournal(journals []journalModels.JournalDoc) error
	ImportChartOfAccount(chartOfAccounts []chartOfAccountModels.ChartOfAccountDoc) error
}

type MigrationService struct {
	logger         logger.ILogger
	mongoPersister microservice.IPersisterMongo
	mqPersister    microservice.IProducer
	userShopRepo   shop.IShopUserRepository
}

func NewMigrationService(logger logger.ILogger, mongoPersister microservice.IPersisterMongo, mqPersister microservice.IProducer) *MigrationService {

	userShopRepo := shop.NewShopUserRepository(mongoPersister)

	return &MigrationService{
		logger:         logger,
		mongoPersister: mongoPersister,
		mqPersister:    mqPersister,
		userShopRepo:   userShopRepo,
	}
}
