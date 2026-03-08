package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/advancepayment/config"
	"smlcloudplatform/internal/transaction/advancepayment/models"
	"smlcloudplatform/pkg/microservice"
)

type IAdvancePaymentMessageQueueRepository interface {
	Create(doc models.AdvancePaymentDoc) error
	Update(doc models.AdvancePaymentDoc) error
	Delete(doc models.AdvancePaymentDoc) error
	CreateInBatch(docList []models.AdvancePaymentDoc) error
	UpdateInBatch(docList []models.AdvancePaymentDoc) error
	DeleteInBatch(docList []models.AdvancePaymentDoc) error
}

type AdvancePaymentMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.AdvancePaymentDoc]
}

func NewAdvancePaymentMessageQueueRepository(prod microservice.IProducer) AdvancePaymentMessageQueueRepository {
	mqKey := ""

	insRepo := AdvancePaymentMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.AdvancePaymentDoc](prod, config.AdvancePaymentMessageQueueConfig{}, "")
	return insRepo
}
