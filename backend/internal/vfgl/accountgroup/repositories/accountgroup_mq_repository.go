package repositories

import (
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/vfgl/accountgroup/config"
	"smlcloudplatform/internal/vfgl/accountgroup/models"
	"smlcloudplatform/pkg/microservice"
)

type IAccountGroupMqRepository interface {
	Create(doc models.AccountGroupDoc) error
	Update(doc models.AccountGroupDoc) error
	Delete(doc common.Identity) error
	CreateInBatch(docList []models.AccountGroupDoc) error
}

type AccountGroupMqRepository struct {
	prod  microservice.IProducer
	mqKey string
	repositories.KafkaRepository[models.AccountGroupDoc]
}

func NewAccountGroupMqRepository(prod microservice.IProducer) AccountGroupMqRepository {
	mqKey := ""

	insRepo := AccountGroupMqRepository{
		prod:  prod,
		mqKey: mqKey,
	}
	insRepo.KafkaRepository = repositories.NewKafkaRepository[models.AccountGroupDoc](prod, config.AccountGroupMessageQueueConfig{}, "")
	return insRepo
}
