package services

import (
	"smlcloudplatform/internal/pos/temp/repositories"
	"time"
)

type IPOSTempService interface {
	SaveTemp(shopID string, branchCode string, doc string) error
	InfoTemp(shopID string, branchCode string) (string, error)
	DeleteTemp(shopID string, branchCode string) error
}

type POSTempService struct {
	repo repositories.ICacheRepository
}

func NewPOSTempService(repo repositories.ICacheRepository) *POSTempService {

	insSvc := &POSTempService{
		repo: repo,
	}
	return insSvc
}

func (svc POSTempService) SaveTemp(shopID string, branchCode string, doc string) error {
	err := svc.repo.Save(shopID, branchCode, doc, time.Hour*24*7)
	if err != nil {
		return err
	}

	return nil
}

func (svc POSTempService) InfoTemp(shopID string, branchCode string) (string, error) {

	result, err := svc.repo.Get(shopID, branchCode)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (svc POSTempService) DeleteTemp(shopID string, branchCode string) error {

	err := svc.repo.Delete(shopID, branchCode)
	if err != nil {
		return err
	}

	return nil
}
