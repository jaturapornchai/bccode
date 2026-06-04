package saleinvoicereturn

import (
	"smlcloudplatform/internal/logger"
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
)

type ISaleInvoiceReturnTransactionConsumerService interface {
	Upsert(holdingCode string, docNo string, doc models.SaleInvoiceReturnTransactionPG) error
	Delete(holdingCode string, docNo string) error
}

type SaleInvoiceReturnTransactionConsumerService struct {
	repo ISaleInvoiceReturnTransactionPGRepository
}

func NewSaleInvoiceReturnTransactionConsumerService(repo ISaleInvoiceReturnTransactionPGRepository) ISaleInvoiceReturnTransactionConsumerService {
	return &SaleInvoiceReturnTransactionConsumerService{
		repo: repo,
	}
}

func (s *SaleInvoiceReturnTransactionConsumerService) Upsert(holdingCode string, docNo string, doc models.SaleInvoiceReturnTransactionPG) error {
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

		if !isEqual {
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

func (s *SaleInvoiceReturnTransactionConsumerService) Delete(holdingCode string, docNo string) error {
	err := s.repo.DeleteData(holdingCode, docNo, models.SaleInvoiceReturnTransactionPG{
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
