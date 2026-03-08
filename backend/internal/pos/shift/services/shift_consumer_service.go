package services

import (
	"smlcloudplatform/internal/pos/shift/models"
	"smlcloudplatform/internal/pos/shift/repositories"
)

type IShiftConsumerService interface {
	Upsert(shopID string, guidFixed string, doc models.ShiftPG) error
	Delete(shopID string, guidFixed string) error
}

type ShiftConsumerService struct {
	repo repositories.IShiftPostgresRepository
}

func NewShiftConsumerService(repo repositories.IShiftPostgresRepository) IShiftConsumerService {
	return &ShiftConsumerService{
		repo: repo,
	}
}

func (s *ShiftConsumerService) Upsert(shopID string, guidFixed string, doc models.ShiftPG) error {
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

func (s *ShiftConsumerService) Delete(shopID string, guidFixed string) error {
	err := s.repo.Delete(shopID, guidFixed)
	if err != nil {
		return err
	}
	return nil
}
