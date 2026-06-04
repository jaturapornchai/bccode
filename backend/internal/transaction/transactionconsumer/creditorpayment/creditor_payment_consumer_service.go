package creditorpayment

import (
	pkgModels "smlcloudplatform/internal/models"
	models "smlcloudplatform/internal/transaction/models"
)

type ICreditorPaymentTransactionConsumerService interface {
	Upsert(holdingCode string, docNo string, doc models.CreditorPaymentTransactionPG) error
	Delete(holdingCode string, docNo string) error
}

type CreditorPaymentTransactionConsumerService struct {
	repo ICreditorPaymentTransactionPGRepository
}

func NewCreditorPaymentTransactionConsumerService(repo ICreditorPaymentTransactionPGRepository) ICreditorPaymentTransactionConsumerService {
	return &CreditorPaymentTransactionConsumerService{
		repo: repo,
	}
}

func (s *CreditorPaymentTransactionConsumerService) Upsert(holdingCode string, docNo string, doc models.CreditorPaymentTransactionPG) error {
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

func (s *CreditorPaymentTransactionConsumerService) Delete(holdingCode string, docNo string) error {
	err := s.repo.Delete(holdingCode, docNo, models.CreditorPaymentTransactionPG{
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
