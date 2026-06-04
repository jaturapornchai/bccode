package stockpickupproduct

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
)

type IStockPickupTransactionConsumerService interface {
	Upsert(holdingCode string, docNo string, doc models.StockPickUpTransactionPG) error
	Delete(holdingCode string, docNo string) error
}

type StockPickupTransactionConsumerService struct {
	repo IStockPickupTransactionPGRepository
}

func NewStockPickupTransactionConsumerService(repo IStockPickupTransactionPGRepository) IStockPickupTransactionConsumerService {

	return &StockPickupTransactionConsumerService{
		repo: repo,
	}
}

func (s *StockPickupTransactionConsumerService) Upsert(holdingCode string, docNo string, doc models.StockPickUpTransactionPG) error {
	findDoc, err := s.repo.Get(holdingCode, docNo)
	if err != nil {
		err = s.repo.Create(doc)
		if err != nil {
			return err
		}
	} else {

		isEqual := findDoc.CompareTo(&doc)

		if !isEqual {
			err = s.repo.Update(holdingCode, docNo, doc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *StockPickupTransactionConsumerService) Delete(holdingCode string, docNo string) error {
	err := s.repo.DeleteData(holdingCode, docNo, models.StockPickUpTransactionPG{
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
