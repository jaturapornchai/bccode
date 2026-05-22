package jobproject

import (
	"smlcloudplatform/internal/organization/jobproject/models"
	"smlcloudplatform/pkg/microservice"

	"gorm.io/gorm"
)

type IJobProjectPGRepository interface {
	Get(shopID string, guidFixed string) (*models.JobProjectPg, error)
	Create(doc models.JobProjectPg) error
	Update(shopID string, guidFixed string, doc models.JobProjectPg) error
	Delete(shopID string, guidFixed string) error
}

type JobProjectPGRepository struct {
	pst microservice.IPersister
}

func NewJobProjectPGRepository(pst microservice.IPersister) *JobProjectPGRepository {
	return &JobProjectPGRepository{pst: pst}
}

func (repo *JobProjectPGRepository) Get(shopID string, guidFixed string) (*models.JobProjectPg, error) {
	var result models.JobProjectPg
	_, err := repo.pst.First(&result, "shopid=? AND guidfixed=?", shopID, guidFixed)
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

func (repo *JobProjectPGRepository) Update(shopID string, guidFixed string, doc models.JobProjectPg) error {
	err := repo.pst.Update(&doc, map[string]interface{}{
		"shopid":    shopID,
		"guid_fixed": guidFixed,
	})
	if err != nil {
		return err
	}
	return nil
}

func (repo *JobProjectPGRepository) Delete(shopID string, guidFixed string) error {
	err := repo.pst.Delete(&models.JobProjectPg{}, map[string]interface{}{
		"shopid":    shopID,
		"guid_fixed": guidFixed,
	})
	if err != nil {
		return err
	}
	return nil
}
