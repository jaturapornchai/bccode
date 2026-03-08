package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/purchasedebitnote/config"
	"smlcloudplatform/internal/transaction/purchasedebitnote/models"
	"smlcloudplatform/pkg/microservice"
)

type IPurchaseDebitNoteMessageQueueRepository interface {
	Create(doc models.PurchaseDebitNoteDoc) error
	Update(doc models.PurchaseDebitNoteDoc) error
	Delete(doc models.PurchaseDebitNoteDoc) error
	CreateInBatch(docList []models.PurchaseDebitNoteDoc) error
	UpdateInBatch(docList []models.PurchaseDebitNoteDoc) error
	DeleteInBatch(docList []models.PurchaseDebitNoteDoc) error
}

type PurchaseDebitNoteMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.PurchaseDebitNoteDoc]
}

func NewPurchaseDebitNoteMessageQueueRepository(prod microservice.IProducer) PurchaseDebitNoteMessageQueueRepository {
	mqKey := ""

	insRepo := PurchaseDebitNoteMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.PurchaseDebitNoteDoc](prod, config.PurchaseDebitNoteMessageQueueConfig{}, "")
	return insRepo
}
