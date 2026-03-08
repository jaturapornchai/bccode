package repositories

import (
	"smlcloudplatform/internal/product/bom/config"
	"smlcloudplatform/internal/product/bom/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IBomMessageQueueRepository interface {
	Create(doc models.ProductBarcodeBOMViewDoc) error
	Update(doc models.ProductBarcodeBOMViewDoc) error
	Delete(doc models.ProductBarcodeBOMViewDoc) error
	CreateInBatch(docList []models.ProductBarcodeBOMViewDoc) error
	UpdateInBatch(docList []models.ProductBarcodeBOMViewDoc) error
	DeleteInBatch(docList []models.ProductBarcodeBOMViewDoc) error
}

type BomMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.ProductBarcodeBOMViewDoc]
}

func NewBomMessageQueueRepository(prod microservice.IProducer) BomMessageQueueRepository {
	mqKey := ""

	insRepo := BomMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.ProductBarcodeBOMViewDoc](prod, config.BomMessageQueueConfig{}, "")
	return insRepo
}
