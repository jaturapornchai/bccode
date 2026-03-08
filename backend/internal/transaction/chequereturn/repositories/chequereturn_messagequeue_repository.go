package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/chequereturn/config"
	"smlcloudplatform/internal/transaction/chequereturn/models"
	"smlcloudplatform/pkg/microservice"
)

type IChequeReturnMessageQueueRepository interface {
	Create(doc models.ChequeReturnDoc) error
	Update(doc models.ChequeReturnDoc) error
	Delete(doc models.ChequeReturnDoc) error
	CreateInBatch(docList []models.ChequeReturnDoc) error
	UpdateInBatch(docList []models.ChequeReturnDoc) error
	DeleteInBatch(docList []models.ChequeReturnDoc) error
}

type ChequeReturnMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.ChequeReturnDoc]
}

func NewChequeReturnMessageQueueRepository(prod microservice.IProducer) ChequeReturnMessageQueueRepository {
	mqKey := ""

	insRepo := ChequeReturnMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.ChequeReturnDoc](prod, config.ChequeReturnMessageQueueConfig{}, "")
	return insRepo
}
