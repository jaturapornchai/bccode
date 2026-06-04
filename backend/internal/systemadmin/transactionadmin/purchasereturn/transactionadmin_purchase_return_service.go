package purchasereturn

import (
	"context"
	purchaseReturnRepository "smlcloudplatform/internal/transaction/purchasereturn/repositories"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IPurchaseReturnTransactionAdminService interface {
	ResyncPurchaseReturnDoc(holdingCode string) error
	ResyncPurchaseReturnDeleteDoc(holdingCode string) error
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

func (s *PurchaseReturnTransactionAdminService) ResyncPurchaseReturnDoc(holdingCode string) error {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	docs, err := s.mongoRepo.FindPurchaseReturnDocByHoldingCode(ctx, holdingCode)
	if err != nil {
		return err
	}

	err = s.kafkaRepo.CreateInBatch(docs)
	if err != nil {
		return err
	}

	return nil
}

func (s *PurchaseReturnTransactionAdminService) ResyncPurchaseReturnDeleteDoc(holdingCode string) error {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	docs, err := s.mongoRepo.FindPurchaseReturnDeleteDocByHoldingCode(ctx, holdingCode)

	if err != nil {
		return err
	}

	err = s.kafkaRepo.DeleteInBatch(docs)
	if err != nil {
		return err
	}
	return nil
}
