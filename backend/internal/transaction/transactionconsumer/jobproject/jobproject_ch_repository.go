package jobproject

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/organization/jobproject/models"
	"smlcloudplatform/pkg/microservice"
)

type IJobProjectCHRepository interface {
	Upsert(doc models.JobProjectPg) error
	Delete(holdingCode string, guidFixed string) error
}

type JobProjectCHRepository struct {
	pst microservice.IPersisterClickHouse
}

func NewJobProjectCHRepository(pst microservice.IPersisterClickHouse) *JobProjectCHRepository {
	if pst == nil {
		return nil
	}
	return &JobProjectCHRepository{pst: pst}
}

func (repo *JobProjectCHRepository) Upsert(doc models.JobProjectPg) error {
	conn := repo.pst.Conn()
	err := conn.Exec(context.Background(),
		`INSERT INTO organization_job_project (holdingcode, guidfixed, code, names, parentcode) VALUES (?, ?, ?, ?, ?)`,
		doc.HoldingCode, doc.GuidFixed, doc.Code, doc.Names, doc.ParentCode,
	)
	if err != nil {
		return fmt.Errorf("jobproject ch upsert error: %w", err)
	}
	return nil
}

func (repo *JobProjectCHRepository) Delete(holdingCode string, guidFixed string) error {
	conn := repo.pst.Conn()
	err := conn.Exec(context.Background(),
		`ALTER TABLE organization_job_project DELETE WHERE holdingcode=? AND guidfixed=?`,
		holdingCode, guidFixed,
	)
	if err != nil {
		return fmt.Errorf("jobproject ch delete error: %w", err)
	}
	return nil
}
