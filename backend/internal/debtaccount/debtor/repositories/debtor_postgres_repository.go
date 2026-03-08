package repositories

import (
	debtorModels "smlcloudplatform/internal/debtaccount/debtor/models"
	"smlcloudplatform/pkg/microservice"

	"gorm.io/gorm"
)

type IDebtorPostgresRepository interface {
	Get(shopID string, code string) (*debtorModels.DebtorPG, error)
	Create(doc debtorModels.DebtorPG) error
	Update(shopID string, code string, doc debtorModels.DebtorPG) error
	Delete(shopID string, code string) error
}

type DebtorPostgresRepository struct {
	pst microservice.IPersister
}

func NewDebtorPostgresRepository(pst microservice.IPersister) IDebtorPostgresRepository {
	return &DebtorPostgresRepository{
		pst: pst,
	}
}

func (repo *DebtorPostgresRepository) Get(shopID string, code string) (*debtorModels.DebtorPG, error) {
	var result debtorModels.DebtorPG
	_, err := repo.pst.First(&result, "shopid=? AND code=?", shopID, code)
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

func (repo *DebtorPostgresRepository) Update(shopID string, code string, doc debtorModels.DebtorPG) error {
	err := repo.pst.Update(&doc, map[string]interface{}{
		"shopid": shopID,
		"code":   code,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *DebtorPostgresRepository) Delete(shopID string, code string) error {
	err := repo.pst.Delete(&debtorModels.DebtorPG{}, map[string]interface{}{
		"shopid": shopID,
		"code":   code,
	})
	if err != nil {
		return err
	}
	return nil
}
