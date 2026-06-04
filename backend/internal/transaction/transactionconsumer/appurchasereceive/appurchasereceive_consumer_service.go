package appurchasereceive

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
)

type IAPPurchaseReceiveTransactionConsumerService interface {
	Upsert(holdingCode string, docNo string, doc models.APPurchaseReceivePG) error
	Delete(holdingCode string, docNo string) error
}

type APPurchaseReceiveTransactionConsumerService struct {
	repo IAPPurchaseReceiveTransactionPostgresRepository
}

func NewAPPurchaseReceiveTransactionConsumerService(repo IAPPurchaseReceiveTransactionPostgresRepository) IAPPurchaseReceiveTransactionConsumerService {
	return &APPurchaseReceiveTransactionConsumerService{
		repo: repo,
	}
}

func (s *APPurchaseReceiveTransactionConsumerService) Upsert(holdingCode string, docNo string, doc models.APPurchaseReceivePG) error {
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

		isEqual := foundDocument.CompareTo(&doc)

		if isEqual == false {
			err = s.repo.Update(holdingCode, docNo, doc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *APPurchaseReceiveTransactionConsumerService) Delete(holdingCode string, docNo string) error {

	err := s.repo.DeleteData(holdingCode, docNo, models.APPurchaseReceivePG{
		TransactionPG: models.TransactionPG{
			HoldingCodeentity: pkgModels.HoldingCodeentity{
				HoldingCode: holdingCode,
			},
			DocNo: docNo,
		},
	})
	if err != nil {
		return err
	}
	return nil
}
