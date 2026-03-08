package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/paidadvance/config"
	"smlcloudplatform/internal/transaction/paidadvance/models"
	"smlcloudplatform/pkg/microservice"
)

type IPaidAdvanceMessageQueueRepository interface {
	Create(doc models.PaidAdvanceDoc) error
	Update(doc models.PaidAdvanceDoc) error
	Delete(doc models.PaidAdvanceDoc) error
	CreateInBatch(docList []models.PaidAdvanceDoc) error
	UpdateInBatch(docList []models.PaidAdvanceDoc) error
	DeleteInBatch(docList []models.PaidAdvanceDoc) error
}

type PaidAdvanceMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.PaidAdvanceDoc]
}

func NewPaidAdvanceMessageQueueRepository(prod microservice.IProducer) PaidAdvanceMessageQueueRepository {
	mqKey := ""

	insRepo := PaidAdvanceMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.PaidAdvanceDoc](prod, config.PaidAdvanceMessageQueueConfig{}, "")
	return insRepo
}
