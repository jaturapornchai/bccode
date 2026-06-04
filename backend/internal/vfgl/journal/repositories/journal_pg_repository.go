package repositories

import (
	"smlcloudplatform/internal/vfgl/journal/models"
	"smlcloudplatform/pkg/microservice"

	"gorm.io/gorm/clause"
)

type IJournalPgRepository interface {
	CreateInBatch(docList []models.JournalPg) error
	Create(doc models.JournalPg) error
	Update(holdingCode string, docNo string, doc models.JournalPg) error
	Delete(holdingCode string, docNo string) error
	Get(holdingCode string, docNo string) (*models.JournalPg, error)
}

type JournalPgRepository struct {
	pst microservice.IPersister
}

func NewJournalPgRepository(pst microservice.IPersister) JournalPgRepository {
	return JournalPgRepository{
		pst: pst,
	}
}

func (repo JournalPgRepository) CreateInBatch(docList []models.JournalPg) error {
	err := repo.pst.CreateInBatch(docList, len(docList))
	if err != nil {
		return err
	}
	return nil
}

func (repo JournalPgRepository) Create(doc models.JournalPg) error {
	err := repo.pst.Create(doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo JournalPgRepository) Update(holdingCode string, docNo string, doc models.JournalPg) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"holding_code": holdingCode,
		"docno":        docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo JournalPgRepository) Delete(holdingCode string, docNo string) error {

	var details *[]models.JournalDetailPg
	tx := repo.pst.DBClient().Begin()
	tx.Model(&models.JournalDetailPg{}).Where(" holding_code=? AND docno=?", holdingCode, docNo).Find(&details)
	for _, tmp := range *details {
		// mark delete
		tx.Delete(&models.JournalDetailPg{}, tmp.ID)
	}

	var vats []*models.JournalVatPg
	tx.Model(&models.JournalVatPg{}).Where(" holding_code=? AND docno=?", holdingCode, docNo).Find(&vats)
	for _, tmp := range vats {
		// mark delete
		tx.Delete(&models.JournalVatPg{}, tmp.ID)
	}

	var taxes []*models.JournalTaxPg
	tx.Model(&models.JournalTaxPg{}).Where(" holding_code=? AND docno=?", holdingCode, docNo).Find(&taxes)
	for _, tmp := range taxes {
		// mark delete
		tx.Delete(&models.JournalTaxPg{}, tmp.ID)
	}

	err := tx.Delete(models.JournalPg{}, map[string]interface{}{
		"holding_code": holdingCode,
		"docno":        docNo,
	}).Error

	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (repo JournalPgRepository) Get(holdingCode string, docNo string) (*models.JournalPg, error) {

	var data models.JournalPg

	err := repo.pst.DBClient().Preload(clause.Associations).
		Where("holding_code=? AND docno=?", holdingCode, docNo).
		First(&data).Error
	if err != nil {
		return nil, err
	}

	return &data, nil
}
