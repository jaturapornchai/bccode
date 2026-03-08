package services

import (
	"smlcloudplatform/internal/debtaccount/creditor/models"
	"smlcloudplatform/internal/debtaccount/creditor/repositories"
)

type ICreditorConsumerService interface {
	Upsert(shopID string, guidFixed string, doc models.CreditorPG) error
	Delete(shopID string, guidFixed string) error
}

type CreditorConsumerService struct {
	repo repositories.ICreditorPostgresRepository
}

func NewCreditorConsumerService(repo repositories.ICreditorPostgresRepository) ICreditorConsumerService {
	return &CreditorConsumerService{
		repo: repo,
	}
}

func (s *CreditorConsumerService) Upsert(shopID string, guidFixed string, doc models.CreditorPG) error {
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

func (s *CreditorConsumerService) Delete(shopID string, guidFixed string) error {
	err := s.repo.Delete(shopID, guidFixed)
	if err != nil {
		return err
	}
	return nil
}
