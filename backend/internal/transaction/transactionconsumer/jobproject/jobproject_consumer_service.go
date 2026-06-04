package jobproject

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/organization/jobproject/models"
)

type IJobProjectConsumerService interface {
	Upsert(holdingCode string, guidFixed string, doc models.JobProjectDoc) error
	Delete(holdingCode string, guidFixed string) error
}

type JobProjectConsumerService struct {
	pgRepo IJobProjectPGRepository
	chRepo IJobProjectCHRepository
}

func NewJobProjectConsumerService(pgRepo IJobProjectPGRepository, chRepo IJobProjectCHRepository) IJobProjectConsumerService {
	return &JobProjectConsumerService{
		pgRepo: pgRepo,
		chRepo: chRepo,
	}
}

func (s *JobProjectConsumerService) Upsert(holdingCode string, guidFixed string, doc models.JobProjectDoc) error {
	names := pkgModels.JSONB{}
	if doc.Names != nil {
		names = pkgModels.JSONB(*doc.Names)
	}

	pgDoc := models.JobProjectPg{
		HoldingCode: holdingCode,
		GuidFixed:   guidFixed,
		Code:        doc.Code,
		Names:       names,
		ParentCode:  doc.ParentCode,
	}

	found, err := s.pgRepo.Get(holdingCode, guidFixed)
	if err != nil {
		return err
	}

	if found == nil {
		err = s.pgRepo.Create(pgDoc)
	} else {
		err = s.pgRepo.Update(holdingCode, guidFixed, pgDoc)
	}
	if err != nil {
		return err
	}

	if s.chRepo != nil {
		_ = s.chRepo.Upsert(pgDoc)
	}

	return nil
}

func (s *JobProjectConsumerService) Delete(holdingCode string, guidFixed string) error {
	err := s.pgRepo.Delete(holdingCode, guidFixed)
	if err != nil {
		return err
	}

	if s.chRepo != nil {
		_ = s.chRepo.Delete(holdingCode, guidFixed)
	}

	return nil
}
