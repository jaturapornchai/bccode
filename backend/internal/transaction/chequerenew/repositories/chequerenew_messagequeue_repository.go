package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/chequerenew/config"
	"smlcloudplatform/internal/transaction/chequerenew/models"
	"smlcloudplatform/pkg/microservice"
)

type IChequeRenewMessageQueueRepository interface {
	Create(doc models.ChequeRenewDoc) error
	Update(doc models.ChequeRenewDoc) error
	Delete(doc models.ChequeRenewDoc) error
	CreateInBatch(docList []models.ChequeRenewDoc) error
	UpdateInBatch(docList []models.ChequeRenewDoc) error
	DeleteInBatch(docList []models.ChequeRenewDoc) error
}

type ChequeRenewMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.ChequeRenewDoc]
}

func NewChequeRenewMessageQueueRepository(prod microservice.IProducer) ChequeRenewMessageQueueRepository {
	mqKey := ""

	insRepo := ChequeRenewMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.ChequeRenewDoc](prod, config.ChequeRenewMessageQueueConfig{}, "")
	return insRepo
}
