package services

import (
	"smlcloudplatform/internal/logger"
	"smlcloudplatform/internal/warehouse/models"
	"smlcloudplatform/internal/warehouse/repositories"
)

type IWarehouseConsumerService interface {
	Upsert(shopID string, docNo string, doc models.WarehousePG) error
	Delete(shopID string, docNo string) error
}

type WarehouseConsumerService struct {
	repo repositories.IWarehousePGRepository
}

func NewWarehouseConsumerService(repo repositories.IWarehousePGRepository) IWarehouseConsumerService {
	return &WarehouseConsumerService{
		repo: repo,
	}
}

func (s *WarehouseConsumerService) Upsert(shopID string, guidFixed string, doc models.WarehousePG) error {
	foundDocument, err := s.repo.Get(shopID, guidFixed)

	if err != nil && err.Error() != "record not found" {
		return err
	}

	if foundDocument.GuidFixed == "" {
		err = s.repo.Create(doc)
		if err != nil {
			return err
		}
	} else {

		isEqual := foundDocument.CompareTo(&doc)

		if !isEqual {
			err = s.repo.Update(shopID, guidFixed, doc)
			if err != nil {
				return err
			}
		} else {
			logger.GetLogger().Debug("Doc is equal, skip update")
		}
	}

	return nil
}

func (s *WarehouseConsumerService) Delete(shopID string, guidFixed string) error {

	err := s.repo.Delete(shopID, guidFixed)
	if err != nil {
		return err
	}
	return nil
}
