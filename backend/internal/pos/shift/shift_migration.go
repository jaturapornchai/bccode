package shift

import (
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/pos/shift/models"
	"smlcloudplatform/pkg/microservice"
)

func MigrationDatabase(ms *microservice.Microservice, cfg config.IConfig) {
	pst := ms.Persister(cfg.PersisterConfig())
	pst.AutoMigrate(
		models.ShiftPG{},
	)
}
