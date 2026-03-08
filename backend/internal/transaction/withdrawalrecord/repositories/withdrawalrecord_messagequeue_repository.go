package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/withdrawalrecord/config"
	"smlcloudplatform/internal/transaction/withdrawalrecord/models"
	"smlcloudplatform/pkg/microservice"
)

type IWithdrawalRecordMessageQueueRepository interface {
	Create(doc models.WithdrawalRecordDoc) error
	Update(doc models.WithdrawalRecordDoc) error
	Delete(doc models.WithdrawalRecordDoc) error
	CreateInBatch(docList []models.WithdrawalRecordDoc) error
	UpdateInBatch(docList []models.WithdrawalRecordDoc) error
	DeleteInBatch(docList []models.WithdrawalRecordDoc) error
}

type WithdrawalRecordMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.WithdrawalRecordDoc]
}

func NewWithdrawalRecordMessageQueueRepository(prod microservice.IProducer) WithdrawalRecordMessageQueueRepository {
	mqKey := ""

	insRepo := WithdrawalRecordMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.WithdrawalRecordDoc](prod, config.WithdrawalRecordMessageQueueConfig{}, "")
	return insRepo
}
