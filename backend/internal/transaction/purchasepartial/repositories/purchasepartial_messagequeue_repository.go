package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/purchasepartial/config"
	"smlcloudplatform/internal/transaction/purchasepartial/models"
	"smlcloudplatform/pkg/microservice"
)

type IPurchasepartialMessageQueueRepository interface {
	Create(doc models.PurchasepartialDoc) error
	Update(doc models.PurchasepartialDoc) error
	Delete(doc models.PurchasepartialDoc) error
	CreateInBatch(docList []models.PurchasepartialDoc) error
	UpdateInBatch(docList []models.PurchasepartialDoc) error
	DeleteInBatch(docList []models.PurchasepartialDoc) error
}

type PurchasepartialMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.PurchasepartialDoc]
}

func NewPurchasepartialMessageQueueRepository(prod microservice.IProducer) PurchasepartialMessageQueueRepository {
	mqKey := ""

	insRepo := PurchasepartialMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.PurchasepartialDoc](prod, config.PurchasepartialMessageQueueConfig{}, "")
	return insRepo
}
