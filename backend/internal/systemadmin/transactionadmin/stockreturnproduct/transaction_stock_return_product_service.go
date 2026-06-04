package stockreturnproduct

import (
	"context"
	stockreturnproductrepositories "smlcloudplatform/internal/transaction/stockreturnproduct/repositories"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IStockReturnProductTransactionAdminService interface {
	ReSyncStockReturnProductDoc(holdingCode string) error
	ReSyncStockReturnProductDeleteDoc(holdingCode string) error
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

func (s *StockReturnProductTransactionAdminService) ReSyncStockReturnProductDoc(holdingCode string) error {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	docs, err := s.mongoRepo.FindStockReturnProductDocByHoldingCode(ctx, holdingCode)
	if err != nil {
		return err
	}

	err = s.kafkaRepo.CreateInBatch(docs)
	if err != nil {
		return err
	}
	return nil

}

func (s *StockReturnProductTransactionAdminService) ReSyncStockReturnProductDeleteDoc(holdingCode string) error {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	docs, err := s.mongoRepo.FindStockReturnProductDeleteDocByHoldingCode(ctx, holdingCode)

	if err != nil {
		return err
	}

	err = s.kafkaRepo.DeleteInBatch(docs)
	if err != nil {
		return err
	}
	return nil
}
