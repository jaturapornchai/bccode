package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/accrualreceive/config"
	"smlcloudplatform/internal/transaction/accrualreceive/models"
	"smlcloudplatform/pkg/microservice"
)

type IAccrualreceiveMessageQueueRepository interface {
	Create(doc models.AccrualreceiveDoc) error
	Update(doc models.AccrualreceiveDoc) error
	Delete(doc models.AccrualreceiveDoc) error
	CreateInBatch(docList []models.AccrualreceiveDoc) error
	UpdateInBatch(docList []models.AccrualreceiveDoc) error
	DeleteInBatch(docList []models.AccrualreceiveDoc) error
}

type AccrualreceiveMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.AccrualreceiveDoc]
}

func NewAccrualreceiveMessageQueueRepository(prod microservice.IProducer) AccrualreceiveMessageQueueRepository {
	mqKey := ""

	insRepo := AccrualreceiveMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.AccrualreceiveDoc](prod, config.AccrualreceiveMessageQueueConfig{}, "")
	return insRepo
}
