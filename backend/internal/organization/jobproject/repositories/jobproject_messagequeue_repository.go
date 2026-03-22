package repositories

import (
	"smlcloudplatform/internal/organization/jobproject/config"
	"smlcloudplatform/internal/organization/jobproject/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IJobProjectMessageQueueRepository interface {
	Create(doc models.JobProjectDoc) error
	Update(doc models.JobProjectDoc) error
	Delete(doc models.JobProjectDoc) error
	CreateInBatch(docList []models.JobProjectDoc) error
	UpdateInBatch(docList []models.JobProjectDoc) error
	DeleteInBatch(docList []models.JobProjectDoc) error
}

type JobProjectMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.JobProjectDoc]
}

func NewJobProjectMessageQueueRepository(prod microservice.IProducer) JobProjectMessageQueueRepository {
	insRepo := JobProjectMessageQueueRepository{prod: prod, mqKey: ""}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.JobProjectDoc](prod, config.JobProjectMessageQueueConfig{}, "")
	return insRepo
}
