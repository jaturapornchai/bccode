package repositories

import (
	"smlcloudplatform/internal/product/ordertype/config"
	"smlcloudplatform/internal/product/ordertype/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IOrderTypeMessageQueueRepository interface {
	Create(doc models.OrderTypeDoc) error
	Update(doc models.OrderTypeDoc) error
	Delete(doc models.OrderTypeDoc) error
	CreateInBatch(docList []models.OrderTypeDoc) error
	UpdateInBatch(docList []models.OrderTypeDoc) error
	DeleteInBatch(docList []models.OrderTypeDoc) error
}

type OrderTypeMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.OrderTypeDoc]
}

func NewOrderTypeMessageQueueRepository(prod microservice.IProducer) OrderTypeMessageQueueRepository {
	mqKey := ""

	insRepo := OrderTypeMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.OrderTypeDoc](prod, config.OrderTypeMessageQueueConfig{}, "")
	return insRepo
}
