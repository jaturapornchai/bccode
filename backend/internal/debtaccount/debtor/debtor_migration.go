package debtor

import (
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/debtaccount/debtor/models"
	"smlcloudplatform/pkg/microservice"
)

func MigrationDatabase(ms *microservice.Microservice, cfg config.IConfig) {
	pst := ms.Persister(cfg.PersisterConfig())
	pst.AutoMigrate(
		models.DebtorPG{},
	)
}
