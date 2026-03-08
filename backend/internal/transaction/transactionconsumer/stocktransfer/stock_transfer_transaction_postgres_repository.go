package stocktransfer

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IStockTransferTransactionPGRepository interface {
	Get(shopID string, docNo string) (*models.StockTransferTransactionPG, error)
	Create(doc models.StockTransferTransactionPG) error
	Update(shopID string, docNo string, doc models.StockTransferTransactionPG) error
	Delete(shopID string, docNo string, doc models.StockTransferTransactionPG) error
	DeleteData(shopID string, docNo string, doc models.StockTransferTransactionPG) error
}

type StockTransferTransactionPGRepository struct {
	pst microservice.IPersister
	repositories.ITransactionConsumerRepository[models.StockTransferTransactionPG]
}

func NewStockAdjustmentTransactionPGRepository(pst microservice.IPersister) IStockTransferTransactionPGRepository {

	repo := &StockTransferTransactionPGRepository{
		pst: pst,
	}

	repo.ITransactionConsumerRepository = repositories.NewTransactionConsumerRepository[models.StockTransferTransactionPG](pst)
	return repo
}

func (repo StockTransferTransactionPGRepository) Create(doc models.StockTransferTransactionPG) error {
	err := repo.pst.Create(&doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo StockTransferTransactionPGRepository) Update(shopID string, docNo string, doc models.StockTransferTransactionPG) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"shopid": shopID,
		"docno":  docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *StockTransferTransactionPGRepository) DeleteData(shopID string, docNo string, doc models.StockTransferTransactionPG) error {

	var details *[]models.StockTransferTransactionDetailPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.StockTransferTransactionDetailPG{}).Where(" shopid=? AND docno=?", shopID, docNo).Find(&details)
	for _, tmp := range *details {
		// mark delete
		tx.Delete(&models.StockTransferTransactionDetailPG{}, tmp.ID)
	}

	err := tx.Delete(models.StockTransferTransactionPG{}, map[string]interface{}{
		"shopid": shopID,
		"docno":  docNo,
	}).Error

	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}
