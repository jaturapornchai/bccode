package purchaserequisition

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
)

type IPurchaseRequisitionTransactionConsumerService interface {
	Upsert(holdingCode string, docNo string, doc models.PurchaseRequisitionTransactionPG) error
	Delete(holdingCode string, docNo string) error
}

type PurchaseRequisitionTransactionConsumerService struct {
	repo IPurchaseRequisitionTransactionPGRepository
}

func NewPurchaseRequisitionTransactionService(repo IPurchaseRequisitionTransactionPGRepository) IPurchaseRequisitionTransactionConsumerService {
	return &PurchaseRequisitionTransactionConsumerService{
		repo: repo,
	}
}

func (s *PurchaseRequisitionTransactionConsumerService) Upsert(holdingCode string, docNo string, doc models.PurchaseRequisitionTransactionPG) error {
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

func (s *PurchaseRequisitionTransactionConsumerService) Delete(holdingCode string, docNo string) error {
	err := s.repo.DeleteData(holdingCode, docNo, models.PurchaseRequisitionTransactionPG{
		TransactionPG: models.TransactionPG{
			HoldingCodeentity: pkgModels.HoldingCodeentity{
				HoldingCode: holdingCode,
			},
			DocNo: docNo,
		},
	})
	return err
}
