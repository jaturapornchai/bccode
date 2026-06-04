package usecase

import (
	"smlcloudplatform/internal/transaction/paymentdetail/models"
	"smlcloudplatform/internal/transaction/paymentdetail/repositories"
)

type IPaymentDetailUsecase interface {
	Upsert(holdingCode string, docNo string, doc models.TransactionPaymentDetail) error
	Delete(holdingCode string, docNo string) error
}

type PaymentDetailUsecase struct {
	repo repositories.IPaymentDetailRepository
}

func NewPaymentDetailUsecase(repo repositories.IPaymentDetailRepository) *PaymentDetailUsecase {
	return &PaymentDetailUsecase{
		repo: repo,
	}
}

func (s *PaymentDetailUsecase) Upsert(holdingCode string, docNo string, doc models.TransactionPaymentDetail) error {
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

func (s *PaymentDetailUsecase) Delete(holdingCode string, docNo string) error {
	err := s.repo.Delete(holdingCode, docNo, models.TransactionPaymentDetail{
		HoldingCode: holdingCode,
		DocNo:       docNo,
	})
	if err != nil {
		return err
	}
	return nil
}
