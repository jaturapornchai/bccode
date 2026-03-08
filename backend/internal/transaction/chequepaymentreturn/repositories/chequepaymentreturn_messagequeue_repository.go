package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/chequepaymentreturn/config"
	"smlcloudplatform/internal/transaction/chequepaymentreturn/models"
	"smlcloudplatform/pkg/microservice"
)

type IChequePaymentReturnMessageQueueRepository interface {
	Create(doc models.ChequePaymentReturnDoc) error
	Update(doc models.ChequePaymentReturnDoc) error
	Delete(doc models.ChequePaymentReturnDoc) error
	CreateInBatch(docList []models.ChequePaymentReturnDoc) error
	UpdateInBatch(docList []models.ChequePaymentReturnDoc) error
	DeleteInBatch(docList []models.ChequePaymentReturnDoc) error
}

type ChequePaymentReturnMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.ChequePaymentReturnDoc]
}

func NewChequePaymentReturnMessageQueueRepository(prod microservice.IProducer) ChequePaymentReturnMessageQueueRepository {
	mqKey := ""

	insRepo := ChequePaymentReturnMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.ChequePaymentReturnDoc](prod, config.ChequePaymentReturnMessageQueueConfig{}, "")
	return insRepo
}
