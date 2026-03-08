package apdepositpaymentrefund

import (
	pkgConfig "smlcloudplatform/internal/config"
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/pkg/microservice"
)

func MigrationDatabase(ms *microservice.Microservice, cfg pkgConfig.IConfig) error {
	pst := ms.Persister(cfg.PersisterConfig())
	pst.AutoMigrate(
		models.APDepositPaymentRefundTransactionPG{},
		models.APDepositPaymentRefundTransactionDetailPG{},
	)
	return nil
}
