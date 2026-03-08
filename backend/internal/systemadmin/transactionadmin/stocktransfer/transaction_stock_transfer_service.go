package stocktransfer

import (
	"context"
	stocktransferrepositories "smlcloudplatform/internal/transaction/stocktransfer/repositories"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IStockTransferTransactionAdminService interface {
	ReSyncStockTransferDoc(shopID string) error
	ReSyncStockTransferDeleteDoc(shopID string) error
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

func (s *StockTransferTransactionAdminService) ReSyncStockTransferDoc(shopID string) error {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	docs, err := s.mongoRepo.FindStockTransferDocByShopID(ctx, shopID)
	if err != nil {
		return err
	}

	err = s.kafkaRepo.CreateInBatch(docs)
	if err != nil {
		return err
	}
	return nil

}

func (s *StockTransferTransactionAdminService) ReSyncStockTransferDeleteDoc(shopID string) error {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	docs, err := s.mongoRepo.FindStockTransferDocDeleteByShopID(ctx, shopID)

	if err != nil {
		return err
	}

	err = s.kafkaRepo.DeleteInBatch(docs)
	if err != nil {
		return err
	}
	return nil
}
