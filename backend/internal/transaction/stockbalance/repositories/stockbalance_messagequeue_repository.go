package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/stockbalance/config"
	"smlcloudplatform/internal/transaction/stockbalance/models"
	"smlcloudplatform/pkg/microservice"
)

type IStockBalanceMessageQueueRepository interface {
	Create(doc models.StockBalanceMessage) error
	Update(doc models.StockBalanceMessage) error
	Delete(doc models.StockBalanceMessage) error
	CreateInBatch(docList []models.StockBalanceMessage) error
	UpdateInBatch(docList []models.StockBalanceMessage) error
	DeleteInBatch(docList []models.StockBalanceMessage) error
}

type StockBalanceMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.StockBalanceMessage]
}

func NewStockBalanceMessageQueueRepository(prod microservice.IProducer) StockBalanceMessageQueueRepository {
	mqKey := ""

	insRepo := StockBalanceMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.StockBalanceMessage](prod, config.StockBalanceMessageQueueConfig{}, "")
	return insRepo
}
