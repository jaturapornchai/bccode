package repositories

import (
	creditorModels "smlcloudplatform/internal/debtaccount/creditor/models"
	"smlcloudplatform/pkg/microservice"

	"gorm.io/gorm"
)

type ICreditorPostgresRepository interface {
	Get(holdingCode string, creditorCode string) (*creditorModels.CreditorPG, error)
	Create(doc creditorModels.CreditorPG) error
	Update(holdingCode string, creditorCode string, doc creditorModels.CreditorPG) error
	Delete(holdingCode string, creditorCode string) error
}

type CreditorPostgresRepository struct {
	pst microservice.IPersister
}

func NewCreditorPostgresRepository(pst microservice.IPersister) ICreditorPostgresRepository {
	return &CreditorPostgresRepository{
		pst: pst,
	}
}

func (repo *CreditorPostgresRepository) Get(holdingCode string, creditorCode string) (*creditorModels.CreditorPG, error) {
	var result creditorModels.CreditorPG
	_, err := repo.pst.First(&result, "holding_code=? AND code=?", holdingCode, creditorCode)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return &result, nil
}

func (repo *CreditorPostgresRepository) Create(doc creditorModels.CreditorPG) error {
	err := repo.pst.Create(doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo *CreditorPostgresRepository) Update(holdingCode string, creditorCode string, doc creditorModels.CreditorPG) error {
	err := repo.pst.Update(&doc, map[string]interface{}{
		"holding_code": holdingCode,
		"code":         creditorCode,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *CreditorPostgresRepository) Delete(holdingCode string, creditorCode string) error {
	err := repo.pst.Delete(&creditorModels.CreditorPG{}, map[string]interface{}{
		"holding_code": holdingCode,
		"code":         creditorCode,
	})

	if err != nil {
		return err
	}
	return nil
}
