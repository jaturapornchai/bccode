package purchaseorder

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
)

type IPurchaseOrderTransactionConsumerService interface {
	Upsert(holdingCode string, docNo string, doc models.PurchaseOrderTransactionPG) error
	Delete(holdingCode string, docNo string) error
}

type PurchaseOrderTransactionConsumerService struct {
	repo IPurchaseOrderTransactionPGRepository
}

func NewPurchaseOrderTransactionService(repo IPurchaseOrderTransactionPGRepository) IPurchaseOrderTransactionConsumerService {
	return &PurchaseOrderTransactionConsumerService{
		repo: repo,
	}
}

func (s *PurchaseOrderTransactionConsumerService) Upsert(holdingCode string, docNo string, doc models.PurchaseOrderTransactionPG) error {
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
		}
	}

	return nil
}

func (s *PurchaseOrderTransactionConsumerService) Delete(holdingCode string, docNo string) error {

	err := s.repo.DeleteData(holdingCode, docNo, models.PurchaseOrderTransactionPG{
		TransactionPG: models.TransactionPG{
			HoldingCodeentity: pkgModels.HoldingCodeentity{
				HoldingCode: holdingCode,
			},
			DocNo: docNo,
		},
	})

	return err
}
