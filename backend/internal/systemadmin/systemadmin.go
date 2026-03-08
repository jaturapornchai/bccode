package systemadmin

import (
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/systemadmin/accountadmin"
	"smlcloudplatform/internal/systemadmin/chartofaccountadmin"
	"smlcloudplatform/internal/systemadmin/creditoradmin"
	"smlcloudplatform/internal/systemadmin/datamigration"
	"smlcloudplatform/internal/systemadmin/debtoradmin"
	journal "smlcloudplatform/internal/systemadmin/journaladmin"
	"smlcloudplatform/internal/systemadmin/productadmin"
	"smlcloudplatform/internal/systemadmin/servicetools"
	"smlcloudplatform/internal/systemadmin/shopadmin"
	"smlcloudplatform/internal/systemadmin/transactionadmin"
	"smlcloudplatform/pkg/microservice"
)

const SYSTEM_ADMIN_ROUTE_PREFIX = "/systemadm"

type ISystemAdmin interface {
	RegisterHttp()
}

type SystemAdmin struct {
	ms                      *microservice.Microservice
	migrationHttp           datamigration.IMigrationAPI
	serviceTools            servicetools.IServiceTools
	shopAdminHttp           shopadmin.IShopAdminHttp
	accountAdminHttp        accountadmin.IAccountAdminHttp
	chartOfAccountAdminHttp chartofaccountadmin.IChartOfAccountAdminHttp
	productAdminHttp        productadmin.IProductAdminHttp
	creditorAdminHttp       creditoradmin.ICreditorAdminHttp
	debtorAdminHttp         debtoradmin.IDebtorAdminHttp
	transactionAdminHttp    transactionadmin.ITransactionAdminHttp
	journalAdminHttp        journal.IJournalTransactionAdminHttp
}

func NewSystemAdmin(ms *microservice.Microservice, cfg config.IConfig) ISystemAdmin {

	migrationHttp := datamigration.NewMigrationAPI(ms, cfg)
	servicetools := servicetools.NewServiceTools(ms.Logger, cfg, ms.MongoPersister(cfg.MongoPersisterConfig()))
	shopAdminHttp := shopadmin.NewShopAdminHttp(ms, cfg)
	accountAdminHttp := accountadmin.NewAccountAdminHttp(ms, cfg)
	productAdminHttp := productadmin.NewProductAdminHttp(ms, cfg)
	creditorAdminHttp := creditoradmin.NewCreditorAdminHttp(ms, cfg)
	debtorAdminHttp := debtoradmin.NewDebtorAdminHttp(ms, cfg)
	transactionAdminHttp := transactionadmin.NewTransactionAdminHttp(ms, cfg)
	chartOfAccountAdminHttp := chartofaccountadmin.NewChartOfAccountAdminHttp(ms, cfg)
	journalAdminHttp := journal.NewJournalTransactionAdminHttp(ms, cfg)

	return &SystemAdmin{
		ms:                      ms,
		migrationHttp:           migrationHttp,
		serviceTools:            servicetools,
		shopAdminHttp:           shopAdminHttp,
		accountAdminHttp:        accountAdminHttp,
		productAdminHttp:        productAdminHttp,
		creditorAdminHttp:       creditorAdminHttp,
		debtorAdminHttp:         debtorAdminHttp,
		transactionAdminHttp:    transactionAdminHttp,
		chartOfAccountAdminHttp: chartOfAccountAdminHttp,
		journalAdminHttp:        journalAdminHttp,
	}
}

func (s *SystemAdmin) RegisterHttp() {
	s.migrationHttp.RegisterHttp(s.ms, SYSTEM_ADMIN_ROUTE_PREFIX)
	s.serviceTools.RegisterHttp(s.ms, SYSTEM_ADMIN_ROUTE_PREFIX)
	s.shopAdminHttp.RegisterHttp(s.ms, SYSTEM_ADMIN_ROUTE_PREFIX)
	s.accountAdminHttp.RegisterHttp(s.ms, SYSTEM_ADMIN_ROUTE_PREFIX)
	s.productAdminHttp.RegisterHttp(s.ms, SYSTEM_ADMIN_ROUTE_PREFIX)
	s.transactionAdminHttp.RegisterHttp(s.ms, SYSTEM_ADMIN_ROUTE_PREFIX)
	s.chartOfAccountAdminHttp.RegisterHttp(s.ms, SYSTEM_ADMIN_ROUTE_PREFIX)
	s.journalAdminHttp.RegisterHttp(s.ms, SYSTEM_ADMIN_ROUTE_PREFIX)
	s.creditorAdminHttp.RegisterHttp(s.ms, SYSTEM_ADMIN_ROUTE_PREFIX)
	s.debtorAdminHttp.RegisterHttp(s.ms, SYSTEM_ADMIN_ROUTE_PREFIX)
}
