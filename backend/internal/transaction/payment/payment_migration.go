package payment

import (
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/transaction/payment/models"
	"smlcloudplatform/pkg/microservice"
)

func MigrationDatabase(ms *microservice.Microservice, cfg config.IConfig) error {
	pst := ms.Persister(cfg.PersisterConfig())
	pst.AutoMigrate(
		models.TransactionPayment{},
	)
	return nil
}
