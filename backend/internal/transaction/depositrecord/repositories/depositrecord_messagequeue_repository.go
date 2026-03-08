package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/depositrecord/config"
	"smlcloudplatform/internal/transaction/depositrecord/models"
	"smlcloudplatform/pkg/microservice"
)

type IDepositRecordMessageQueueRepository interface {
	Create(doc models.DepositRecordDoc) error
	Update(doc models.DepositRecordDoc) error
	Delete(doc models.DepositRecordDoc) error
	CreateInBatch(docList []models.DepositRecordDoc) error
	UpdateInBatch(docList []models.DepositRecordDoc) error
	DeleteInBatch(docList []models.DepositRecordDoc) error
}

type DepositRecordMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.DepositRecordDoc]
}

func NewDepositRecordMessageQueueRepository(prod microservice.IProducer) DepositRecordMessageQueueRepository {
	mqKey := ""

	insRepo := DepositRecordMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.DepositRecordDoc](prod, config.DepositRecordMessageQueueConfig{}, "")
	return insRepo
}
