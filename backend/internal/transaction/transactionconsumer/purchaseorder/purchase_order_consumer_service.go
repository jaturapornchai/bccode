package purchaseorder

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
)

type IPurchaseOrderTransactionConsumerService interface {
	Upsert(shopID string, docNo string, doc models.PurchaseOrderTransactionPG) error
	Delete(shopID string, docNo string) error
}

type PurchaseOrderTransactionConsumerService struct {
	repo IPurchaseOrderTransactionPGRepository
}

func NewPurchaseOrderTransactionService(repo IPurchaseOrderTransactionPGRepository) IPurchaseOrderTransactionConsumerService {
	return &PurchaseOrderTransactionConsumerService{
		repo: repo,
	}
}

func (s *PurchaseOrderTransactionConsumerService) Upsert(shopID string, docNo string, doc models.PurchaseOrderTransactionPG) error {
	foundDocument, err := s.repo.Get(shopID, docNo)

	if err != nil && err.Error() != "record not found" {
		return err
	}

	if foundDocument == nil {
		return s.repo.Create(doc)
	} else {

		isEqual := foundDocument.CompareTo(&doc)
		if !isEqual {
			return s.repo.Update(shopID, docNo, doc)
		}
	}

	return nil
}

func (s *PurchaseOrderTransactionConsumerService) Delete(shopID string, docNo string) error {

	err := s.repo.DeleteData(shopID, docNo, models.PurchaseOrderTransactionPG{
		TransactionPG: models.TransactionPG{
			ShopIdentity: pkgModels.ShopIdentity{
				ShopID: shopID,
			},
			DocNo: docNo,
		},
	})

	return err
}
