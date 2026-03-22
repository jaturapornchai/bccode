package costcenter

import (
	"smlcloudplatform/internal/organization/costcenter/models"
	pkgModels "smlcloudplatform/internal/models"
)

type ICostCenterConsumerService interface {
	Upsert(shopID string, guidFixed string, doc models.CostCenterDoc) error
	Delete(shopID string, guidFixed string) error
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

func (s *CostCenterConsumerService) Upsert(shopID string, guidFixed string, doc models.CostCenterDoc) error {
	names := pkgModels.JSONB{}
	if doc.Names != nil {
		names = pkgModels.JSONB(*doc.Names)
	}

	pgDoc := models.CostCenterPg{
		ShopID:    shopID,
		GuidFixed: guidFixed,
		Code:      doc.Code,
		Names:     names,
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

func (s *CostCenterConsumerService) Delete(shopID string, guidFixed string) error {
	err := s.pgRepo.Delete(shopID, guidFixed)
	if err != nil {
		return err
	}

	if s.chRepo != nil {
		_ = s.chRepo.Delete(shopID, guidFixed)
	}

	return nil
}
