package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/chequechange/config"
	"smlcloudplatform/internal/transaction/chequechange/models"
	"smlcloudplatform/pkg/microservice"
)

type IChequeChangeMessageQueueRepository interface {
	Create(doc models.ChequeChangeDoc) error
	Update(doc models.ChequeChangeDoc) error
	Delete(doc models.ChequeChangeDoc) error
	CreateInBatch(docList []models.ChequeChangeDoc) error
	UpdateInBatch(docList []models.ChequeChangeDoc) error
	DeleteInBatch(docList []models.ChequeChangeDoc) error
}

type ChequeChangeMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.ChequeChangeDoc]
}

func NewChequeChangeMessageQueueRepository(prod microservice.IProducer) ChequeChangeMessageQueueRepository {
	mqKey := ""

	insRepo := ChequeChangeMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.ChequeChangeDoc](prod, config.ChequeChangeMessageQueueConfig{}, "")
	return insRepo
}
