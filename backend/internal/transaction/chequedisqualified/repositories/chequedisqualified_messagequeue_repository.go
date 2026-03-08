package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/chequedisqualified/config"
	"smlcloudplatform/internal/transaction/chequedisqualified/models"
	"smlcloudplatform/pkg/microservice"
)

type IChequeDisqualifiedMessageQueueRepository interface {
	Create(doc models.ChequeDisqualifiedDoc) error
	Update(doc models.ChequeDisqualifiedDoc) error
	Delete(doc models.ChequeDisqualifiedDoc) error
	CreateInBatch(docList []models.ChequeDisqualifiedDoc) error
	UpdateInBatch(docList []models.ChequeDisqualifiedDoc) error
	DeleteInBatch(docList []models.ChequeDisqualifiedDoc) error
}

type ChequeDisqualifiedMessageQueueRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.ChequeDisqualifiedDoc]
}

func NewChequeDisqualifiedMessageQueueRepository(prod microservice.IProducer) ChequeDisqualifiedMessageQueueRepository {
	mqKey := ""

	insRepo := ChequeDisqualifiedMessageQueueRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.ChequeDisqualifiedDoc](prod, config.ChequeDisqualifiedMessageQueueConfig{}, "")
	return insRepo
}
