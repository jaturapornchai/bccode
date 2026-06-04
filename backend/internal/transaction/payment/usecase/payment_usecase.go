package usecase

import (
	"smlcloudplatform/internal/transaction/payment/models"
	"smlcloudplatform/internal/transaction/payment/repositories"
)

type IPaymentUsecase interface {
	Upsert(holdingCode string, docNo string, doc models.TransactionPayment) error
	Delete(holdingCode string, docNo string) error
}

type PaymentUsecase struct {
	repo repositories.IPaymentRepository
}

func NewPaymentUsecase(repo repositories.IPaymentRepository) *PaymentUsecase {
	return &PaymentUsecase{
		repo: repo,
	}
}

func (s *PaymentUsecase) Upsert(holdingCode string, docNo string, doc models.TransactionPayment) error {
	foundDocument, err := s.repo.Get(holdingCode, docNo)

	if err != nil && err.Error() != "record not found" {
		return err
	}

	if foundDocument == nil {
		err = s.repo.Create(doc)
		if err != nil {
			return err
		}
	} else {

		if !foundDocument.CompareTo(&doc) {
			err = s.repo.Update(holdingCode, docNo, doc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *PaymentUsecase) Delete(holdingCode string, docNo string) error {
	err := s.repo.Delete(holdingCode, docNo, models.TransactionPayment{
		HoldingCode: holdingCode,
		DocNo:       docNo,
	})
	if err != nil {
		return err
	}
	return nil
}
