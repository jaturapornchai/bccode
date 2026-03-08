package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/stocktransfer/config"
	"smlcloudplatform/internal/transaction/stocktransfer/models"
	"smlcloudplatform/pkg/microservice"
)

type IStockTransferMessageQueueRepository interface {
	Create(doc models.StockTransferDoc) error
	Update(doc models.StockTransferDoc) error
	Delete(doc models.StockTransferDoc) error
	CreateInBatch(docList []models.StockTransferDoc) error
	UpdateInBatch(docList []models.StockTransferDoc) error
	DeleteInBatch(docList []models.StockTransferDoc) error
}

type StockTransferMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.StockTransferDoc]
}

func NewStockTransferMessageQueueRepository(prod microservice.IProducer) StockTransferMessageQueueRepository {
	mqKey := ""

	insRepo := StockTransferMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.StockTransferDoc](prod, config.StockTransferMessageQueueConfig{}, "")
	return insRepo
}
