package saleinvoicereturn

import (
	"context"
	saleInvoiceReturnRepo "smlcloudplatform/internal/transaction/saleinvoicereturn/repositories"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type ISaleInvoiceReturnTransactionAdminService interface {
	ReSyncSaleInvoiceReturnDoc(holdingCode string) error
	ReSyncSaleInvoiceReturnDeleteDoc(holdingCode string) error
}

type SaleInvoiceReturnTransactionAdminService struct {
	mongoRepo       ISaleInvoiceReturnTransactionAdminRepositories
	kafkaRepo       saleInvoiceReturnRepo.ISaleInvoiceReturnMessageQueueRepository
	timeoutDuration time.Duration
}

func NewSaleInvoiceReturnTransactionAdminService(pst microservice.IPersisterMongo, kfProducer microservice.IProducer) ISaleInvoiceReturnTransactionAdminService {

	mongoRepo := NewSaleInvoiceReturnTransactionAdminRepositories(pst)
	kafkaRepo := saleInvoiceReturnRepo.NewSaleInvoiceReturnMessageQueueRepository(kfProducer)
	return &SaleInvoiceReturnTransactionAdminService{
		mongoRepo:       mongoRepo,
		kafkaRepo:       kafkaRepo,
		timeoutDuration: time.Duration(30) * time.Second,
	}
}

func (s *SaleInvoiceReturnTransactionAdminService) ReSyncSaleInvoiceReturnDoc(holdingCode string) error {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	docs, err := s.mongoRepo.FindSaleInvoiceReturnDocByHoldingCode(ctx, holdingCode)
	if err != nil {
		return err
	}

	err = s.kafkaRepo.CreateInBatch(docs)
	if err != nil {
		return err
	}

	return nil
}

func (s *SaleInvoiceReturnTransactionAdminService) ReSyncSaleInvoiceReturnDeleteDoc(holdingCode string) error {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	docs, err := s.mongoRepo.FindSaleInvoiceReturnDeleteDocByHoldingCode(ctx, holdingCode)

	if err != nil {
		return err
	}

	err = s.kafkaRepo.DeleteInBatch(docs)
	if err != nil {
		return err
	}
	return nil
}
