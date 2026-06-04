package stocktransfer

import (
	"context"
	stocktransferrepositories "smlcloudplatform/internal/transaction/stocktransfer/repositories"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IStockTransferTransactionAdminService interface {
	ReSyncStockTransferDoc(holdingCode string) error
	ReSyncStockTransferDeleteDoc(holdingCode string) error
}

type StockTransferTransactionAdminService struct {
	mongoRepo       IStockTransferTransactionAdminRepository
	kafkaRepo       stocktransferrepositories.IStockTransferMessageQueueRepository
	timeoutDuration time.Duration
}

func NewStockTransferTransactionAdminService(pst microservice.IPersisterMongo, kfProducer microservice.IProducer) IStockTransferTransactionAdminService {

	mongoRepo := NewStockTransferTransactionAdminRepository(pst)
	kafkaRepo := stocktransferrepositories.NewStockTransferMessageQueueRepository(kfProducer)

	return &StockTransferTransactionAdminService{
		mongoRepo:       mongoRepo,
		kafkaRepo:       kafkaRepo,
		timeoutDuration: time.Duration(30) * time.Second,
	}
}

func (s *StockTransferTransactionAdminService) ReSyncStockTransferDoc(holdingCode string) error {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	docs, err := s.mongoRepo.FindStockTransferDocByHoldingCode(ctx, holdingCode)
	if err != nil {
		return err
	}

	err = s.kafkaRepo.CreateInBatch(docs)
	if err != nil {
		return err
	}
	return nil

}

func (s *StockTransferTransactionAdminService) ReSyncStockTransferDeleteDoc(holdingCode string) error {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	docs, err := s.mongoRepo.FindStockTransferDocDeleteByHoldingCode(ctx, holdingCode)

	if err != nil {
		return err
	}

	err = s.kafkaRepo.DeleteInBatch(docs)
	if err != nil {
		return err
	}
	return nil
}
