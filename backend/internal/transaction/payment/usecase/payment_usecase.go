package usecase

import (
	"smlcloudplatform/internal/transaction/payment/models"
	"smlcloudplatform/internal/transaction/payment/repositories"
)

type IPaymentUsecase interface {
	Upsert(shopID string, docNo string, doc models.TransactionPayment) error
	Delete(shopID string, docNo string) error
}

type PaymentUsecase struct {
	repo repositories.IPaymentRepository
}

func NewPaymentUsecase(repo repositories.IPaymentRepository) *PaymentUsecase {
	return &PaymentUsecase{
		repo: repo,
	}
}

func (s *PaymentUsecase) Upsert(shopID string, docNo string, doc models.TransactionPayment) error {
	foundDocument, err := s.repo.Get(shopID, docNo)

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
			err = s.repo.Update(shopID, docNo, doc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *PaymentUsecase) Delete(shopID string, docNo string) error {
	err := s.repo.Delete(shopID, docNo, models.TransactionPayment{
		ShopID: shopID,
		DocNo:  docNo,
	})
	if err != nil {
		return err
	}
	return nil
}
