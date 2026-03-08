package stockreturnproduct

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
)

type IStockReturnProductConsumerService interface {
	Upsert(shopID string, docNo string, doc models.StockReturnProductTransactionPG) error
	Delete(shopID string, docNo string) error
}

type StockReturnProductConsumerService struct {
	repo IStockReturnTransactionPGRepository
}

func NewStockReturnProductConsumerService(repo IStockReturnTransactionPGRepository) IStockReturnProductConsumerService {
	return &StockReturnProductConsumerService{
		repo: repo,
	}
}

func (s *StockReturnProductConsumerService) Upsert(shopID string, docNo string, doc models.StockReturnProductTransactionPG) error {
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

func (s *StockReturnProductConsumerService) Delete(shopID string, docNo string) error {
	err := s.repo.DeleteData(shopID, docNo, models.StockReturnProductTransactionPG{
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
