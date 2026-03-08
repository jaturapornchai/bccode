package repositories

import (
	"smlcloudplatform/internal/debtaccount/creditor/config"
	"smlcloudplatform/internal/debtaccount/creditor/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
)

type ICreditorMessageQueueRepository interface {
	Create(doc models.CreditorDoc) error
	Update(doc models.CreditorDoc) error
	Delete(doc models.CreditorDoc) error
	CreateInBatch(docList []models.CreditorDoc) error
	UpdateInBatch(docList []models.CreditorDoc) error
	DeleteInBatch(docList []models.CreditorDoc) error
}

type CreditorMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.CreditorDoc]
}

func NewCreditorMessageQueueRepository(prod microservice.IProducer) CreditorMessageQueueRepository {
	mqKey := ""

	insRepo := CreditorMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.CreditorDoc](prod, config.CreditorMessageQueueConfig{}, "")
	return insRepo
}
