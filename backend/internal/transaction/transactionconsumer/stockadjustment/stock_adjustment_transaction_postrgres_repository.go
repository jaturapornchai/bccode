package stockadjustment

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IStockAdjustmentTransactionPGRepository interface {
	Get(holdingCode string, docNo string) (*models.StockAdjustmentTransactionPG, error)
	Create(doc models.StockAdjustmentTransactionPG) error
	Update(holdingCode string, docNo string, doc models.StockAdjustmentTransactionPG) error
	Delete(holdingCode string, docNo string, doc models.StockAdjustmentTransactionPG) error
	DeleteData(holdingCode string, docNo string, doc models.StockAdjustmentTransactionPG) error
}

type StockAdjustmentTransactionPGRepository struct {
	pst microservice.IPersister
	repositories.ITransactionConsumerRepository[models.StockAdjustmentTransactionPG]
}

func NewStockAdjustmentTransactionPGRepository(pst microservice.IPersister) IStockAdjustmentTransactionPGRepository {

	repo := &StockAdjustmentTransactionPGRepository{
		pst: pst,
	}

	repo.ITransactionConsumerRepository = repositories.NewTransactionConsumerRepository[models.StockAdjustmentTransactionPG](pst)
	return repo
}

func (repo StockAdjustmentTransactionPGRepository) Create(doc models.StockAdjustmentTransactionPG) error {
	err := repo.pst.Create(&doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo StockAdjustmentTransactionPGRepository) Update(holdingCode string, docNo string, doc models.StockAdjustmentTransactionPG) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"holdingcode": holdingCode,
		"docno":       docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *StockAdjustmentTransactionPGRepository) DeleteData(holdingCode string, docNo string, doc models.StockAdjustmentTransactionPG) error {

	var details *[]models.StockAdjustmentTransactionDetailPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.StockAdjustmentTransactionDetailPG{}).Where(" holdingcode=? AND docno=?", holdingCode, docNo).Find(&details)
	for _, tmp := range *details {
		// mark delete
		tx.Delete(&models.StockAdjustmentTransactionDetailPG{}, tmp.ID)
	}

	err := tx.Delete(models.StockAdjustmentTransactionPG{}, map[string]interface{}{
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
