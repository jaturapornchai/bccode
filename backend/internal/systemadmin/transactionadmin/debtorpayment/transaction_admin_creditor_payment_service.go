package debtorpayment

import (
	"context"
	debtorPaymentRepository "smlcloudplatform/internal/transaction/paid/repositories"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IDebtorPaymentTransactionAdminService interface {
	ReSyncDebtorPaymentDoc(shopID string) error
}

type DebtorPaymentTransactionAdminService struct {
	mongoRepo       IDebtorPaymentTransactionAdminRepository
	kafkaRepo       debtorPaymentRepository.IDebtorPaymentMessageQueueRepository
	timeoutDuration time.Duration
}

func NewDebtorPaymentTransactionAdminService(pst microservice.IPersisterMongo, kfProducer microservice.IProducer) IDebtorPaymentTransactionAdminService {

	kafkaRepo := debtorPaymentRepository.NewPaidMessageQueueRepository(kfProducer)
	mongoRepo := NewCreditorPaymentTransactionAdminRepository(pst)
	return &DebtorPaymentTransactionAdminService{
		kafkaRepo:       kafkaRepo,
		mongoRepo:       mongoRepo,
		timeoutDuration: time.Duration(30) * time.Second,
	}
}

func (s *DebtorPaymentTransactionAdminService) ReSyncDebtorPaymentDoc(shopID string) error {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	docs, err := s.mongoRepo.FindDebtorPaymentDocByShopID(ctx, shopID)
	if err != nil {
		return err
	}

	err = s.kafkaRepo.CreateInBatch(docs)
	if err != nil {
		return err
	}

	return nil
}
