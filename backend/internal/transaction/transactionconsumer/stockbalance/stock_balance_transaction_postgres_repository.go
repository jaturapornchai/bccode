package stockbalance

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IStockReceiveTransactionPGRepository interface {
	Get(holdingCode string, docNo string) (*models.StockBalanceTransactionPG, error)
	Create(doc models.StockBalanceTransactionPG) error
	Update(holdingCode string, docNo string, doc models.StockBalanceTransactionPG) error
	Delete(holdingCode string, docNo string, doc models.StockBalanceTransactionPG) error
	DeleteData(holdingCode string, docNo string, doc models.StockBalanceTransactionPG) error
}

type StockReceiveTransactionPGRepository struct {
	pst microservice.IPersister
	repositories.ITransactionConsumerRepository[models.StockBalanceTransactionPG]
}

func NewStockReceiveTransactionPGRepository(pst microservice.IPersister) IStockReceiveTransactionPGRepository {

	repo := &StockReceiveTransactionPGRepository{
		pst: pst,
	}

	repo.ITransactionConsumerRepository = repositories.NewTransactionConsumerRepository[models.StockBalanceTransactionPG](pst)
	return repo
}

func (repo StockReceiveTransactionPGRepository) Create(doc models.StockBalanceTransactionPG) error {
	err := repo.pst.Create(&doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo StockReceiveTransactionPGRepository) Update(holdingCode string, docNo string, doc models.StockBalanceTransactionPG) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"holdingcode": holdingCode,
		"docno":       docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *StockReceiveTransactionPGRepository) DeleteData(holdingCode string, docNo string, doc models.StockBalanceTransactionPG) error {

	var details *[]models.StockBalanceTransactionDetailPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.StockBalanceTransactionDetailPG{}).Where(" holdingcode=? AND docno=?", holdingCode, docNo).Find(&details)
	for _, tmp := range *details {
		// mark delete
		tx.Delete(&models.StockBalanceTransactionDetailPG{}, tmp.ID)
	}

	err := tx.Delete(models.StockBalanceTransactionPG{}, map[string]interface{}{
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
