package saleinvoicereturn

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type ISaleInvoiceReturnTransactionPGRepository interface {
	Get(holdingCode string, docNo string) (*models.SaleInvoiceReturnTransactionPG, error)
	Create(doc models.SaleInvoiceReturnTransactionPG) error
	Update(holdingCode string, docNo string, doc models.SaleInvoiceReturnTransactionPG) error
	Delete(holdingCode string, docNo string, doc models.SaleInvoiceReturnTransactionPG) error
	DeleteData(holdingCode string, docNo string, doc models.SaleInvoiceReturnTransactionPG) error
}

type SaleInvoiceReturnTransactionPGRepository struct {
	pst microservice.IPersister
	repositories.ITransactionConsumerRepository[models.SaleInvoiceReturnTransactionPG]
}

func NewSaleInvoiceReturnTransactionPGRepository(pst microservice.IPersister) ISaleInvoiceReturnTransactionPGRepository {

	repo := &SaleInvoiceReturnTransactionPGRepository{
		pst: pst,
	}

	repo.ITransactionConsumerRepository = repositories.NewTransactionConsumerRepository[models.SaleInvoiceReturnTransactionPG](pst)
	return repo
}

func (repo SaleInvoiceReturnTransactionPGRepository) Create(doc models.SaleInvoiceReturnTransactionPG) error {
	err := repo.pst.Create(&doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo SaleInvoiceReturnTransactionPGRepository) Update(holdingCode string, docNo string, doc models.SaleInvoiceReturnTransactionPG) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"holdingcode": holdingCode,
		"docno":       docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *SaleInvoiceReturnTransactionPGRepository) DeleteData(holdingCode string, docNo string, doc models.SaleInvoiceReturnTransactionPG) error {

	var details *[]models.SaleInvoiceReturnTransactionDetailPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.SaleInvoiceReturnTransactionDetailPG{}).Where(" holdingcode=? AND docno=?", holdingCode, docNo).Find(&details)
	for _, tmp := range *details {
		// mark delete
		tx.Delete(&models.SaleInvoiceReturnTransactionDetailPG{}, tmp.ID)
	}

	err := tx.Delete(models.SaleInvoiceReturnTransactionPG{}, map[string]interface{}{
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
