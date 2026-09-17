package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/billingnote/config"
	"smlcloudplatform/internal/transaction/billingnote/models"
	"smlcloudplatform/pkg/microservice"
)

type IBillingNoteMessageQueueRepository interface {
	Create(doc models.BillingNoteDoc) error
	Update(doc models.BillingNoteDoc) error
	Delete(doc models.BillingNoteDoc) error
	CreateInBatch(docList []models.BillingNoteDoc) error
	UpdateInBatch(docList []models.BillingNoteDoc) error
	DeleteInBatch(docList []models.BillingNoteDoc) error
}

type BillingNoteMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.BillingNoteDoc]
}

func NewBillingNoteMessageQueueRepository(prod microservice.IProducer) IBillingNoteMessageQueueRepository {
	mqKey := ""

	insRepo := BillingNoteMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.BillingNoteDoc](prod, config.MessageQueueConfig{}, "")
	return insRepo
}
