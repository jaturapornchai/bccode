package stockreceiveproduct

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IStockReceiveTransactionPGRepository interface {
	Get(shopID string, docNo string) (*models.StockReceiveProductTransactionPG, error)
	Create(doc models.StockReceiveProductTransactionPG) error
	Update(shopID string, docNo string, doc models.StockReceiveProductTransactionPG) error
	Delete(shopID string, docNo string, doc models.StockReceiveProductTransactionPG) error
	DeleteData(shopID string, docNo string, doc models.StockReceiveProductTransactionPG) error
}

type StockReceiveTransactionPGRepository struct {
	pst microservice.IPersister
	repositories.ITransactionConsumerRepository[models.StockReceiveProductTransactionPG]
}

func NewStockReceiveTransactionPGRepository(pst microservice.IPersister) IStockReceiveTransactionPGRepository {

	repo := &StockReceiveTransactionPGRepository{
		pst: pst,
	}

	repo.ITransactionConsumerRepository = repositories.NewTransactionConsumerRepository[models.StockReceiveProductTransactionPG](pst)
	return repo
}

func (repo StockReceiveTransactionPGRepository) Create(doc models.StockReceiveProductTransactionPG) error {
	err := repo.pst.Create(&doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo StockReceiveTransactionPGRepository) Update(shopID string, docNo string, doc models.StockReceiveProductTransactionPG) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"shopid": shopID,
		"docno":  docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *StockReceiveTransactionPGRepository) DeleteData(shopID string, docNo string, doc models.StockReceiveProductTransactionPG) error {

	var details *[]models.StockReceiveProductTransactionDetailPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.StockReceiveProductTransactionDetailPG{}).Where(" shopid=? AND docno=?", shopID, docNo).Find(&details)
	for _, tmp := range *details {
		// mark delete
		tx.Delete(&models.StockReceiveProductTransactionDetailPG{}, tmp.ID)
	}

	err := tx.Delete(models.StockReceiveProductTransactionPG{}, map[string]interface{}{
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
