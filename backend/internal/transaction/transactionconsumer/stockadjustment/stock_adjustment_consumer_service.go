package stockadjustment

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
)

type IStockAdjustmentTransactionConsumerService interface {
	Upsert(shopID string, docNo string, doc models.StockAdjustmentTransactionPG) error
	Delete(shopID string, docNo string) error
}

type StockAdjustmentTransactionConsumerService struct {
	repo IStockAdjustmentTransactionPGRepository
}

func NewStockAdjustmentTransactionConsumerService(repo IStockAdjustmentTransactionPGRepository) IStockAdjustmentTransactionConsumerService {
	return &StockAdjustmentTransactionConsumerService{
		repo: repo,
	}
}

func (s *StockAdjustmentTransactionConsumerService) Upsert(shopID string, docNo string, doc models.StockAdjustmentTransactionPG) error {
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

func (s *StockAdjustmentTransactionConsumerService) Delete(shopID string, docNo string) error {
	err := s.repo.DeleteData(shopID, docNo, models.StockAdjustmentTransactionPG{
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
