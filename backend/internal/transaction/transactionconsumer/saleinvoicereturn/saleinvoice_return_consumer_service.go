package saleinvoicereturn

import (
	"smlcloudplatform/internal/logger"
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
)

type ISaleInvoiceReturnTransactionConsumerService interface {
	Upsert(shopID string, docNo string, doc models.SaleInvoiceReturnTransactionPG) error
	Delete(shopID string, docNo string) error
}

type SaleInvoiceReturnTransactionConsumerService struct {
	repo ISaleInvoiceReturnTransactionPGRepository
}

func NewSaleInvoiceReturnTransactionConsumerService(repo ISaleInvoiceReturnTransactionPGRepository) ISaleInvoiceReturnTransactionConsumerService {
	return &SaleInvoiceReturnTransactionConsumerService{
		repo: repo,
	}
}

func (s *SaleInvoiceReturnTransactionConsumerService) Upsert(shopID string, docNo string, doc models.SaleInvoiceReturnTransactionPG) error {
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

func (s *SaleInvoiceReturnTransactionConsumerService) Delete(shopID string, docNo string) error {
	err := s.repo.DeleteData(shopID, docNo, models.SaleInvoiceReturnTransactionPG{
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
