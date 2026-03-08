package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/pickandpack/config"
	"smlcloudplatform/internal/transaction/pickandpack/models"
	"smlcloudplatform/pkg/microservice"
)

type IPickandpackMessageQueueRepository interface {
	Create(doc models.PickandpackDoc) error
	Update(doc models.PickandpackDoc) error
	Delete(doc models.PickandpackDoc) error
	CreateInBatch(docList []models.PickandpackDoc) error
	UpdateInBatch(docList []models.PickandpackDoc) error
	DeleteInBatch(docList []models.PickandpackDoc) error
}

type PickandpackMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.PickandpackDoc]
}

func NewPickandpackMessageQueueRepository(prod microservice.IProducer) PickandpackMessageQueueRepository {
	mqKey := ""

	insRepo := PickandpackMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.PickandpackDoc](prod, config.PickandpackMessageQueueConfig{}, "")
	return insRepo
}
