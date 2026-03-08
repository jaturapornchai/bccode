package warehouse

import (
	pkgConfig "smlcloudplatform/internal/config"
	"smlcloudplatform/internal/warehouse/models"
	"smlcloudplatform/pkg/microservice"
)

func MigrationDatabase(ms *microservice.Microservice, cfg pkgConfig.IConfig) error {
	pst := ms.Persister(cfg.PersisterConfig())
	pst.AutoMigrate(
		models.WarehousePG{},
	)
	return nil
}
