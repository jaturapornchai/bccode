package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/saleinvoicebomprice/config"
	"smlcloudplatform/internal/transaction/saleinvoicebomprice/models"
	"smlcloudplatform/pkg/microservice"
)

type ISaleInvoiceBOMPriceMessageQueueRepository interface {
	Create(doc models.SaleInvoiceBomPriceDoc) error
	Update(doc models.SaleInvoiceBomPriceDoc) error
	Delete(doc models.SaleInvoiceBomPriceDoc) error
	CreateInBatch(docList []models.SaleInvoiceBomPriceDoc) error
	UpdateInBatch(docList []models.SaleInvoiceBomPriceDoc) error
	DeleteInBatch(docList []models.SaleInvoiceBomPriceDoc) error
}

type SaleInvoiceBOMPriceMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.SaleInvoiceBomPriceDoc]
}

func NewSaleInvoiceBOMPriceMessageQueueRepository(prod microservice.IProducer) SaleInvoiceBOMPriceMessageQueueRepository {
	mqKey := ""

	insRepo := SaleInvoiceBOMPriceMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.SaleInvoiceBomPriceDoc](prod, config.SaleInvoiceBOMPriceMessageQueueConfig{}, "")
	return insRepo
}
