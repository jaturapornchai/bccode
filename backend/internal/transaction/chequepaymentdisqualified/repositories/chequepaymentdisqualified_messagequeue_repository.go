package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/chequepaymentdisqualified/config"
	"smlcloudplatform/internal/transaction/chequepaymentdisqualified/models"
	"smlcloudplatform/pkg/microservice"
)

type IChequePaymentDisqualifiedMessageQueueRepository interface {
	Create(doc models.ChequePaymentDisqualifiedDoc) error
	Update(doc models.ChequePaymentDisqualifiedDoc) error
	Delete(doc models.ChequePaymentDisqualifiedDoc) error
	CreateInBatch(docList []models.ChequePaymentDisqualifiedDoc) error
	UpdateInBatch(docList []models.ChequePaymentDisqualifiedDoc) error
	DeleteInBatch(docList []models.ChequePaymentDisqualifiedDoc) error
}

type ChequePaymentDisqualifiedMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.ChequePaymentDisqualifiedDoc]
}

func NewChequePaymentDisqualifiedMessageQueueRepository(prod microservice.IProducer) ChequePaymentDisqualifiedMessageQueueRepository {
	mqKey := ""

	insRepo := ChequePaymentDisqualifiedMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.ChequePaymentDisqualifiedDoc](prod, config.ChequePaymentDisqualifiedMessageQueueConfig{}, "")
	return insRepo
}
