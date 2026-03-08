package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/creditcardwithdrawal/config"
	"smlcloudplatform/internal/transaction/creditcardwithdrawal/models"
	"smlcloudplatform/pkg/microservice"
)

type ICreditCardWithdrawalMessageQueueRepository interface {
	Create(doc models.CreditCardWithdrawalDoc) error
	Update(doc models.CreditCardWithdrawalDoc) error
	Delete(doc models.CreditCardWithdrawalDoc) error
	CreateInBatch(docList []models.CreditCardWithdrawalDoc) error
	UpdateInBatch(docList []models.CreditCardWithdrawalDoc) error
	DeleteInBatch(docList []models.CreditCardWithdrawalDoc) error
}

type CreditCardWithdrawalMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.CreditCardWithdrawalDoc]
}

func NewCreditCardWithdrawalMessageQueueRepository(prod microservice.IProducer) CreditCardWithdrawalMessageQueueRepository {
	mqKey := ""

	insRepo := CreditCardWithdrawalMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.CreditCardWithdrawalDoc](prod, config.CreditCardWithdrawalMessageQueueConfig{}, "")
	return insRepo
}
