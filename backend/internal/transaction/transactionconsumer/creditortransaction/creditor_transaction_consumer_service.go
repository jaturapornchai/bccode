package creditortransaction

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/pkg/microservice"

	"gorm.io/gorm"
)

type ICreditorTransactionConsumerService interface {
	Upsert(shopID string, docNo string, doc models.CreditorTransactionPG) error
	Delete(shopID string, docNo string) error
}

type CreditorTransactionConsumerService struct {
	repo ICreditorTransactionPGRepository
}

func NewCreditorTransactionConsumerService(
	pst microservice.IPersister,
	producer microservice.IProducer,
) ICreditorTransactionConsumerService {

	pgRepo := NewCreditorTransactionPGRepository(pst)

	return &CreditorTransactionConsumerService{
		repo: pgRepo,
	}
}

func (s *CreditorTransactionConsumerService) Upsert(shopID string, docNo string, doc models.CreditorTransactionPG) error {

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

func (s *CreditorTransactionConsumerService) Delete(shopID string, docNo string) error {

	err := s.repo.Delete(shopID, docNo, models.CreditorTransactionPG{
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
