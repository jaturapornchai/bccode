package services

import (
	"smlcloudplatform/internal/debtaccount/debtor/models"
	"smlcloudplatform/internal/debtaccount/debtor/repositories"
)

type IDebtorConsumerService interface {
	Upsert(holdingCode string, code string, doc models.DebtorPG) error
	Delete(holdingCode string, code string) error
}

type DebtorConsumerService struct {
	repo repositories.IDebtorPostgresRepository
}

func NewDebtorConsumerService(repo repositories.IDebtorPostgresRepository) IDebtorConsumerService {
	return &DebtorConsumerService{
		repo: repo,
	}
}

func (s *DebtorConsumerService) Upsert(holdingCode string, code string, doc models.DebtorPG) error {
	findDoc, err := s.repo.Get(holdingCode, code)
	if err != nil || findDoc == nil {
		err = s.repo.Create(doc)
		if err != nil {
			return err
		}
	} else {

		isEqual := findDoc.CompareTo(&doc)

		if !isEqual {
			err = s.repo.Update(holdingCode, code, doc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *DebtorConsumerService) Delete(holdingCode string, code string) error {
	err := s.repo.Delete(holdingCode, code)
	if err != nil {
		return err
	}
	return nil
}
