package stockpickupproduct

import (
	"context"
	stockPickupProductRepositories "smlcloudplatform/internal/transaction/stockpickupproduct/repositories"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IStockPickupTransactionAdminService interface {
	ReSyncStockPickupTransaction(shopID string) error
	ReSyncStockPickupDeleteTransaction(shopID string) error
}

type StockPickupTransactionAdminService struct {
	mongoRepo       IStockPickupTransactionAdminRepository
	kafkaRepo       stockPickupProductRepositories.IStockPickupProductMessageQueueRepository
	timeoutDuration time.Duration
}

func NewStockPickupTransactionAdminService(
	pst microservice.IPersisterMongo,
	kfProducer microservice.IProducer,
) IStockPickupTransactionAdminService {

	mongoRepo := NewStockPickupTransactionAdminRepository(pst)
	kafkaRepo := stockPickupProductRepositories.NewStockPickupProductMessageQueueRepository(kfProducer)

	return &StockPickupTransactionAdminService{
		mongoRepo:       mongoRepo,
		kafkaRepo:       kafkaRepo,
		timeoutDuration: time.Duration(30) * time.Second,
	}
}

func (s *StockPickupTransactionAdminService) ReSyncStockPickupTransaction(shopID string) error {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	docs, err := s.mongoRepo.FindStockPickupDocByShopID(ctx, shopID)
	if err != nil {
		return err
	}

	err = s.kafkaRepo.CreateInBatch(docs)
	if err != nil {
		return err
	}

	return nil
}

func (s *StockPickupTransactionAdminService) ReSyncStockPickupDeleteTransaction(shopID string) error {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	docs, err := s.mongoRepo.FindStockPickupDocDeleteByShopID(ctx, shopID)

	if err != nil {
		return err
	}

	err = s.kafkaRepo.DeleteInBatch(docs)
	if err != nil {
		return err
	}
	return nil
}
