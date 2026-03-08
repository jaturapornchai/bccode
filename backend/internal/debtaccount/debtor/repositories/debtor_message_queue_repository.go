package repositories

import (
	"smlcloudplatform/internal/debtaccount/debtor/config"
	"smlcloudplatform/internal/debtaccount/debtor/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IDebtorMessageQueueRepository interface {
	Create(doc models.DebtorDoc) error
	Update(doc models.DebtorDoc) error
	Delete(doc models.DebtorDoc) error
	CreateInBatch(docList []models.DebtorDoc) error
	UpdateInBatch(docList []models.DebtorDoc) error
	DeleteInBatch(docList []models.DebtorDoc) error
}

type DebtorMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.DebtorDoc]
}

func NewDebtorMessageQueueRepository(producer microservice.IProducer) DebtorMessageQueueRepository {
	mqKey := ""

	insRepo := DebtorMessageQueueRepository{
		prod:  producer,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.DebtorDoc](producer, config.DebtorMessageQueueConfig{}, "")
	return insRepo
}
