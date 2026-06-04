package rfq

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
)

type IRFQTransactionConsumerService interface {
	Upsert(holdingCode string, docNo string, doc models.RFQTransactionPG) error
	Delete(holdingCode string, docNo string) error
}

type RFQTransactionConsumerService struct {
	repo IRFQTransactionPGRepository
}

func NewRFQTransactionService(repo IRFQTransactionPGRepository) IRFQTransactionConsumerService {
	return &RFQTransactionConsumerService{
		repo: repo,
	}
}

func (s *RFQTransactionConsumerService) Upsert(holdingCode string, docNo string, doc models.RFQTransactionPG) error {
	foundDocument, err := s.repo.Get(holdingCode, docNo)
	if err != nil && err.Error() != "record not found" {
		return err
	}
	if foundDocument == nil {
		return s.repo.Create(doc)
	} else {
		isEqual := foundDocument.CompareTo(&doc)
		if !isEqual {
			return s.repo.Update(holdingCode, docNo, doc)
		}
	}
	return nil
}

func (s *RFQTransactionConsumerService) Delete(holdingCode string, docNo string) error {
	err := s.repo.DeleteData(holdingCode, docNo, models.RFQTransactionPG{
		TransactionPG: models.TransactionPG{
			HoldingCodeentity: pkgModels.HoldingCodeentity{
				HoldingCode: holdingCode,
			},
			DocNo: docNo,
		},
	})
	return err
}
