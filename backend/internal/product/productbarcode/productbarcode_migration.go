package productbarcode

import (
	msConfig "smlcloudplatform/internal/config"
	"smlcloudplatform/internal/product/productbarcode/models"
	"smlcloudplatform/pkg/microservice"
)

func MigrationDatabase(ms *microservice.Microservice, cfg msConfig.IConfig) error {
	pst := ms.Persister(cfg.PersisterConfig())
	if err := pst.AutoMigrate(
		models.ProductBarcodePg{},
	); err != nil {
		return err
	}
	return pst.Exec(`
		DROP INDEX IF EXISTS uniq_productbarcode_holding_itemcode_barcode;
		CREATE UNIQUE INDEX IF NOT EXISTS uniq_productbarcode_company_barcode
			ON productbarcode (holding_code, businesscode, barcode)
	`)
}
