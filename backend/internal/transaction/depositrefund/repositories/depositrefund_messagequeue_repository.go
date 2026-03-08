package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/depositrefund/config"
	"smlcloudplatform/internal/transaction/depositrefund/models"
	"smlcloudplatform/pkg/microservice"
)

type IDepositRefundMessageQueueRepository interface {
	Create(doc models.DepositRefundDoc) error
	Update(doc models.DepositRefundDoc) error
	Delete(doc models.DepositRefundDoc) error
	CreateInBatch(docList []models.DepositRefundDoc) error
	UpdateInBatch(docList []models.DepositRefundDoc) error
	DeleteInBatch(docList []models.DepositRefundDoc) error
}

type DepositRefundMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.DepositRefundDoc]
}

func NewDepositRefundMessageQueueRepository(prod microservice.IProducer) DepositRefundMessageQueueRepository {
	mqKey := ""

	insRepo := DepositRefundMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.DepositRefundDoc](prod, config.DepositRefundMessageQueueConfig{}, "")
	return insRepo
}
