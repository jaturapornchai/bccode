package costcenter

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/organization/costcenter/models"
)

type ICostCenterConsumerService interface {
	Upsert(holdingCode string, guidFixed string, doc models.CostCenterDoc) error
	Delete(holdingCode string, guidFixed string) error
}

type CostCenterConsumerService struct {
	pgRepo ICostCenterPGRepository
	chRepo ICostCenterCHRepository
}

func NewCostCenterConsumerService(pgRepo ICostCenterPGRepository, chRepo ICostCenterCHRepository) ICostCenterConsumerService {
	return &CostCenterConsumerService{
		pgRepo: pgRepo,
		chRepo: chRepo,
	}
}

func (s *CostCenterConsumerService) Upsert(holdingCode string, guidFixed string, doc models.CostCenterDoc) error {
	names := pkgModels.JSONB{}
	if doc.Names != nil {
		names = pkgModels.JSONB(*doc.Names)
	}

	pgDoc := models.CostCenterPg{
		HoldingCode: holdingCode,
		GuidFixed:   guidFixed,
		Code:        doc.Code,
		Names:       names,
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

func (s *CostCenterConsumerService) Delete(holdingCode string, guidFixed string) error {
	err := s.pgRepo.Delete(holdingCode, guidFixed)
	if err != nil {
		return err
	}

	if s.chRepo != nil {
		_ = s.chRepo.Delete(holdingCode, guidFixed)
	}

	return nil
}
