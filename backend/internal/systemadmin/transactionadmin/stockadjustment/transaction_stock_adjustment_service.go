package stockadjustment

import (
	"context"
	stockAdjustmentRepositories "smlcloudplatform/internal/transaction/stockadjustment/repositories"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IStockAdjustmentTransactionAdminService interface {
	ReSyncStockAdjustmentDoc(shopID string) error
	ReSyncStockAdjustmentDeleteDoc(shopID string) error
}

type StockAdjustmentTransactionAdminService struct {
	mongoRepo       IStockAdjustmentTransactionAdminRepository
	kafkaRepo       stockAdjustmentRepositories.IStockAdjustmentMessageQueueRepository
	timeoutDuration time.Duration
}

func NewStockAdjustmentTransactionAdminService(pst microservice.IPersisterMongo, kfProducer microservice.IProducer) IStockAdjustmentTransactionAdminService {

	mongoRepo := NewStockAdjustmentTransactionAdminRepository(pst)
	kafkaRepo := stockAdjustmentRepositories.NewStockAdjustmentMessageQueueRepository(kfProducer)

	return &StockAdjustmentTransactionAdminService{
		mongoRepo:       mongoRepo,
		kafkaRepo:       kafkaRepo,
		timeoutDuration: time.Duration(30) * time.Second,
	}
}

func (s *StockAdjustmentTransactionAdminService) ReSyncStockAdjustmentDoc(shopID string) error {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	docs, err := s.mongoRepo.FindStockAdjustmentDocByShopID(ctx, shopID)
	if err != nil {
		return err
	}

	err = s.kafkaRepo.CreateInBatch(docs)
	if err != nil {
		return err
	}
	return nil

}

func (s *StockAdjustmentTransactionAdminService) ReSyncStockAdjustmentDeleteDoc(shopID string) error {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	docs, err := s.mongoRepo.FindStockAdjustmentDocDeleteByShopID(ctx, shopID)

	if err != nil {
		return err
	}

	err = s.kafkaRepo.DeleteInBatch(docs)
	if err != nil {
		return err
	}
	return nil
}
