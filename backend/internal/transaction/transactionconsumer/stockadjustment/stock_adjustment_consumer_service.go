package stockadjustment

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
)

type IStockAdjustmentTransactionConsumerService interface {
	Upsert(holdingCode string, docNo string, doc models.StockAdjustmentTransactionPG) error
	Delete(holdingCode string, docNo string) error
}

type StockAdjustmentTransactionConsumerService struct {
	repo IStockAdjustmentTransactionPGRepository
}

func NewStockAdjustmentTransactionConsumerService(repo IStockAdjustmentTransactionPGRepository) IStockAdjustmentTransactionConsumerService {
	return &StockAdjustmentTransactionConsumerService{
		repo: repo,
	}
}

func (s *StockAdjustmentTransactionConsumerService) Upsert(holdingCode string, docNo string, doc models.StockAdjustmentTransactionPG) error {
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

func (s *StockAdjustmentTransactionConsumerService) Delete(holdingCode string, docNo string) error {
	err := s.repo.DeleteData(holdingCode, docNo, models.StockAdjustmentTransactionPG{
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
