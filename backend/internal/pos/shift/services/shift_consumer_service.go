package services

import (
	"smlcloudplatform/internal/pos/shift/models"
	"smlcloudplatform/internal/pos/shift/repositories"
)

type IShiftConsumerService interface {
	Upsert(holdingCode string, guidFixed string, doc models.ShiftPG) error
	Delete(holdingCode string, guidFixed string) error
}

type ShiftConsumerService struct {
	repo repositories.IShiftPostgresRepository
}

func NewShiftConsumerService(repo repositories.IShiftPostgresRepository) IShiftConsumerService {
	return &ShiftConsumerService{
		repo: repo,
	}
}

func (s *ShiftConsumerService) Upsert(holdingCode string, guidFixed string, doc models.ShiftPG) error {
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

func (s *ShiftConsumerService) Delete(holdingCode string, guidFixed string) error {
	err := s.repo.Delete(holdingCode, guidFixed)
	if err != nil {
		return err
	}
	return nil
}
