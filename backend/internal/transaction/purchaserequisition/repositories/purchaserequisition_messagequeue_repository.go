package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/purchaserequisition/config"
	"smlcloudplatform/internal/transaction/purchaserequisition/models"
	"smlcloudplatform/pkg/microservice"
)

type IPurchaseRequisitionMessageQueueRepository interface {
	Create(doc models.PurchaseRequisitionDoc) error
	Update(doc models.PurchaseRequisitionDoc) error
	Delete(doc models.PurchaseRequisitionDoc) error
	CreateInBatch(docList []models.PurchaseRequisitionDoc) error
	UpdateInBatch(docList []models.PurchaseRequisitionDoc) error
	DeleteInBatch(docList []models.PurchaseRequisitionDoc) error
}

type PurchaseRequisitionMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.PurchaseRequisitionDoc]
}

func NewPurchaseRequisitionMessageQueueRepository(prod microservice.IProducer) PurchaseRequisitionMessageQueueRepository {
	mqKey := ""
	insRepo := PurchaseRequisitionMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.PurchaseRequisitionDoc](prod, config.PurchaseRequisitionMessageQueueConfig{}, "")
	return insRepo
}
