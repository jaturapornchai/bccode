package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/saledebitnote/config"
	"smlcloudplatform/internal/transaction/saledebitnote/models"
	"smlcloudplatform/pkg/microservice"
)

type ISaleDebitNoteMessageQueueRepository interface {
	Create(doc models.SaleDebitNoteDoc) error
	Update(doc models.SaleDebitNoteDoc) error
	Delete(doc models.SaleDebitNoteDoc) error
	CreateInBatch(docList []models.SaleDebitNoteDoc) error
	UpdateInBatch(docList []models.SaleDebitNoteDoc) error
	DeleteInBatch(docList []models.SaleDebitNoteDoc) error
}

type SaleDebitNoteMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.SaleDebitNoteDoc]
}

func NewSaleDebitNoteMessageQueueRepository(prod microservice.IProducer) SaleDebitNoteMessageQueueRepository {
	mqKey := ""

	insRepo := SaleDebitNoteMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.SaleDebitNoteDoc](prod, config.SaleDebitNoteMessageQueueConfig{}, "")
	return insRepo
}
