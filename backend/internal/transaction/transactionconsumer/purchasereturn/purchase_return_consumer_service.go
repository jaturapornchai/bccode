package purchasereturn

import (
	"smlcloudplatform/internal/logger"
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
)

type IPurchaseReturnTransactionConsumerService interface {
	Upsert(shopID string, docNo string, doc models.PurchaseReturnTransactionPG) error
	Delete(shopID string, docNo string) error
}

type PurchaseReturnTransactionConsumerService struct {
	repo IPurchaseReturnTransactionPGRepository
}

func NewPurchaseReturnTransactionService(repo IPurchaseReturnTransactionPGRepository) IPurchaseReturnTransactionConsumerService {
	return &PurchaseReturnTransactionConsumerService{
		repo: repo,
	}
}

func (s *PurchaseReturnTransactionConsumerService) Upsert(shopID string, docNo string, doc models.PurchaseReturnTransactionPG) error {
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

		isEqual := foundDocument.CompareTo(&doc)

		if !isEqual {
			err = s.repo.Update(shopID, docNo, doc)
			if err != nil {
				return err
			}
		} else {
			logger.GetLogger().Debug("Doc is equal, skip update")
		}
	}

	return nil
}

func (s *PurchaseReturnTransactionConsumerService) Delete(shopID string, docNo string) error {
	err := s.repo.DeleteData(shopID, docNo, models.PurchaseReturnTransactionPG{
		TransactionPG: models.TransactionPG{
			ShopIdentity: pkgModels.ShopIdentity{
				ShopID: shopID,
			},
			DocNo: docNo,
		},
	})
	if err != nil {
		return err
	}
	return nil
}
