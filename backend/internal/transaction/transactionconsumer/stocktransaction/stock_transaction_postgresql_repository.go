package stocktransaction

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/pkg/microservice"

	"gorm.io/gorm/clause"
)

type IStockTransactionPGRepository interface {
	Get(holdingCode string, docNo string) (*models.StockTransaction, error)
	Create(doc models.StockTransaction) error
	Update(holdingCode string, docNo string, doc models.StockTransaction) error
	Delete(holdingCode string, docNo string) error
}

func NewStockTransactionPGRepository(pst microservice.IPersister) IStockTransactionPGRepository {
	return &StockTransactionPGRepository{
		pst: pst,
	}
}

type StockTransactionPGRepository struct {
	pst microservice.IPersister
}

func (repo *StockTransactionPGRepository) Get(holdingCode string, docNo string) (*models.StockTransaction, error) {
	var data models.StockTransaction

	err := repo.pst.DBClient().Preload(clause.Associations).
		Where("holdingcode=? AND docno=?", holdingCode, docNo).
		First(&data).Error
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (repo *StockTransactionPGRepository) Create(doc models.StockTransaction) error {
	err := repo.pst.Create(doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo *StockTransactionPGRepository) Update(holdingCode string, docNo string, doc models.StockTransaction) error {
	err := repo.pst.Update(&doc, map[string]interface{}{
		"holdingcode": holdingCode,
		"docno":       docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *StockTransactionPGRepository) Delete(holdingCode string, docNo string) error {
	var details *[]models.StockTransactionDetail
	tx := repo.pst.DBClient().Begin()
	tx.Model(&models.StockTransactionDetail{}).Where(" holdingcode=? AND docno=?", holdingCode, docNo).Find(&details)
	for _, tmp := range *details {
		// mark delete
		tx.Delete(&models.StockTransactionDetail{}, tmp.ID)
	}

	err := tx.Delete(models.StockTransaction{}, map[string]interface{}{
		"holdingcode": holdingCode,
		"docno":       docNo,
	}).Error

	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}
