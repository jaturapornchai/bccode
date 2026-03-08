package productbarcode

import (
	msConfig "smlcloudplatform/internal/config"
	"smlcloudplatform/internal/product/productbarcode/models"
	"smlcloudplatform/pkg/microservice"
)

func MigrationDatabase(ms *microservice.Microservice, cfg msConfig.IConfig) error {
	pst := ms.Persister(cfg.PersisterConfig())
	pst.AutoMigrate(
		models.ProductBarcodePg{},
	)
	return nil
}
