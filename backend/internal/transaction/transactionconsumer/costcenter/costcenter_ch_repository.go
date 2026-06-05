package costcenter

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/organization/costcenter/models"
	"smlcloudplatform/pkg/microservice"
)

type ICostCenterCHRepository interface {
	Upsert(doc models.CostCenterPg) error
	Delete(holdingCode string, guidFixed string) error
}

type CostCenterCHRepository struct {
	pst microservice.IPersisterClickHouse
}

func NewCostCenterCHRepository(pst microservice.IPersisterClickHouse) *CostCenterCHRepository {
	if pst == nil {
		return nil
	}
	return &CostCenterCHRepository{pst: pst}
}

func (repo *CostCenterCHRepository) Upsert(doc models.CostCenterPg) error {
	conn := repo.pst.Conn()
	err := conn.Exec(context.Background(),
		`INSERT INTO organization_cost_center (holdingcode, guidfixed, code, names) VALUES (?, ?, ?, ?)`,
		doc.HoldingCode, doc.GuidFixed, doc.Code, doc.Names,
	)
	if err != nil {
		return fmt.Errorf("costcenter ch upsert error: %w", err)
	}
	return nil
}

func (repo *CostCenterCHRepository) Delete(holdingCode string, guidFixed string) error {
	conn := repo.pst.Conn()
	err := conn.Exec(context.Background(),
		`ALTER TABLE organization_cost_center DELETE WHERE holdingcode=? AND guidfixed=?`,
		holdingCode, guidFixed,
	)
	if err != nil {
		return fmt.Errorf("costcenter ch delete error: %w", err)
	}
	return nil
}
