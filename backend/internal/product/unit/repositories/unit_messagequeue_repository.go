package repositories

import (
	"smlcloudplatform/internal/product/unit/config"
	"smlcloudplatform/internal/product/unit/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IUnitMessageQueueRepository interface {
	Create(doc models.UnitDoc) error
	Update(doc models.UnitDoc) error
	Delete(doc models.UnitDoc) error
	CreateInBatch(docList []models.UnitDoc) error
	UpdateInBatch(docList []models.UnitDoc) error
	DeleteInBatch(docList []models.UnitDoc) error
}

type UnitMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.UnitDoc]
}

func NewUnitMessageQueueRepository(prod microservice.IProducer) UnitMessageQueueRepository {
	mqKey := ""

	insRepo := UnitMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.UnitDoc](prod, config.UnitMessageQueueConfig{}, "")
	return insRepo
}
