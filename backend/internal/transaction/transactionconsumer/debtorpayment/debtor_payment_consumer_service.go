package debtorpayment

import (
	pkgModels "smlcloudplatform/internal/models"
	models "smlcloudplatform/internal/transaction/models"
)

type IDebtorPaymentConsumerService interface {
	Upsert(shopID string, docNo string, doc models.DebtorPaymentTransactionPG) error
	Delete(shopID string, docNo string) error
}

type DebtorPaymentConsumerService struct {
	repo IDebtorPaymentTransactionPGRepository
}

func NewDebtorPaymentConsumerService(repo IDebtorPaymentTransactionPGRepository) IDebtorPaymentConsumerService {

	return &DebtorPaymentConsumerService{
		repo: repo,
	}
}

func (s *DebtorPaymentConsumerService) Upsert(shopID string, docNo string, doc models.DebtorPaymentTransactionPG) error {
	findDoc, err := s.repo.Get(shopID, docNo)
	if err != nil {
		err = s.repo.Create(doc)
		if err != nil {
			return err
		}
	} else {

		isEqual := findDoc.CompareTo(&doc)

		if isEqual == false {
			err = s.repo.Update(shopID, docNo, doc)
			if err != nil {
				return err
			}
		} else {
			// logger.GetLogger().Debug("Doc is equal, skip update")
		}
	}

	return nil
}

func (s *DebtorPaymentConsumerService) Delete(shopID string, docNo string) error {
	err := s.repo.Delete(shopID, docNo, models.DebtorPaymentTransactionPG{
		ShopIdentity: pkgModels.ShopIdentity{
			ShopID: shopID,
		},
		DocNo: docNo,
	})
	if err != nil {
		return err
	}
	return nil
}
