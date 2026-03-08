package stockreceiveproduct

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
)

type IStockReceiveTransactionConsumerService interface {
	Upsert(shopID string, docNo string, doc models.StockReceiveProductTransactionPG) error
	Delete(shopID string, docNo string) error
}

type StockReceiveTransactionConsumerService struct {
	repo IStockReceiveTransactionPGRepository
}

func NewStockReceiveTransactionConsumerService(repo IStockReceiveTransactionPGRepository) IStockReceiveTransactionConsumerService {

	return &StockReceiveTransactionConsumerService{
		repo: repo,
	}
}

func (s *StockReceiveTransactionConsumerService) Upsert(shopID string, docNo string, doc models.StockReceiveProductTransactionPG) error {
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

func (s *StockReceiveTransactionConsumerService) Delete(shopID string, docNo string) error {
	err := s.repo.DeleteData(shopID, docNo, models.StockReceiveProductTransactionPG{
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
