package jobproject

import (
	"smlcloudplatform/internal/organization/jobproject/models"
	"smlcloudplatform/pkg/microservice"

	"gorm.io/gorm"
)

type IJobProjectPGRepository interface {
	Get(holdingCode string, guidFixed string) (*models.JobProjectPg, error)
	Create(doc models.JobProjectPg) error
	Update(holdingCode string, guidFixed string, doc models.JobProjectPg) error
	Delete(holdingCode string, guidFixed string) error
}

type JobProjectPGRepository struct {
	pst microservice.IPersister
}

func NewJobProjectPGRepository(pst microservice.IPersister) *JobProjectPGRepository {
	return &JobProjectPGRepository{pst: pst}
}

func (repo *JobProjectPGRepository) Get(holdingCode string, guidFixed string) (*models.JobProjectPg, error) {
	var result models.JobProjectPg
	_, err := repo.pst.First(&result, "holdingcode=? AND guidfixed=?", holdingCode, guidFixed)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}

func (repo *JobProjectPGRepository) Create(doc models.JobProjectPg) error {
	err := repo.pst.Create(&doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo *JobProjectPGRepository) Update(holdingCode string, guidFixed string, doc models.JobProjectPg) error {
	err := repo.pst.Update(&doc, map[string]interface{}{
		"holdingcode": holdingCode,
		"guidfixed":   guidFixed,
	})
	if err != nil {
		return err
	}
	return nil
}

func (repo *JobProjectPGRepository) Delete(holdingCode string, guidFixed string) error {
	err := repo.pst.Delete(&models.JobProjectPg{}, map[string]interface{}{
		"holdingcode": holdingCode,
		"guidfixed":   guidFixed,
	})
	if err != nil {
		return err
	}
	return nil
}
