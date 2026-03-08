package creditor

import (
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/debtaccount/creditor/models"
	"smlcloudplatform/pkg/microservice"
)

func MigrationDatabase(ms *microservice.Microservice, cfg config.IConfig) {
	pst := ms.Persister(cfg.PersisterConfig())
	pst.AutoMigrate(
		models.CreditorPG{},
	)
}
