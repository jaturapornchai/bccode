package repositories

import (
	"smlcloudplatform/internal/organization/costcenter/config"
	"smlcloudplatform/internal/organization/costcenter/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
)

type ICostCenterMessageQueueRepository interface {
	Create(doc models.CostCenterDoc) error
	Update(doc models.CostCenterDoc) error
	Delete(doc models.CostCenterDoc) error
	CreateInBatch(docList []models.CostCenterDoc) error
	UpdateInBatch(docList []models.CostCenterDoc) error
	DeleteInBatch(docList []models.CostCenterDoc) error
}

type CostCenterMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.CostCenterDoc]
}

func NewCostCenterMessageQueueRepository(prod microservice.IProducer) CostCenterMessageQueueRepository {
	insRepo := CostCenterMessageQueueRepository{prod: prod, mqKey: ""}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.CostCenterDoc](prod, config.CostCenterMessageQueueConfig{}, "")
	return insRepo
}
