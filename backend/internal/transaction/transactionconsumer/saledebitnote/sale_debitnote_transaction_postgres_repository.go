package saledebitnote

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type ISaleDebitNoteTransactionPGRepository interface {
	Get(holdingCode string, docNo string) (*models.SaleDebitNoteTransactionPG, error)
	Create(doc models.SaleDebitNoteTransactionPG) error
	Update(holdingCode string, docNo string, doc models.SaleDebitNoteTransactionPG) error
	Delete(holdingCode string, docNo string, doc models.SaleDebitNoteTransactionPG) error
	DeleteData(holdingCode string, docNo string, doc models.SaleDebitNoteTransactionPG) error
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

func (repo SaleDebitNoteTransactionPGRepository) Update(holdingCode string, docNo string, doc models.SaleDebitNoteTransactionPG) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"holdingcode": holdingCode,
		"docno":       docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *SaleDebitNoteTransactionPGRepository) DeleteData(holdingCode string, docNo string, doc models.SaleDebitNoteTransactionPG) error {

	var details *[]models.SaleDebitNoteTransactionDetailPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.SaleDebitNoteTransactionDetailPG{}).Where(" holdingcode=? AND docno=?", holdingCode, docNo).Find(&details)
	for _, tmp := range *details {
		// mark delete
		tx.Delete(&models.SaleDebitNoteTransactionDetailPG{}, tmp.ID)
	}

	err := tx.Delete(&models.SaleDebitNoteTransactionPG{}, " holdingcode=? AND docno=?", holdingCode, docNo).Error
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
