package stockpickupproduct

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
)

type IStockPickupTransactionConsumerService interface {
	Upsert(shopID string, docNo string, doc models.StockPickUpTransactionPG) error
	Delete(shopID string, docNo string) error
}

type StockPickupTransactionConsumerService struct {
	repo IStockPickupTransactionPGRepository
}

func NewStockPickupTransactionConsumerService(repo IStockPickupTransactionPGRepository) IStockPickupTransactionConsumerService {

	return &StockPickupTransactionConsumerService{
		repo: repo,
	}
}

func (s *StockPickupTransactionConsumerService) Upsert(shopID string, docNo string, doc models.StockPickUpTransactionPG) error {
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

func (s *StockPickupTransactionConsumerService) Delete(shopID string, docNo string) error {
	err := s.repo.DeleteData(shopID, docNo, models.StockPickUpTransactionPG{
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
