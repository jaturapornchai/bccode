package purchasereturn

import (
	"context"
	purchaseReturnRepository "smlcloudplatform/internal/transaction/purchasereturn/repositories"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IPurchaseReturnTransactionAdminService interface {
	ResyncPurchaseReturnDoc(shopID string) error
	ResyncPurchaseReturnDeleteDoc(shopID string) error
}

type PurchaseReturnTransactionAdminService struct {
	mongoRepo       IPurchaseReturnTransactionAdminRepository
	kafkaRepo       purchaseReturnRepository.IPurchaseReturnMessageQueueRepository
	timeoutDuration time.Duration
}

func NewPurchaseReturnTransactionAdminService(
	pst microservice.IPersisterMongo,
	kfProducer microservice.IProducer,
) IPurchaseReturnTransactionAdminService {

	kafkaRepo := purchaseReturnRepository.NewPurchaseReturnMessageQueueRepository(kfProducer)
	mongoRepo := NewPurchaseReturnTransactionAdminRepository(pst)
	return &PurchaseReturnTransactionAdminService{
		kafkaRepo:       kafkaRepo,
		mongoRepo:       mongoRepo,
		timeoutDuration: time.Duration(30) * time.Second,
	}
}

func (s *PurchaseReturnTransactionAdminService) ResyncPurchaseReturnDoc(shopID string) error {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	docs, err := s.mongoRepo.FindPurchaseReturnDocByShopID(ctx, shopID)
	if err != nil {
		return err
	}

	err = s.kafkaRepo.CreateInBatch(docs)
	if err != nil {
		return err
	}

	return nil
}

func (s *PurchaseReturnTransactionAdminService) ResyncPurchaseReturnDeleteDoc(shopID string) error {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	docs, err := s.mongoRepo.FindPurchaseReturnDeleteDocByShopID(ctx, shopID)

	if err != nil {
		return err
	}

	err = s.kafkaRepo.DeleteInBatch(docs)
	if err != nil {
		return err
	}
	return nil
}
