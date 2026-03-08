package stocktransfer

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
)

type IStockTransferTransactionConsumerService interface {
	Upsert(shopID string, docNo string, doc models.StockTransferTransactionPG) error
	Delete(shopID string, docNo string) error
}

type StockTransferTransactionConsumerService struct {
	repo IStockTransferTransactionPGRepository
}

func NewStockTransferTransactionConsumerService(repo IStockTransferTransactionPGRepository) IStockTransferTransactionConsumerService {
	return &StockTransferTransactionConsumerService{
		repo: repo,
	}
}

func (s *StockTransferTransactionConsumerService) Upsert(shopID string, docNo string, doc models.StockTransferTransactionPG) error {
	findDoc, err := s.repo.Get(shopID, docNo)
	if err != nil {
		err = s.repo.Create(doc)
		if err != nil {
			return err
		}
	} else {

		isEqual := findDoc.CompareTo(&doc)

		if !isEqual {
			err = s.repo.Update(shopID, docNo, doc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *StockTransferTransactionConsumerService) Delete(shopID string, docNo string) error {
	err := s.repo.DeleteData(shopID, docNo, models.StockTransferTransactionPG{
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
