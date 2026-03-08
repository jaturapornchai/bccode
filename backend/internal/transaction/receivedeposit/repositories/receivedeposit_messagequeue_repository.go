package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/receivedeposit/config"
	"smlcloudplatform/internal/transaction/receivedeposit/models"
	"smlcloudplatform/pkg/microservice"
)

type IReceiveDepositMessageQueueRepository interface {
	Create(doc models.ReceiveDepositDoc) error
	Update(doc models.ReceiveDepositDoc) error
	Delete(doc models.ReceiveDepositDoc) error
	CreateInBatch(docList []models.ReceiveDepositDoc) error
	UpdateInBatch(docList []models.ReceiveDepositDoc) error
	DeleteInBatch(docList []models.ReceiveDepositDoc) error
}

type ReceiveDepositMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.ReceiveDepositDoc]
}

func NewReceiveDepositMessageQueueRepository(prod microservice.IProducer) ReceiveDepositMessageQueueRepository {
	mqKey := ""

	insRepo := ReceiveDepositMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.ReceiveDepositDoc](prod, config.ReceiveDepositMessageQueueConfig{}, "")
	return insRepo
}
