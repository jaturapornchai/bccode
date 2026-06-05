package repositories

import (
	debtorModels "smlcloudplatform/internal/debtaccount/debtor/models"
	"smlcloudplatform/pkg/microservice"

	"gorm.io/gorm"
)

type IDebtorPostgresRepository interface {
	Get(holdingCode string, code string) (*debtorModels.DebtorPG, error)
	Create(doc debtorModels.DebtorPG) error
	Update(holdingCode string, code string, doc debtorModels.DebtorPG) error
	Delete(holdingCode string, code string) error
}

type DebtorPostgresRepository struct {
	pst microservice.IPersister
}

func NewDebtorPostgresRepository(pst microservice.IPersister) IDebtorPostgresRepository {
	return &DebtorPostgresRepository{
		pst: pst,
	}
}

func (repo *DebtorPostgresRepository) Get(holdingCode string, code string) (*debtorModels.DebtorPG, error) {
	var result debtorModels.DebtorPG
	_, err := repo.pst.First(&result, "holdingcode=? AND code=?", holdingCode, code)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return &result, nil
}

func (repo *DebtorPostgresRepository) Create(doc debtorModels.DebtorPG) error {
	err := repo.pst.Create(doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo *DebtorPostgresRepository) Update(holdingCode string, code string, doc debtorModels.DebtorPG) error {
	err := repo.pst.Update(&doc, map[string]interface{}{
		"holdingcode": holdingCode,
		"code":        code,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *DebtorPostgresRepository) Delete(holdingCode string, code string) error {
	err := repo.pst.Delete(&debtorModels.DebtorPG{}, map[string]interface{}{
		"holdingcode": holdingCode,
		"code":        code,
	})
	if err != nil {
		return err
	}
	return nil
}
