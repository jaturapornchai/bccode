package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/saleorder/config"
	"smlcloudplatform/internal/transaction/saleorder/models"
	"smlcloudplatform/pkg/microservice"
)

type ISaleOrderMessageQueueRepository interface {
	Create(doc models.SaleOrderDoc) error
	Update(doc models.SaleOrderDoc) error
	Delete(doc models.SaleOrderDoc) error
	CreateInBatch(docList []models.SaleOrderDoc) error
	UpdateInBatch(docList []models.SaleOrderDoc) error
	DeleteInBatch(docList []models.SaleOrderDoc) error
}

type SaleOrderMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.SaleOrderDoc]
}

func NewSaleOrderMessageQueueRepository(prod microservice.IProducer) SaleOrderMessageQueueRepository {
	mqKey := ""

	insRepo := SaleOrderMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.SaleOrderDoc](prod, config.SaleOrderMessageQueueConfig{}, "")
	return insRepo
}
