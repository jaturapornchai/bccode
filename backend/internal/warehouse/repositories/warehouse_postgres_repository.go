package repositories

import (
	"smlcloudplatform/internal/warehouse/models"
	"smlcloudplatform/pkg/microservice"

	"gorm.io/gorm/clause"
)

type IWarehousePGRepository interface {
	Get(shopID string, guidFixed string) (models.WarehousePG, error)
	Create(doc models.WarehousePG) error
	Update(shopID string, guidFixed string, doc models.WarehousePG) error
	Delete(shopID string, guidFixed string) error
}

type WarehousePGRepository struct {
	pst microservice.IPersister
}

func NewWarehousePGRepository(pst microservice.IPersister) IWarehousePGRepository {

	repo := &WarehousePGRepository{
		pst: pst,
	}

	return repo
}

func (repo WarehousePGRepository) Get(shopID string, guidFixed string) (models.WarehousePG, error) {

	var data models.WarehousePG
	err := repo.pst.DBClient().Preload(clause.Associations).
		Where("shopid=? AND guidfixed=?", shopID, guidFixed).
		First(&data).Error

	if err != nil {
		return models.WarehousePG{}, err
	}

	return data, nil
}

func (repo WarehousePGRepository) Create(doc models.WarehousePG) error {
	err := repo.pst.Create(&doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo WarehousePGRepository) Update(shopID string, guidFixed string, doc models.WarehousePG) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"shopid":    shopID,
		"guidfixed": guidFixed,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *WarehousePGRepository) Delete(shopID string, guidFixed string) error {

	err := repo.pst.Delete(models.WarehousePG{}, map[string]interface{}{
		"shopid":    shopID,
		"guidfixed": guidFixed,
	})

	if err != nil {
		return err
	}

	return nil
}
