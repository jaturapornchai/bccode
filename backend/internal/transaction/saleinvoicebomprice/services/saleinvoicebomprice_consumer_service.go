package services

import (
	"smlcloudplatform/internal/transaction/saleinvoicebomprice/models"
	"smlcloudplatform/internal/transaction/saleinvoicebomprice/repositories"
)

type ISaleInvoiceBomPriceConsumerService interface {
	Upsert(shopID string, guidFixed string, doc models.SaleInvoiceBomPricePg) error
	Delete(shopID string, guidFixed string) error
}

type SaleInvoiceBomPriceConsumerService struct {
	repo repositories.ISaleInvoiceBomPricePostgresRepository
}

func NewSaleInvoiceBomPriceConsumerService(repo repositories.ISaleInvoiceBomPricePostgresRepository) ISaleInvoiceBomPriceConsumerService {
	return &SaleInvoiceBomPriceConsumerService{
		repo: repo,
	}
}

func (s *SaleInvoiceBomPriceConsumerService) Upsert(shopID string, guidFixed string, doc models.SaleInvoiceBomPricePg) error {
	findDoc, err := s.repo.Get(shopID, guidFixed)
	if err != nil || findDoc == nil {
		err = s.repo.Create(doc)
		if err != nil {
			return err
		}
	} else {

		isEqual := findDoc.CompareTo(&doc)

		if !isEqual {
			err = s.repo.Update(shopID, guidFixed, doc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *SaleInvoiceBomPriceConsumerService) Delete(shopID string, guidFixed string) error {
	err := s.repo.Delete(shopID, guidFixed)
	if err != nil {
		return err
	}
	return nil
}
