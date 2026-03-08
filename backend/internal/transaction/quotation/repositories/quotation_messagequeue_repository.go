package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/quotation/config"
	"smlcloudplatform/internal/transaction/quotation/models"
	"smlcloudplatform/pkg/microservice"
)

type IQuotationMessageQueueRepository interface {
	Create(doc models.QuotationDoc) error
	Update(doc models.QuotationDoc) error
	Delete(doc models.QuotationDoc) error
	CreateInBatch(docList []models.QuotationDoc) error
	UpdateInBatch(docList []models.QuotationDoc) error
	DeleteInBatch(docList []models.QuotationDoc) error
}

type QuotationMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.QuotationDoc]
}

func NewQuotationMessageQueueRepository(prod microservice.IProducer) QuotationMessageQueueRepository {
	mqKey := ""

	insRepo := QuotationMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.QuotationDoc](prod, config.QuotationMessageQueueConfig{}, "")
	return insRepo
}
