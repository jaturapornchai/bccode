package stockpickupproduct

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IStockPickupTransactionPGRepository interface {
	Get(shopID string, docNo string) (*models.StockPickUpTransactionPG, error)
	Create(doc models.StockPickUpTransactionPG) error
	Update(shopID string, docNo string, doc models.StockPickUpTransactionPG) error
	Delete(shopID string, docNo string, doc models.StockPickUpTransactionPG) error
	DeleteData(shopID string, docNo string, doc models.StockPickUpTransactionPG) error
}

type StockPickupTransactionPGRepository struct {
	pst microservice.IPersister
	repositories.ITransactionConsumerRepository[models.StockPickUpTransactionPG]
}

func NewStockPickupTransactionPGRepository(pst microservice.IPersister) IStockPickupTransactionPGRepository {

	repo := &StockPickupTransactionPGRepository{
		pst: pst,
	}

	repo.ITransactionConsumerRepository = repositories.NewTransactionConsumerRepository[models.StockPickUpTransactionPG](pst)
	return repo
}

func (repo StockPickupTransactionPGRepository) Create(doc models.StockPickUpTransactionPG) error {
	err := repo.pst.Create(&doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo StockPickupTransactionPGRepository) Update(shopID string, docNo string, doc models.StockPickUpTransactionPG) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"shopid": shopID,
		"docno":  docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *StockPickupTransactionPGRepository) DeleteData(shopID string, docNo string, doc models.StockPickUpTransactionPG) error {

	var details *[]models.StockPickUpTransactionDetailPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.StockPickUpTransactionDetailPG{}).Where(" shopid=? AND docno=?", shopID, docNo).Find(&details)
	for _, tmp := range *details {
		// mark delete
		tx.Delete(&models.StockPickUpTransactionDetailPG{}, tmp.ID)
	}

	err := tx.Delete(models.StockPickUpTransactionPG{}, map[string]interface{}{
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
