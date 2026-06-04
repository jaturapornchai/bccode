package stocktransfer

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
)

type IStockTransferTransactionConsumerService interface {
	Upsert(holdingCode string, docNo string, doc models.StockTransferTransactionPG) error
	Delete(holdingCode string, docNo string) error
}

type StockTransferTransactionConsumerService struct {
	repo IStockTransferTransactionPGRepository
}

func NewStockTransferTransactionConsumerService(repo IStockTransferTransactionPGRepository) IStockTransferTransactionConsumerService {
	return &StockTransferTransactionConsumerService{
		repo: repo,
	}
}

func (s *StockTransferTransactionConsumerService) Upsert(holdingCode string, docNo string, doc models.StockTransferTransactionPG) error {
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

func (s *StockTransferTransactionConsumerService) Delete(holdingCode string, docNo string) error {
	err := s.repo.DeleteData(holdingCode, docNo, models.StockTransferTransactionPG{
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
