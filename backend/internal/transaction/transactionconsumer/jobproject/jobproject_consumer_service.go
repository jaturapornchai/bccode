package jobproject

import (
	"smlcloudplatform/internal/organization/jobproject/models"
	pkgModels "smlcloudplatform/internal/models"
)

type IJobProjectConsumerService interface {
	Upsert(shopID string, guidFixed string, doc models.JobProjectDoc) error
	Delete(shopID string, guidFixed string) error
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

func (s *JobProjectConsumerService) Upsert(shopID string, guidFixed string, doc models.JobProjectDoc) error {
	names := pkgModels.JSONB{}
	if doc.Names != nil {
		names = pkgModels.JSONB(*doc.Names)
	}

	pgDoc := models.JobProjectPg{
		ShopID:     shopID,
		GuidFixed:  guidFixed,
		Code:       doc.Code,
		Names:      names,
		ParentCode: doc.ParentCode,
	}

	found, err := s.pgRepo.Get(shopID, guidFixed)
	if err != nil {
		return err
	}

	if found == nil {
		err = s.pgRepo.Create(pgDoc)
	} else {
		err = s.pgRepo.Update(shopID, guidFixed, pgDoc)
	}
	if err != nil {
		return err
	}

	if s.chRepo != nil {
		_ = s.chRepo.Upsert(pgDoc)
	}

	return nil
}

func (s *JobProjectConsumerService) Delete(shopID string, guidFixed string) error {
	err := s.pgRepo.Delete(shopID, guidFixed)
	if err != nil {
		return err
	}

	if s.chRepo != nil {
		_ = s.chRepo.Delete(shopID, guidFixed)
	}

	return nil
}
