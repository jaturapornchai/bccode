package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/rfq/config"
	"smlcloudplatform/internal/transaction/rfq/models"
	"smlcloudplatform/pkg/microservice"
)

type IRFQMessageQueueRepository interface {
	Create(doc models.RFQDoc) error
	Update(doc models.RFQDoc) error
	Delete(doc models.RFQDoc) error
	CreateInBatch(docList []models.RFQDoc) error
	UpdateInBatch(docList []models.RFQDoc) error
	DeleteInBatch(docList []models.RFQDoc) error
}

type RFQMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.RFQDoc]
}

func NewRFQMessageQueueRepository(prod microservice.IProducer) RFQMessageQueueRepository {
	mqKey := ""
	insRepo := RFQMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.RFQDoc](prod, config.RFQMessageQueueConfig{}, "")
	return insRepo
}
