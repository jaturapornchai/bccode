package repositories

import (
	shiftModels "smlcloudplatform/internal/pos/shift/models"
	"smlcloudplatform/pkg/microservice"

	"gorm.io/gorm"
)

type IShiftPostgresRepository interface {
	Get(shopID string, shiftCode string) (*shiftModels.ShiftPG, error)
	Create(doc shiftModels.ShiftPG) error
	Update(shopID string, shiftCode string, doc shiftModels.ShiftPG) error
	Delete(shopID string, shiftCode string) error
}

type ShiftPostgresRepository struct {
	pst microservice.IPersister
}

func NewShiftPostgresRepository(pst microservice.IPersister) IShiftPostgresRepository {
	return &ShiftPostgresRepository{
		pst: pst,
	}
}

func (repo *ShiftPostgresRepository) Get(shopID string, shiftCode string) (*shiftModels.ShiftPG, error) {
	var result shiftModels.ShiftPG
	_, err := repo.pst.First(&result, "shopid=? AND guidfixed=?", shopID, shiftCode)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return &result, nil
}

func (repo *ShiftPostgresRepository) Create(doc shiftModels.ShiftPG) error {
	err := repo.pst.Create(doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo *ShiftPostgresRepository) Update(shopID string, shiftCode string, doc shiftModels.ShiftPG) error {
	err := repo.pst.Update(&doc, map[string]interface{}{
		"shopid":    shopID,
		"guidfixed": shiftCode,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *ShiftPostgresRepository) Delete(shopID string, shiftCode string) error {
	err := repo.pst.Delete(&shiftModels.ShiftPG{}, map[string]interface{}{
		"shopid":    shopID,
		"guidfixed": shiftCode,
	})

	if err != nil {
		return err
	}
	return nil
}
