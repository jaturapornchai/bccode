package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/purchaseorder/config"
	"smlcloudplatform/internal/transaction/purchaseorder/models"
	"smlcloudplatform/pkg/microservice"
)

type IPurchaseOrderMessageQueueRepository interface {
	Create(doc models.PurchaseOrderDoc) error
	Update(doc models.PurchaseOrderDoc) error
	Delete(doc models.PurchaseOrderDoc) error
	CreateInBatch(docList []models.PurchaseOrderDoc) error
	UpdateInBatch(docList []models.PurchaseOrderDoc) error
	DeleteInBatch(docList []models.PurchaseOrderDoc) error
}

type PurchaseOrderMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.PurchaseOrderDoc]
}

func NewPurchaseOrderMessageQueueRepository(prod microservice.IProducer) PurchaseOrderMessageQueueRepository {
	mqKey := ""

	insRepo := PurchaseOrderMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.PurchaseOrderDoc](prod, config.PurchaseOrderMessageQueueConfig{}, "")
	return insRepo
}
