package stockreturnproduct

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
)

type IStockReturnProductConsumerService interface {
	Upsert(holdingCode string, docNo string, doc models.StockReturnProductTransactionPG) error
	Delete(holdingCode string, docNo string) error
}

type StockReturnProductConsumerService struct {
	repo IStockReturnTransactionPGRepository
}

func NewStockReturnProductConsumerService(repo IStockReturnTransactionPGRepository) IStockReturnProductConsumerService {
	return &StockReturnProductConsumerService{
		repo: repo,
	}
}

func (s *StockReturnProductConsumerService) Upsert(holdingCode string, docNo string, doc models.StockReturnProductTransactionPG) error {
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

func (s *StockReturnProductConsumerService) Delete(holdingCode string, docNo string) error {
	err := s.repo.DeleteData(holdingCode, docNo, models.StockReturnProductTransactionPG{
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
