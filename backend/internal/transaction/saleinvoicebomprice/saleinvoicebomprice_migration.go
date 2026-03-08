package saleinvoicebomprice

import (
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/transaction/saleinvoicebomprice/models"
	"smlcloudplatform/pkg/microservice"
)

func MigrationDatabase(ms *microservice.Microservice, cfg config.IConfig) {
	pst := ms.Persister(cfg.PersisterConfig())
	pst.AutoMigrate(
		models.SaleInvoiceBomPricePg{},
	)
}
