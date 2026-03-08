package usecase

import (
	"smlcloudplatform/internal/transaction/paymentdetail/models"
	"smlcloudplatform/internal/transaction/paymentdetail/repositories"
)

type IPaymentDetailUsecase interface {
	Upsert(shopID string, docNo string, doc models.TransactionPaymentDetail) error
	Delete(shopID string, docNo string) error
}

type PaymentDetailUsecase struct {
	repo repositories.IPaymentDetailRepository
}

func NewPaymentDetailUsecase(repo repositories.IPaymentDetailRepository) *PaymentDetailUsecase {
	return &PaymentDetailUsecase{
		repo: repo,
	}
}

func (s *PaymentDetailUsecase) Upsert(shopID string, docNo string, doc models.TransactionPaymentDetail) error {
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

func (s *PaymentDetailUsecase) Delete(shopID string, docNo string) error {
	err := s.repo.Delete(shopID, docNo, models.TransactionPaymentDetail{
		ShopID: shopID,
		DocNo:  docNo,
	})
	if err != nil {
		return err
	}
	return nil
}
