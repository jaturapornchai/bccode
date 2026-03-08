package saleinvoice

import (
	"smlcloudplatform/internal/logger"
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
)

type ISaleInvoiceTransactionConsumerService interface {
	Upsert(shopID string, docNo string, doc models.SaleInvoiceTransactionPG) error
	Delete(shopID string, docNo string) error
}

type SaleInvoiceTransactionConsumerService struct {
	repo ISaleInvoiceTransactionPGRepository
}

func NewSaleInvoiceTransactionConsumerService(repo ISaleInvoiceTransactionPGRepository) ISaleInvoiceTransactionConsumerService {
	return &SaleInvoiceTransactionConsumerService{
		repo: repo,
	}
}

func (s *SaleInvoiceTransactionConsumerService) Upsert(shopID string, docNo string, doc models.SaleInvoiceTransactionPG) error {
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

func (s *SaleInvoiceTransactionConsumerService) Delete(shopID string, docNo string) error {
	err := s.repo.DeleteData(shopID, docNo, models.SaleInvoiceTransactionPG{
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
