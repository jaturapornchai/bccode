package bom

import (
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/product/bom/models"
	"smlcloudplatform/pkg/microservice"
)

func MigrationDatabase(ms *microservice.Microservice, cfg config.IConfig) {
	pst := ms.Persister(cfg.PersisterConfig())
	pst.AutoMigrate(
		models.ProductBarcodeBOMViewPG{},
	)
}
