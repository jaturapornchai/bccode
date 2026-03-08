package purchase

import (
	"context"
	purchaseRepository "smlcloudplatform/internal/transaction/purchase/repositories"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IPurchaseTransactionAdminService interface {
	ReSyncPurchaseDoc(shopID string) error
	ReSyncPurchaseDeleteDoc(shopID string) error
}

type PurchaseTransactionAdminService struct {
	mongoRepo       IPurchaseTransactionAdminRepositories
	kafkaRepo       purchaseRepository.IPurchaseMessageQueueRepository
	timeoutDuration time.Duration
}

func NewPurchaseTransactionAdminService(pst microservice.IPersisterMongo, kfProducer microservice.IProducer) IPurchaseTransactionAdminService {

	kafkaRepo := purchaseRepository.NewPurchaseMessageQueueRepository(kfProducer)
	mongoRepo := NewPurchaseTransactionAdminRepositories(pst)
	return &PurchaseTransactionAdminService{
		kafkaRepo:       kafkaRepo,
		mongoRepo:       mongoRepo,
		timeoutDuration: time.Duration(30) * time.Second,
	}
}

func (s *PurchaseTransactionAdminService) ReSyncPurchaseDoc(shopID string) error {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	docs, err := s.mongoRepo.FindPurchaseDocByShopID(ctx, shopID)
	if err != nil {
		return err
	}

	err = s.kafkaRepo.CreateInBatch(docs)
	if err != nil {
		return err
	}

	return nil
}

func (s *PurchaseTransactionAdminService) ReSyncPurchaseDeleteDoc(shopID string) error {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	docs, err := s.mongoRepo.FindPurchaseDocDeleteByShopID(ctx, shopID)

	if err != nil {
		return err
	}

	err = s.kafkaRepo.DeleteInBatch(docs)
	if err != nil {
		return err
	}
	return nil
}
