package purchasereceive

import (
	"smlcloudplatform/internal/logger"
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
)

type IPurchaseReceiveTransactionConsumerService interface {
	Upsert(holdingCode string, docNo string, doc models.PurchaseReceiveTransactionPG) error
	Delete(holdingCode string, docNo string) error
}

type PurchaseReceiveTransactionConsumerService struct {
	repo IPurchaseReceiveTransactionPGRepository
}

func NewPurchaseReceiveTransactionService(repo IPurchaseReceiveTransactionPGRepository) IPurchaseReceiveTransactionConsumerService {
	return &PurchaseReceiveTransactionConsumerService{
		repo: repo,
	}
}

func (s *PurchaseReceiveTransactionConsumerService) Upsert(holdingCode string, docNo string, doc models.PurchaseReceiveTransactionPG) error {
	foundDocument, err := s.repo.Get(holdingCode, docNo)

	if err != nil && err.Error() != "record not found" {
		return err
	}

	if foundDocument == nil {
		return s.repo.Create(doc)
	} else {

		isEqual := foundDocument.CompareTo(&doc)
		if !isEqual {
			return s.repo.Update(holdingCode, docNo, doc)
		} else {
			logger.GetLogger().Debug("Doc is equal, skip update")
		}
	}

	return nil
}

func (s *PurchaseReceiveTransactionConsumerService) Delete(holdingCode string, docNo string) error {

	err := s.repo.DeleteData(holdingCode, docNo, models.PurchaseReceiveTransactionPG{
		TransactionPG: models.TransactionPG{
			HoldingCodeentity: pkgModels.HoldingCodeentity{
				HoldingCode: holdingCode,
			},
			DocNo: docNo,
		},
	})

	return err

}
