package services

import (
	"smlcloudplatform/internal/transaction/saleinvoicebomprice/models"
	"smlcloudplatform/internal/transaction/saleinvoicebomprice/repositories"
)

type ISaleInvoiceBomPriceConsumerService interface {
	Upsert(holdingCode string, guidFixed string, doc models.SaleInvoiceBomPricePg) error
	Delete(holdingCode string, guidFixed string) error
}

type SaleInvoiceBomPriceConsumerService struct {
	repo repositories.ISaleInvoiceBomPricePostgresRepository
}

func NewSaleInvoiceBomPriceConsumerService(repo repositories.ISaleInvoiceBomPricePostgresRepository) ISaleInvoiceBomPriceConsumerService {
	return &SaleInvoiceBomPriceConsumerService{
		repo: repo,
	}
}

func (s *SaleInvoiceBomPriceConsumerService) Upsert(holdingCode string, guidFixed string, doc models.SaleInvoiceBomPricePg) error {
	findDoc, err := s.repo.Get(holdingCode, guidFixed)
	if err != nil || findDoc == nil {
		err = s.repo.Create(doc)
		if err != nil {
			return err
		}
	} else {

		isEqual := findDoc.CompareTo(&doc)

		if !isEqual {
			err = s.repo.Update(holdingCode, guidFixed, doc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *SaleInvoiceBomPriceConsumerService) Delete(holdingCode string, guidFixed string) error {
	err := s.repo.Delete(holdingCode, guidFixed)
	if err != nil {
		return err
	}
	return nil
}
