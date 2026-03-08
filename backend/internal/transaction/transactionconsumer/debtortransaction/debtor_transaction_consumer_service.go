package debtortransaction

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/pkg/microservice"

	"gorm.io/gorm"
)

type IDebtorTransactionConsumerService interface {
	Upsert(shopID string, docNo string, doc models.DebtorTransactionPG) error
	Delete(shopID string, docNo string) error
}

type DebtorTransactionConsumerService struct {
	repo IDebtorTransactionPGRepository
}

func NewDebtorTransactionService(
	pst microservice.IPersister,
	producer microservice.IProducer,
) IDebtorTransactionConsumerService {

	repo := NewDebtorTransactionPGRepository(pst)
	return &DebtorTransactionConsumerService{
		repo: repo,
	}
}

func (s *DebtorTransactionConsumerService) Upsert(shopID string, docNo string, doc models.DebtorTransactionPG) error {

	findTrx, err := s.repo.Get(shopID, docNo)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
	}

	if findTrx == nil {
		err = s.repo.Create(doc)
		if err != nil {
			return err
		}
	} else {

		isEqual := findTrx.CompareTo(&doc)

		if isEqual == false {
			err = s.repo.Update(shopID, docNo, doc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *DebtorTransactionConsumerService) Delete(shopID string, docNo string) error {
	err := s.repo.Delete(shopID, docNo, models.DebtorTransactionPG{
		ShopIdentity: pkgModels.ShopIdentity{
			ShopID: shopID,
		},
		DocNo: docNo,
	})
	if err != nil {
		return err
	}

	return nil
}
