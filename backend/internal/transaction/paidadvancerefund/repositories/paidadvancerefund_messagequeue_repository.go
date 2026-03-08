package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/paidadvancerefund/config"
	"smlcloudplatform/internal/transaction/paidadvancerefund/models"
	"smlcloudplatform/pkg/microservice"
)

type IPaidAdvanceRefundMessageQueueRepository interface {
	Create(doc models.PaidAdvanceRefundDoc) error
	Update(doc models.PaidAdvanceRefundDoc) error
	Delete(doc models.PaidAdvanceRefundDoc) error
	CreateInBatch(docList []models.PaidAdvanceRefundDoc) error
	UpdateInBatch(docList []models.PaidAdvanceRefundDoc) error
	DeleteInBatch(docList []models.PaidAdvanceRefundDoc) error
}

type PaidAdvanceRefundMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.PaidAdvanceRefundDoc]
}

func NewPaidAdvanceRefundMessageQueueRepository(prod microservice.IProducer) PaidAdvanceRefundMessageQueueRepository {
	mqKey := ""

	insRepo := PaidAdvanceRefundMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.PaidAdvanceRefundDoc](prod, config.PaidAdvanceRefundMessageQueueConfig{}, "")
	return insRepo
}
