package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/chequepaymentchange/config"
	"smlcloudplatform/internal/transaction/chequepaymentchange/models"
	"smlcloudplatform/pkg/microservice"
)

type IChequePaymentChangeMessageQueueRepository interface {
	Create(doc models.ChequePaymentChangeDoc) error
	Update(doc models.ChequePaymentChangeDoc) error
	Delete(doc models.ChequePaymentChangeDoc) error
	CreateInBatch(docList []models.ChequePaymentChangeDoc) error
	UpdateInBatch(docList []models.ChequePaymentChangeDoc) error
	DeleteInBatch(docList []models.ChequePaymentChangeDoc) error
}

type ChequePaymentChangeMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.ChequePaymentChangeDoc]
}

func NewChequePaymentChangeMessageQueueRepository(prod microservice.IProducer) ChequePaymentChangeMessageQueueRepository {
	mqKey := ""

	insRepo := ChequePaymentChangeMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.ChequePaymentChangeDoc](prod, config.ChequePaymentChangeMessageQueueConfig{}, "")
	return insRepo
}
