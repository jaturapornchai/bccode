package stockreturnproduct

import (
	"context"
	stockreturnproductrepositories "smlcloudplatform/internal/transaction/stockreturnproduct/repositories"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IStockReturnProductTransactionAdminService interface {
	ReSyncStockReturnProductDoc(shopID string) error
	ReSyncStockReturnProductDeleteDoc(shopID string) error
}

type StockReturnProductTransactionAdminService struct {
	mongoRepo       IStockReturnProductTransactionAdminRepository
	kafkaRepo       stockreturnproductrepositories.IStockReturnProductMessageQueueRepository
	timeoutDuration time.Duration
}

func NewStockReturnProductTransactionAdminService(pst microservice.IPersisterMongo, kfProducer microservice.IProducer) IStockReturnProductTransactionAdminService {

	mongoRepo := NewStockReturnProductTransactionAdminRepository(pst)
	kafkaRepo := stockreturnproductrepositories.NewStockReturnProductMessageQueueRepository(kfProducer)

	return &StockReturnProductTransactionAdminService{
		mongoRepo:       mongoRepo,
		kafkaRepo:       kafkaRepo,
		timeoutDuration: time.Duration(30) * time.Second,
	}
}

func (s *StockReturnProductTransactionAdminService) ReSyncStockReturnProductDoc(shopID string) error {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	docs, err := s.mongoRepo.FindStockReturnProductDocByShopID(ctx, shopID)
	if err != nil {
		return err
	}

	err = s.kafkaRepo.CreateInBatch(docs)
	if err != nil {
		return err
	}
	return nil

}

func (s *StockReturnProductTransactionAdminService) ReSyncStockReturnProductDeleteDoc(shopID string) error {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	docs, err := s.mongoRepo.FindStockReturnProductDeleteDocByShopID(ctx, shopID)

	if err != nil {
		return err
	}

	err = s.kafkaRepo.DeleteInBatch(docs)
	if err != nil {
		return err
	}
	return nil
}
