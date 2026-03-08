package creditorpayment

import (
	pkgModels "smlcloudplatform/internal/models"
	models "smlcloudplatform/internal/transaction/models"
)

type ICreditorPaymentTransactionConsumerService interface {
	Upsert(shopID string, docNo string, doc models.CreditorPaymentTransactionPG) error
	Delete(shopID string, docNo string) error
}

type CreditorPaymentTransactionConsumerService struct {
	repo ICreditorPaymentTransactionPGRepository
}

func NewCreditorPaymentTransactionConsumerService(repo ICreditorPaymentTransactionPGRepository) ICreditorPaymentTransactionConsumerService {
	return &CreditorPaymentTransactionConsumerService{
		repo: repo,
	}
}

func (s *CreditorPaymentTransactionConsumerService) Upsert(shopID string, docNo string, doc models.CreditorPaymentTransactionPG) error {
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

func (s *CreditorPaymentTransactionConsumerService) Delete(shopID string, docNo string) error {
	err := s.repo.Delete(shopID, docNo, models.CreditorPaymentTransactionPG{
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
