package costcenter

import (
	"smlcloudplatform/internal/organization/costcenter/models"
	"smlcloudplatform/pkg/microservice"

	"gorm.io/gorm"
)

type ICostCenterPGRepository interface {
	Get(shopID string, guidFixed string) (*models.CostCenterPg, error)
	Create(doc models.CostCenterPg) error
	Update(shopID string, guidFixed string, doc models.CostCenterPg) error
	Delete(shopID string, guidFixed string) error
}

type CostCenterPGRepository struct {
	pst microservice.IPersister
}

func NewCostCenterPGRepository(pst microservice.IPersister) *CostCenterPGRepository {
	return &CostCenterPGRepository{pst: pst}
}

func (repo *CostCenterPGRepository) Get(shopID string, guidFixed string) (*models.CostCenterPg, error) {
	var result models.CostCenterPg
	_, err := repo.pst.First(&result, "shopid=? AND guidfixed=?", shopID, guidFixed)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}

func (repo *CostCenterPGRepository) Create(doc models.CostCenterPg) error {
	err := repo.pst.Create(&doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo *CostCenterPGRepository) Update(shopID string, guidFixed string, doc models.CostCenterPg) error {
	err := repo.pst.Update(&doc, map[string]interface{}{
		"shopid":    shopID,
		"guidfixed": guidFixed,
	})
	if err != nil {
		return err
	}
	return nil
}

func (repo *CostCenterPGRepository) Delete(shopID string, guidFixed string) error {
	err := repo.pst.Delete(&models.CostCenterPg{}, map[string]interface{}{
		"shopid":    shopID,
		"guidfixed": guidFixed,
	})
	if err != nil {
		return err
	}
	return nil
}
