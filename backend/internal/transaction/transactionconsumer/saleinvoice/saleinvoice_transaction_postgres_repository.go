package saleinvoice

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type ISaleInvoiceTransactionPGRepository interface {
	Get(shopID string, docNo string) (*models.SaleInvoiceTransactionPG, error)
	Create(doc models.SaleInvoiceTransactionPG) error
	Update(shopID string, docNo string, doc models.SaleInvoiceTransactionPG) error
	Delete(shopID string, docNo string, doc models.SaleInvoiceTransactionPG) error
	DeleteData(shopID string, docNo string, doc models.SaleInvoiceTransactionPG) error
}

type SaleInvoiceTransactionPGRepository struct {
	pst microservice.IPersister
	repositories.ITransactionConsumerRepository[models.SaleInvoiceTransactionPG]
}

func NewSaleInvoiceTransactionPGRepository(pst microservice.IPersister) ISaleInvoiceTransactionPGRepository {

	repo := &SaleInvoiceTransactionPGRepository{
		pst: pst,
	}

	repo.ITransactionConsumerRepository = repositories.NewTransactionConsumerRepository[models.SaleInvoiceTransactionPG](pst)
	return repo
}

func (repo SaleInvoiceTransactionPGRepository) Create(doc models.SaleInvoiceTransactionPG) error {
	err := repo.pst.Create(&doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo SaleInvoiceTransactionPGRepository) Update(shopID string, docNo string, doc models.SaleInvoiceTransactionPG) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"shopid": shopID,
		"docno":  docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *SaleInvoiceTransactionPGRepository) DeleteData(shopID string, docNo string, doc models.SaleInvoiceTransactionPG) error {

	var details *[]models.SaleInvoiceTransactionDetailPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.SaleInvoiceTransactionDetailPG{}).Where(" shopid=? AND docno=?", shopID, docNo).Find(&details)
	for _, tmp := range *details {
		// mark delete
		tx.Delete(&models.SaleInvoiceTransactionDetailPG{}, tmp.ID)
	}

	err := tx.Delete(models.SaleInvoiceTransactionPG{}, map[string]interface{}{
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
