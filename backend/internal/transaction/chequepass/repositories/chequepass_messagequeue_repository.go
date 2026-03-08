package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/chequepass/config"
	"smlcloudplatform/internal/transaction/chequepass/models"
	"smlcloudplatform/pkg/microservice"
)

type IChequePassMessageQueueRepository interface {
	Create(doc models.ChequePassDoc) error
	Update(doc models.ChequePassDoc) error
	Delete(doc models.ChequePassDoc) error
	CreateInBatch(docList []models.ChequePassDoc) error
	UpdateInBatch(docList []models.ChequePassDoc) error
	DeleteInBatch(docList []models.ChequePassDoc) error
}

type ChequePassMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.ChequePassDoc]
}

func NewChequePassMessageQueueRepository(prod microservice.IProducer) ChequePassMessageQueueRepository {
	mqKey := ""

	insRepo := ChequePassMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.ChequePassDoc](prod, config.ChequePassMessageQueueConfig{}, "")
	return insRepo
}
