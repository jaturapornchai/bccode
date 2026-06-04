package creditorpayment

import (
	"context"
	creditPaymentRepository "smlcloudplatform/internal/transaction/pay/repositories"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type ICreditorPaymentTransactionAdminService interface {
	ReSyncCreditorPaymentDoc(holdingCode string) error
}

type CreditorPaymentTransactionAdminService struct {
	mongoRepo       ICreditorPaymentTransactionAdminRepository
	kafkaRepo       creditPaymentRepository.ICreditorPaymentMessageQueueRepository
	timeoutDuration time.Duration
}

func NewCreditorPaymentTransactionAdminService(pst microservice.IPersisterMongo, kfProducer microservice.IProducer) ICreditorPaymentTransactionAdminService {

	kafkaRepo := creditPaymentRepository.NewPaidMessageQueueRepository(kfProducer)
	mongoRepo := NewCreditorPaymentTransactionAdminRepository(pst)
	return &CreditorPaymentTransactionAdminService{
		kafkaRepo:       kafkaRepo,
		mongoRepo:       mongoRepo,
		timeoutDuration: time.Duration(30) * time.Second,
	}
}

func (s *CreditorPaymentTransactionAdminService) ReSyncCreditorPaymentDoc(holdingCode string) error {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	docs, err := s.mongoRepo.FindCreditorPaymentDocByHoldingCode(ctx, holdingCode)
	if err != nil {
		return err
	}

	err = s.kafkaRepo.CreateInBatch(docs)
	if err != nil {
		return err
	}

	return nil
}
