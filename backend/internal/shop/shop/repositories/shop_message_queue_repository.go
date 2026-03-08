package repositories

import (
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/shop/models"
	"smlcloudplatform/internal/shop/shop/config"
	"smlcloudplatform/pkg/microservice"
)

type ICreditorPaymentMessageQueueRepository interface {
	Create(doc models.ShopDoc) error
}

type ShopMessageQueueRepository struct {
	prod     microservice.IProducer
	mqKey    string
	mqConfig config.ShopMessageQueueConfig
	repositories.KafkaRepository[models.ShopDoc]
}

func NewShopMessageQueueRepository(prod microservice.IProducer, mqConfig config.ShopMessageQueueConfig) ICreditorPaymentMessageQueueRepository {
	mqKey := ""

	insRepo := ShopMessageQueueRepository{
		prod:     prod,
		mqKey:    mqKey,
		mqConfig: mqConfig,
	}
	return insRepo
}

func (repo ShopMessageQueueRepository) Create(doc models.ShopDoc) error {
	err := repo.prod.SendMessage(repo.mqConfig.TopicCreated(), repo.mqKey, doc)

	if err != nil {
		return err
	}

	return nil
}
