package repositories

import (
	"smlcloudplatform/internal/vfgl/chartofaccount/models"
	"smlcloudplatform/pkg/microservice"
)

type IChartOfAccountPgRepository interface {
	CreateInBatch(docList []models.ChartOfAccountPG) error
	Create(doc models.ChartOfAccountPG) error
	Update(shopID string, accountCode string, doc models.ChartOfAccountPG) error
	Delete(shopID string, accountCode string) error
	Get(shopID string, accountCode string) (*models.ChartOfAccountPG, error)
}

type ChartOfAccountPgRepository struct {
	pst microservice.IPersister
}

func NewChartOfAccountPgRepository(pst microservice.IPersister) ChartOfAccountPgRepository {
	return ChartOfAccountPgRepository{
		pst: pst,
	}
}

func (repo ChartOfAccountPgRepository) CreateInBatch(docList []models.ChartOfAccountPG) error {
	err := repo.pst.CreateInBatchOnConflict(docList, len(docList))
	if err != nil {
		return err
	}
	return nil
}

func (repo ChartOfAccountPgRepository) Create(doc models.ChartOfAccountPG) error {
	err := repo.pst.Create(doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo ChartOfAccountPgRepository) Update(shopID string, accountCode string, doc models.ChartOfAccountPG) error {
	err := repo.pst.Update(&doc, map[string]interface{}{
		"shopid":      shopID,
		"accountcode": accountCode,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo ChartOfAccountPgRepository) Delete(shopID string, accountCode string) error {
	err := repo.pst.Delete(models.ChartOfAccountPG{}, map[string]interface{}{
		"shopid":      shopID,
		"accountcode": accountCode,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo ChartOfAccountPgRepository) Get(shopID string, accountCode string) (*models.ChartOfAccountPG, error) {
	var result models.ChartOfAccountPG
	_, err := repo.pst.First(&result, "shopid=? AND accountcode=?", shopID, accountCode)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
