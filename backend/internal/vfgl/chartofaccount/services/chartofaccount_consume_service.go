package services

import (
	"encoding/json"
	"smlcloudplatform/internal/vfgl/chartofaccount/models"
	"smlcloudplatform/internal/vfgl/chartofaccount/repositories"

	"gorm.io/gorm"
)

type IChartOfAccountConsumeService interface {
	Create(doc models.ChartOfAccountDoc) error
	Update(shopID string, accountCode string, doc models.ChartOfAccountDoc) error
	Delete(shopID string, guid string) error
	SaveInBatch(docList []models.ChartOfAccountDoc) error
	Upsert(models.ChartOfAccountDoc) (*models.ChartOfAccountPG, error)
}

type ChartOfAccountConsumeService struct {
	repo repositories.IChartOfAccountPgRepository
}

func NewChartOfAccountConsumeService(repo repositories.IChartOfAccountPgRepository) ChartOfAccountConsumeService {
	return ChartOfAccountConsumeService{
		repo: repo,
	}
}

func (svc *ChartOfAccountConsumeService) Create(doc models.ChartOfAccountDoc) (*models.ChartOfAccountPG, error) {
	pgDoc := models.ChartOfAccountPG{}

	tmpJsonDoc, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(tmpJsonDoc), &pgDoc)
	if err != nil {
		return nil, err
	}

	err = svc.repo.Create(pgDoc)

	if err != nil {
		return nil, err
	}
	return &pgDoc, nil
}

func (svc *ChartOfAccountConsumeService) Update(shopID string, accountCode string, doc models.ChartOfAccountDoc) error {
	pgDoc := models.ChartOfAccountPG{}

	tmpJsonDoc, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(tmpJsonDoc), &pgDoc)
	if err != nil {
		return err
	}

	err = svc.repo.Update(shopID, accountCode, pgDoc)

	if err != nil {
		return err
	}
	return nil
}

func (svc *ChartOfAccountConsumeService) Delete(shopID string, accountCode string) error {
	err := svc.repo.Delete(shopID, accountCode)

	if err != nil {
		return err
	}
	return nil
}

func (svc *ChartOfAccountConsumeService) SaveInBatch(docList []models.ChartOfAccountDoc) error {
	pgDocList := []models.ChartOfAccountPG{}

	for _, doc := range docList {
		tmpJsonDoc, err := json.Marshal(doc)
		if err != nil {
			return err
		}
		tmpDoc := models.ChartOfAccountPG{}
		err = json.Unmarshal([]byte(tmpJsonDoc), &tmpDoc)
		if err != nil {
			return err
		}
		pgDocList = append(pgDocList, tmpDoc)
	}

	err := svc.repo.CreateInBatch(pgDocList)
	if err != nil {
		return err
	}

	return nil
}

func (svc *ChartOfAccountConsumeService) Upsert(shopID string, doc models.ChartOfAccountDoc) (*models.ChartOfAccountPG, error) {

	// get
	data, err := svc.repo.Get(shopID, doc.AccountCode)
	if err == gorm.ErrRecordNotFound {
		if data, err = svc.Create(doc); err != nil {
			return nil, err
		}
	} else if data != nil {
		data.AccountName = doc.AccountName
		data.FinancialStatements = doc.FinancialStatements
		data.AccountBalanceType = int16(doc.AccountBalanceType)
		data.AccountCategory = int16(doc.AccountCategory)
		data.AccountLevel = int16(doc.AccountLevel)
		data.AccountGroup = doc.AccountGroup
		data.ConsolidateAccountCode = doc.ConsolidateAccountCode

		if err = svc.repo.Update(shopID, doc.AccountCode, *data); err != nil {
			return nil, err
		}
	}

	return data, nil
}
