package repositories

import (
	"smlcloudplatform/internal/product/productgroup/config"
	"smlcloudplatform/internal/product/productgroup/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IProductGroupMessageQueueRepository interface {
	Create(doc models.ProductGroupDoc) error
	Update(doc models.ProductGroupDoc) error
	Delete(doc models.ProductGroupDoc) error
	CreateInBatch(docList []models.ProductGroupDoc) error
	UpdateInBatch(docList []models.ProductGroupDoc) error
	DeleteInBatch(docList []models.ProductGroupDoc) error
}

type ProductGroupMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.ProductGroupDoc]
}

func NewProductGroupMessageQueueRepository(prod microservice.IProducer) ProductGroupMessageQueueRepository {
	mqKey := ""

	insRepo := ProductGroupMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.ProductGroupDoc](prod, config.ProductGroupMessageQueueConfig{}, "")
	return insRepo
}
