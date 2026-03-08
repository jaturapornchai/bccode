package stockreturnproduct

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IStockReturnTransactionPGRepository interface {
	Get(shopID string, docNo string) (*models.StockReturnProductTransactionPG, error)
	Create(doc models.StockReturnProductTransactionPG) error
	Update(shopID string, docNo string, doc models.StockReturnProductTransactionPG) error
	Delete(shopID string, docNo string, doc models.StockReturnProductTransactionPG) error
	DeleteData(shopID string, docNo string, doc models.StockReturnProductTransactionPG) error
}

type StockReturnTransactionPGRepository struct {
	pst microservice.IPersister
	repositories.ITransactionConsumerRepository[models.StockReturnProductTransactionPG]
}

func NewStockReturnTransactionPGRepository(pst microservice.IPersister) IStockReturnTransactionPGRepository {

	repo := &StockReturnTransactionPGRepository{
		pst: pst,
	}

	repo.ITransactionConsumerRepository = repositories.NewTransactionConsumerRepository[models.StockReturnProductTransactionPG](pst)
	return repo
}

func (repo StockReturnTransactionPGRepository) Create(doc models.StockReturnProductTransactionPG) error {
	err := repo.pst.Create(&doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo StockReturnTransactionPGRepository) Update(shopID string, docNo string, doc models.StockReturnProductTransactionPG) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"shopid": shopID,
		"docno":  docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *StockReturnTransactionPGRepository) DeleteData(shopID string, docNo string, doc models.StockReturnProductTransactionPG) error {

	var details *[]models.StockReturnProductTransactionDetailPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.StockReturnProductTransactionDetailPG{}).Where(" shopid=? AND docno=?", shopID, docNo).Find(&details)
	for _, tmp := range *details {
		// mark delete
		tx.Delete(&models.StockReturnProductTransactionDetailPG{}, tmp.ID)
	}

	err := tx.Delete(models.StockReturnProductTransactionPG{}, map[string]interface{}{
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
