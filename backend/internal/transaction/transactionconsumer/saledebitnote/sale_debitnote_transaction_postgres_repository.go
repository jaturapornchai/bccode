package saledebitnote

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type ISaleDebitNoteTransactionPGRepository interface {
	Get(shopID string, docNo string) (*models.SaleDebitNoteTransactionPG, error)
	Create(doc models.SaleDebitNoteTransactionPG) error
	Update(shopID string, docNo string, doc models.SaleDebitNoteTransactionPG) error
	Delete(shopID string, docNo string, doc models.SaleDebitNoteTransactionPG) error
	DeleteData(shopID string, docNo string, doc models.SaleDebitNoteTransactionPG) error
}

type SaleDebitNoteTransactionPGRepository struct {
	pst microservice.IPersister
	repositories.ITransactionConsumerRepository[models.SaleDebitNoteTransactionPG]
}

func NewSaleDebitNoteTransactionPGRepository(pst microservice.IPersister) ISaleDebitNoteTransactionPGRepository {

	repo := &SaleDebitNoteTransactionPGRepository{
		pst: pst,
	}

	repo.ITransactionConsumerRepository = repositories.NewTransactionConsumerRepository[models.SaleDebitNoteTransactionPG](pst)
	return repo
}

func (repo SaleDebitNoteTransactionPGRepository) Create(doc models.SaleDebitNoteTransactionPG) error {

	err := repo.pst.Create(&doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo SaleDebitNoteTransactionPGRepository) Update(shopID string, docNo string, doc models.SaleDebitNoteTransactionPG) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"shopid": shopID,
		"docno":  docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *SaleDebitNoteTransactionPGRepository) DeleteData(shopID string, docNo string, doc models.SaleDebitNoteTransactionPG) error {

	var details *[]models.SaleDebitNoteTransactionDetailPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.SaleDebitNoteTransactionDetailPG{}).Where(" shopid=? AND docno=?", shopID, docNo).Find(&details)
	for _, tmp := range *details {
		// mark delete
		tx.Delete(&models.SaleDebitNoteTransactionDetailPG{}, tmp.ID)
	}

	err := tx.Delete(&models.SaleDebitNoteTransactionPG{}, " shopid=? AND docno=?", shopID, docNo).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
