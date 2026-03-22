package purchaserequisition

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
)

type IPurchaseRequisitionTransactionConsumerService interface {
	Upsert(shopID string, docNo string, doc models.PurchaseRequisitionTransactionPG) error
	Delete(shopID string, docNo string) error
}

type PurchaseRequisitionTransactionConsumerService struct {
	repo IPurchaseRequisitionTransactionPGRepository
}

func NewPurchaseRequisitionTransactionService(repo IPurchaseRequisitionTransactionPGRepository) IPurchaseRequisitionTransactionConsumerService {
	return &PurchaseRequisitionTransactionConsumerService{
		repo: repo,
	}
}

func (s *PurchaseRequisitionTransactionConsumerService) Upsert(shopID string, docNo string, doc models.PurchaseRequisitionTransactionPG) error {
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

func (s *PurchaseRequisitionTransactionConsumerService) Delete(shopID string, docNo string) error {
	err := s.repo.DeleteData(shopID, docNo, models.PurchaseRequisitionTransactionPG{
		TransactionPG: models.TransactionPG{
			ShopIdentity: pkgModels.ShopIdentity{
				ShopID: shopID,
			},
			DocNo: docNo,
		},
	})
	return err
}
