package paymentdetail

import (
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/transaction/paymentdetail/models"
	"smlcloudplatform/pkg/microservice"
)

func MigrationDatabase(ms *microservice.Microservice, cfg config.IConfig) error {
	pst := ms.Persister(cfg.PersisterConfig())
	pst.AutoMigrate(
		models.TransactionPaymentDetail{},
	)
	return nil
}
