package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/advancepaymentrefund/config"
	"smlcloudplatform/internal/transaction/advancepaymentrefund/models"
	"smlcloudplatform/pkg/microservice"
)

type IAdvancePaymentRefundMessageQueueRepository interface {
	Create(doc models.AdvancePaymentRefundDoc) error
	Update(doc models.AdvancePaymentRefundDoc) error
	Delete(doc models.AdvancePaymentRefundDoc) error
	CreateInBatch(docList []models.AdvancePaymentRefundDoc) error
	UpdateInBatch(docList []models.AdvancePaymentRefundDoc) error
	DeleteInBatch(docList []models.AdvancePaymentRefundDoc) error
}

type AdvancePaymentRefundMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.AdvancePaymentRefundDoc]
}

func NewAdvancePaymentRefundMessageQueueRepository(prod microservice.IProducer) AdvancePaymentRefundMessageQueueRepository {
	mqKey := ""

	insRepo := AdvancePaymentRefundMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.AdvancePaymentRefundDoc](prod, config.AdvancePaymentRefundMessageQueueConfig{}, "")
	return insRepo
}
