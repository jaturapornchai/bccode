package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/deposit/config"
	"smlcloudplatform/internal/transaction/deposit/models"
	"smlcloudplatform/pkg/microservice"
)

type IDepositMessageQueueRepository interface {
	Create(doc models.DepositDoc) error
	Update(doc models.DepositDoc) error
	Delete(doc models.DepositDoc) error
	CreateInBatch(docList []models.DepositDoc) error
	UpdateInBatch(docList []models.DepositDoc) error
	DeleteInBatch(docList []models.DepositDoc) error
}

type DepositMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.DepositDoc]
}

func NewDepositMessageQueueRepository(prod microservice.IProducer) DepositMessageQueueRepository {
	mqKey := ""

	insRepo := DepositMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.DepositDoc](prod, config.DepositMessageQueueConfig{}, "")
	return insRepo
}
