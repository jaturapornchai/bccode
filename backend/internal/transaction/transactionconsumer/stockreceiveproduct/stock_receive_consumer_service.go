package stockreceiveproduct

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
)

type IStockReceiveTransactionConsumerService interface {
	Upsert(holdingCode string, docNo string, doc models.StockReceiveProductTransactionPG) error
	Delete(holdingCode string, docNo string) error
}

type StockReceiveTransactionConsumerService struct {
	repo IStockReceiveTransactionPGRepository
}

func NewStockReceiveTransactionConsumerService(repo IStockReceiveTransactionPGRepository) IStockReceiveTransactionConsumerService {

	return &StockReceiveTransactionConsumerService{
		repo: repo,
	}
}

func (s *StockReceiveTransactionConsumerService) Upsert(holdingCode string, docNo string, doc models.StockReceiveProductTransactionPG) error {
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

func (s *StockReceiveTransactionConsumerService) Delete(holdingCode string, docNo string) error {
	err := s.repo.DeleteData(holdingCode, docNo, models.StockReceiveProductTransactionPG{
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
