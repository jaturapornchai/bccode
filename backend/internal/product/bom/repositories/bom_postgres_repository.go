package repositories

import (
	"smlcloudplatform/internal/product/bom/models"
	"smlcloudplatform/pkg/microservice"

	"gorm.io/gorm"
)

type IBOMPostgresRepository interface {
	Get(holdingCode string, creditorCode string) (*models.ProductBarcodeBOMViewPG, error)
	Create(doc models.ProductBarcodeBOMViewPG) error
	Update(holdingCode string, creditorCode string, doc models.ProductBarcodeBOMViewPG) error
	Delete(holdingCode string, creditorCode string) error
}

type BOMPostgresRepository struct {
	pst microservice.IPersister
}

func NewBOMPostgresRepository(pst microservice.IPersister) IBOMPostgresRepository {
	return &BOMPostgresRepository{
		pst: pst,
	}
}

func (repo *BOMPostgresRepository) Get(holdingCode string, creditorCode string) (*models.ProductBarcodeBOMViewPG, error) {
	var result models.ProductBarcodeBOMViewPG
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

func (repo *BOMPostgresRepository) Create(doc models.ProductBarcodeBOMViewPG) error {
	err := repo.pst.Create(doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo *BOMPostgresRepository) Update(holdingCode string, creditorCode string, doc models.ProductBarcodeBOMViewPG) error {
	err := repo.pst.Update(&doc, map[string]interface{}{
		"holding_code": holdingCode,
		"code":         creditorCode,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *BOMPostgresRepository) Delete(holdingCode string, creditorCode string) error {
	err := repo.pst.Delete(&models.ProductBarcodeBOMViewPG{}, map[string]interface{}{
		"holding_code": holdingCode,
		"code":         creditorCode,
	})

	if err != nil {
		return err
	}
	return nil
}
