package purchase

import (
	"smlcloudplatform/internal/logger"
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
)

type IPurchaseTransactionConsumerService interface {
	Upsert(holdingCode string, docNo string, doc models.PurchaseTransactionPG) error
	Delete(holdingCode string, docNo string) error
}

type PurchaseTransactionConsumerService struct {
	repo IPurchaseTransactionPGRepository
}

func NewPurchaseTransactionService(repo IPurchaseTransactionPGRepository) IPurchaseTransactionConsumerService {
	return &PurchaseTransactionConsumerService{
		repo: repo,
	}
}

func (s *PurchaseTransactionConsumerService) Upsert(holdingCode string, docNo string, doc models.PurchaseTransactionPG) error {
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
		} else {
			logger.GetLogger().Debug("Doc is equal, skip update")
		}
	}

	return nil
}

func (s *PurchaseTransactionConsumerService) Delete(holdingCode string, docNo string) error {

	err := s.repo.DeleteData(holdingCode, docNo, models.PurchaseTransactionPG{
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
