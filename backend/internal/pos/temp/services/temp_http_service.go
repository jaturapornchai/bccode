package services

import (
	"smlcloudplatform/internal/pos/temp/repositories"
	"time"
)

type IPOSTempService interface {
	SaveTemp(holdingCode string, branchCode string, doc string) error
	InfoTemp(holdingCode string, branchCode string) (string, error)
	DeleteTemp(holdingCode string, branchCode string) error
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

func (svc POSTempService) SaveTemp(holdingCode string, branchCode string, doc string) error {
	err := svc.repo.Save(holdingCode, branchCode, doc, time.Hour*24*7)
	if err != nil {
		return err
	}

	return nil
}

func (svc POSTempService) InfoTemp(holdingCode string, branchCode string) (string, error) {

	result, err := svc.repo.Get(holdingCode, branchCode)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (svc POSTempService) DeleteTemp(holdingCode string, branchCode string) error {

	err := svc.repo.Delete(holdingCode, branchCode)
	if err != nil {
		return err
	}

	return nil
}
