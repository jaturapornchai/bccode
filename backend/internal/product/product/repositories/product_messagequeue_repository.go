package repositories

import (
	"smlcloudplatform/internal/product/product/config"
	"smlcloudplatform/internal/product/product/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IProductMessageQueueRepository interface {
	Create(doc models.ProductDoc) error
	Update(doc models.ProductDoc) error
	Delete(doc models.ProductDoc) error
}

type ProductMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.ProductDoc]
}

func NewProductMessageQueueRepository(prod microservice.IProducer) ProductMessageQueueRepository {
	mqKey := ""

	insRepo := ProductMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.ProductDoc](prod, config.ProductMessageQueueConfig{}, "")
	return insRepo
}

func (repo ProductMessageQueueRepository) Create(doc models.ProductDoc) error {
	doc.Barcodes = nil
	return repo.KafkaRepository.Create(doc)
}

func (repo ProductMessageQueueRepository) Update(doc models.ProductDoc) error {
	doc.Barcodes = nil
	return repo.KafkaRepository.Update(doc)
}

func (repo ProductMessageQueueRepository) Delete(doc models.ProductDoc) error {
	doc.Barcodes = nil
	return repo.KafkaRepository.Delete(doc)
}
