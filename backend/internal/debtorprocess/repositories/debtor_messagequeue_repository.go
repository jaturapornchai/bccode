package repositories

import (
	"smlcloudplatform/internal/debtorprocess/config"
	"smlcloudplatform/internal/debtorprocess/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IDebtorProcessMessageQueueRepository interface {
	Create(doc models.DebtorProcessRequest) error
	Update(doc models.DebtorProcessRequest) error
	Delete(doc models.DebtorProcessRequest) error
	CreateInBatch(docList []models.DebtorProcessRequest) error
	UpdateInBatch(docList []models.DebtorProcessRequest) error
	DeleteInBatch(docList []models.DebtorProcessRequest) error
}

type DebtorProcessMessageQueueRepository struct {
	producer microservice.IProducer
	mqKey    string
	repositories.KafkaRepository[models.DebtorProcessRequest]
}

func NewDebtorProcessMessageQueueRepository(producer microservice.IProducer) IDebtorProcessMessageQueueRepository {
	mqKey := ""

	insRepo := DebtorProcessMessageQueueRepository{
		producer: producer,
		mqKey:    mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.DebtorProcessRequest](producer, config.DebtorProcessMessageQueueConfig{}, "")
	return insRepo
}
