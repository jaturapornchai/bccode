package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/chequedeposit/config"
	"smlcloudplatform/internal/transaction/chequedeposit/models"
	"smlcloudplatform/pkg/microservice"
)

type IChequeDepositMessageQueueRepository interface {
	Create(doc models.ChequeDepositDoc) error
	Update(doc models.ChequeDepositDoc) error
	Delete(doc models.ChequeDepositDoc) error
	CreateInBatch(docList []models.ChequeDepositDoc) error
	UpdateInBatch(docList []models.ChequeDepositDoc) error
	DeleteInBatch(docList []models.ChequeDepositDoc) error
}

type ChequeDepositMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.ChequeDepositDoc]
}

func NewChequeDepositMessageQueueRepository(prod microservice.IProducer) ChequeDepositMessageQueueRepository {
	mqKey := ""

	insRepo := ChequeDepositMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.ChequeDepositDoc](prod, config.ChequeDepositMessageQueueConfig{}, "")
	return insRepo
}
