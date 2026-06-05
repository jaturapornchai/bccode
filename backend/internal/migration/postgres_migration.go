package migration

import (
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/goapi/mydb"
	orgBranch "smlcloudplatform/internal/organization/branch/models"
	orgCompany "smlcloudplatform/internal/organization/company/models"
	pbModels "smlcloudplatform/internal/product/productbarcode/models"
	vfgl "smlcloudplatform/internal/vfgl/journal/models"
	whModels "smlcloudplatform/internal/warehouse/models"
	"smlcloudplatform/pkg/microservice"
)

func StartMigrateModel(ms *microservice.Microservice, cfg config.IConfig) error {
	// Initialize the dynamic DB connection/creation hook
	microservice.DBCheckHook = func(dbName string) error {
		_, err := mydb.GetGlobalConnectionFromPool(dbName)
		return err
	}

	// Register models for auto-migration inside tenant databases
	microservice.RegisterTenantModel(
		vfgl.JournalPg{},
		vfgl.JournalDetailPg{},
		vfgl.JournalVatPg{},
		vfgl.JournalTaxPg{},

		orgCompany.CompanyPg{},
		orgBranch.BranchPg{},
		whModels.WarehousePg{},
		whModels.CompanyWarehousePg{},
		whModels.ZonePg{},
		whModels.ShelfPg{},
		pbModels.ProductBarcodePg{},
	)

	// Run migration for the default/admin database
	pst := ms.Persister(cfg.PersisterConfig())
	pst.AutoMigrate(
		vfgl.JournalPg{},
		vfgl.JournalDetailPg{},
		vfgl.JournalVatPg{},
		vfgl.JournalTaxPg{},

		orgCompany.CompanyPg{},
		orgBranch.BranchPg{},
		whModels.WarehousePg{},
		whModels.CompanyWarehousePg{},
		whModels.ZonePg{},
		whModels.ShelfPg{},
		pbModels.ProductBarcodePg{},
	)

	return nil
}
