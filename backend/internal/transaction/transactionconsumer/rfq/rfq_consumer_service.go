package rfq

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
)

type IRFQTransactionConsumerService interface {
	Upsert(shopID string, docNo string, doc models.RFQTransactionPG) error
	Delete(shopID string, docNo string) error
}

type RFQTransactionConsumerService struct {
	repo IRFQTransactionPGRepository
}

func NewRFQTransactionService(repo IRFQTransactionPGRepository) IRFQTransactionConsumerService {
	return &RFQTransactionConsumerService{
		repo: repo,
	}
}

func (s *RFQTransactionConsumerService) Upsert(shopID string, docNo string, doc models.RFQTransactionPG) error {
	foundDocument, err := s.repo.Get(shopID, docNo)
	if err != nil && err.Error() != "record not found" {
		return err
	}
	if foundDocument == nil {
		return s.repo.Create(doc)
	} else {
		isEqual := foundDocument.CompareTo(&doc)
		if !isEqual {
			return s.repo.Update(shopID, docNo, doc)
		}
	}
	return nil
}

func (s *RFQTransactionConsumerService) Delete(shopID string, docNo string) error {
	err := s.repo.DeleteData(shopID, docNo, models.RFQTransactionPG{
		TransactionPG: models.TransactionPG{
			ShopIdentity: pkgModels.ShopIdentity{
				ShopID: shopID,
			},
			DocNo: docNo,
		},
	})
	return err
}
