package debtorpayment

import (
	pkgModels "smlcloudplatform/internal/models"
	models "smlcloudplatform/internal/transaction/models"
)

type IDebtorPaymentConsumerService interface {
	Upsert(holdingCode string, docNo string, doc models.DebtorPaymentTransactionPG) error
	Delete(holdingCode string, docNo string) error
}

type DebtorPaymentConsumerService struct {
	repo IDebtorPaymentTransactionPGRepository
}

func NewDebtorPaymentConsumerService(repo IDebtorPaymentTransactionPGRepository) IDebtorPaymentConsumerService {

	return &DebtorPaymentConsumerService{
		repo: repo,
	}
}

func (s *DebtorPaymentConsumerService) Upsert(holdingCode string, docNo string, doc models.DebtorPaymentTransactionPG) error {
	findDoc, err := s.repo.Get(holdingCode, docNo)
	if err != nil {
		err = s.repo.Create(doc)
		if err != nil {
			return err
		}
	} else {

		isEqual := findDoc.CompareTo(&doc)

		if isEqual == false {
			err = s.repo.Update(holdingCode, docNo, doc)
			if err != nil {
				return err
			}
		} else {
			// logger.GetLogger().Debug("Doc is equal, skip update")
		}
	}

	return nil
}

func (s *DebtorPaymentConsumerService) Delete(holdingCode string, docNo string) error {
	err := s.repo.Delete(holdingCode, docNo, models.DebtorPaymentTransactionPG{
		HoldingCodeentity: pkgModels.HoldingCodeentity{
			HoldingCode: holdingCode,
		},
		DocNo: docNo,
	})
	if err != nil {
		return err
	}
	return nil
}
