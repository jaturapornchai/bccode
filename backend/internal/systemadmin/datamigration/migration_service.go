package datamigration

import (
	"smlcloudplatform/internal/shop"
	shopModel "smlcloudplatform/internal/shop/models"
	accountGroupRepositories "smlcloudplatform/internal/vfgl/accountgroup/repositories"
	accountModel "smlcloudplatform/internal/vfgl/chartofaccount/models"
	chartofaccountrepositories "smlcloudplatform/internal/vfgl/chartofaccount/repositories"
	journalBookRepositories "smlcloudplatform/internal/vfgl/journalbook/repositories"

	// chartOfAccountServices "smlcloudplatform/internal/vfgl/chartofaccount/services"

	"smlcloudplatform/internal/logger"
	journalModels "smlcloudplatform/internal/vfgl/journal/models"
	"smlcloudplatform/pkg/microservice"
)

type IMigrationService interface {
	ImportShop(shops []shopModel.ShopDoc) error
	ImportJournal(journals []journalModels.JournalDoc) error
	ImportChartOfAccount(chartOfAccounts []accountModel.ChartOfAccountDoc) error
	ResyncChartOfAccount(charts []accountModel.ChartOfAccountDoc) error
	InitCenterAccountGroup() error
	InitialChartOfAccountCenter() error
	InitJournalBookCenter() error
}

type MigrationService struct {
	logger          logger.ILogger
	mongoPersister  microservice.IPersisterMongo
	mqPersister     microservice.IProducer
	userShopRepo    shop.IShopUserRepository
	chartRepo       chartofaccountrepositories.IChartOfAccountRepository
	chartMQRepo     chartofaccountrepositories.IChartOfAccountMQRepository
	journalBookRepo journalBookRepositories.IJournalBookMongoRepository

	accountGroupRepo accountGroupRepositories.IAccountGroupMongoRepository
	// chartService   chartOfAccountServices.ChartOfAccountHttpService
}

func NewMigrationService(logger logger.ILogger, mongoPersister microservice.IPersisterMongo, mqPersister microservice.IProducer) *MigrationService {

	chartRepo := chartofaccountrepositories.NewChartOfAccountRepository(mongoPersister)
	chartMQRepo := chartofaccountrepositories.NewChartOfAccountMQRepository(mqPersister)

	userShopRepo := shop.NewShopUserRepository(mongoPersister)
	accountGroupRepo := accountGroupRepositories.NewAccountGroupMongoRepository(mongoPersister)
	journalBookRepo := journalBookRepositories.NewJournalBookMongoRepository(mongoPersister)
	// chartService := chartOfAccountServices.NewChartOfAccountHttpService(chartRepo, nil, chartMQRepo)
	return &MigrationService{
		logger:           logger,
		mongoPersister:   mongoPersister,
		mqPersister:      mqPersister,
		userShopRepo:     userShopRepo,
		chartRepo:        chartRepo,
		chartMQRepo:      chartMQRepo,
		accountGroupRepo: accountGroupRepo,
		journalBookRepo:  journalBookRepo,
	}
}
