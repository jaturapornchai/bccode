package stockreceiveproduct

import (
	"context"
	stockReceiveProductRepositories "smlcloudplatform/internal/transaction/stockreceiveproduct/repositories"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IStockReceiveProductTransactionAdminService interface {
	ReSyncStockReceiveProductDoc(holdingCode string) error
	ReSyncStockReceiveProductDeleteDoc(holdingCode string) error
}

type StockReceiveProductTransactionAdminService struct {
	mongoRepo       IStockReceiveTransactionAdminRepository
	kafkaRepo       stockReceiveProductRepositories.IStockReceiveProductMessageQueueRepository
	timeoutDuration time.Duration
}

func NewStockReceiveProductTransactionAdminService(
	pst microservice.IPersisterMongo,
	kfProducer microservice.IProducer,
) IStockReceiveProductTransactionAdminService {

	mongoRepo := NewStockReceiveTransactionAdminRepository(pst)
	kafkaRepo := stockReceiveProductRepositories.NewStockReceiveProductMessageQueueRepository(kfProducer)

	return &StockReceiveProductTransactionAdminService{
		mongoRepo:       mongoRepo,
		kafkaRepo:       kafkaRepo,
		timeoutDuration: time.Duration(30) * time.Second,
	}
}

func (s *StockReceiveProductTransactionAdminService) ReSyncStockReceiveProductDoc(holdingCode string) error {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	docs, err := s.mongoRepo.FindStockReceiveDocByHoldingCode(ctx, holdingCode)
	if err != nil {
		return err
	}

	err = s.kafkaRepo.CreateInBatch(docs)
	if err != nil {
		return err
	}

	return nil
}

func (s *StockReceiveProductTransactionAdminService) ReSyncStockReceiveProductDeleteDoc(holdingCode string) error {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	docs, err := s.mongoRepo.FindStockReceiveDeleteDocByHoldingCode(ctx, holdingCode)

	if err != nil {
		return err
	}

	err = s.kafkaRepo.DeleteInBatch(docs)
	if err != nil {
		return err
	}
	return nil
}
