package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/banktransferrecord/config"
	"smlcloudplatform/internal/transaction/banktransferrecord/models"
	"smlcloudplatform/pkg/microservice"
)

type IBankTransferRecordMessageQueueRepository interface {
	Create(doc models.BankTransferRecordDoc) error
	Update(doc models.BankTransferRecordDoc) error
	Delete(doc models.BankTransferRecordDoc) error
	CreateInBatch(docList []models.BankTransferRecordDoc) error
	UpdateInBatch(docList []models.BankTransferRecordDoc) error
	DeleteInBatch(docList []models.BankTransferRecordDoc) error
}

type BankTransferRecordMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.BankTransferRecordDoc]
}

func NewBankTransferRecordMessageQueueRepository(prod microservice.IProducer) BankTransferRecordMessageQueueRepository {
	mqKey := ""

	insRepo := BankTransferRecordMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.BankTransferRecordDoc](prod, config.BankTransferRecordMessageQueueConfig{}, "")
	return insRepo
}
