package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/stockpickupproduct/config"
	"smlcloudplatform/internal/transaction/stockpickupproduct/models"
	"smlcloudplatform/pkg/microservice"
)

type IStockPickupProductMessageQueueRepository interface {
	Create(doc models.StockPickupProductDoc) error
	Update(doc models.StockPickupProductDoc) error
	Delete(doc models.StockPickupProductDoc) error
	CreateInBatch(docList []models.StockPickupProductDoc) error
	UpdateInBatch(docList []models.StockPickupProductDoc) error
	DeleteInBatch(docList []models.StockPickupProductDoc) error
}

type StockPickupProductMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.StockPickupProductDoc]
}

func NewStockPickupProductMessageQueueRepository(prod microservice.IProducer) StockPickupProductMessageQueueRepository {
	mqKey := ""

	insRepo := StockPickupProductMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.StockPickupProductDoc](prod, config.StockPickupProductMessageQueueConfig{}, "")
	return insRepo
}
