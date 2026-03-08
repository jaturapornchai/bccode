package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/chequepaymentdeposit/config"
	"smlcloudplatform/internal/transaction/chequepaymentdeposit/models"
	"smlcloudplatform/pkg/microservice"
)

type IChequePaymentDepositMessageQueueRepository interface {
	Create(doc models.ChequePaymentDepositDoc) error
	Update(doc models.ChequePaymentDepositDoc) error
	Delete(doc models.ChequePaymentDepositDoc) error
	CreateInBatch(docList []models.ChequePaymentDepositDoc) error
	UpdateInBatch(docList []models.ChequePaymentDepositDoc) error
	DeleteInBatch(docList []models.ChequePaymentDepositDoc) error
}

type ChequePaymentDepositMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.ChequePaymentDepositDoc]
}

func NewChequePaymentDepositMessageQueueRepository(prod microservice.IProducer) ChequePaymentDepositMessageQueueRepository {
	mqKey := ""

	insRepo := ChequePaymentDepositMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.ChequePaymentDepositDoc](prod, config.ChequePaymentDepositMessageQueueConfig{}, "")
	return insRepo
}
