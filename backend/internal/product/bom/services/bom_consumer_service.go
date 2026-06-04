package services

import (
	"smlcloudplatform/internal/product/bom/models"
	"smlcloudplatform/internal/product/bom/repositories"
)

type IBOMConsumerService interface {
	Upsert(holdingCode string, guidFixed string, doc models.ProductBarcodeBOMViewPG) error
	Delete(holdingCode string, guidFixed string) error
}

type BOMConsumerService struct {
	repo repositories.IBOMPostgresRepository
}

func NewBOMConsumerService(repo repositories.IBOMPostgresRepository) IBOMConsumerService {
	return &BOMConsumerService{
		repo: repo,
	}
}

func (s *BOMConsumerService) Upsert(holdingCode string, guidFixed string, doc models.ProductBarcodeBOMViewPG) error {
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

func (s *BOMConsumerService) Delete(holdingCode string, guidFixed string) error {
	err := s.repo.Delete(holdingCode, guidFixed)
	if err != nil {
		return err
	}
	return nil
}
