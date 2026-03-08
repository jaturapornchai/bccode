package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/receivedepositrefund/config"
	"smlcloudplatform/internal/transaction/receivedepositrefund/models"
	"smlcloudplatform/pkg/microservice"
)

type IReceiveDepositRefundMessageQueueRepository interface {
	Create(doc models.ReceiveDepositRefundDoc) error
	Update(doc models.ReceiveDepositRefundDoc) error
	Delete(doc models.ReceiveDepositRefundDoc) error
	CreateInBatch(docList []models.ReceiveDepositRefundDoc) error
	UpdateInBatch(docList []models.ReceiveDepositRefundDoc) error
	DeleteInBatch(docList []models.ReceiveDepositRefundDoc) error
}

type ReceiveDepositRefundMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.ReceiveDepositRefundDoc]
}

func NewReceiveDepositRefundMessageQueueRepository(prod microservice.IProducer) ReceiveDepositRefundMessageQueueRepository {
	mqKey := ""

	insRepo := ReceiveDepositRefundMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.ReceiveDepositRefundDoc](prod, config.ReceiveDepositRefundMessageQueueConfig{}, "")
	return insRepo
}
