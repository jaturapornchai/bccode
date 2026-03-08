package purchaseorder

import (
	"errors"
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IPurchaseOrderTransactionPGRepository interface {
	Get(shopID string, docNo string) (*models.PurchaseOrderTransactionPG, error)
	Create(doc models.PurchaseOrderTransactionPG) error
	Update(shopID string, docNo string, doc models.PurchaseOrderTransactionPG) error
	DeleteData(shopID string, docNo string, doc models.PurchaseOrderTransactionPG) error
}

type PurchaseOrderTransactionPGRepository struct {
	pst microservice.IPersister
	repositories.ITransactionConsumerRepository[models.PurchaseOrderTransactionPG]
}

func NewPurchaseOrderTransactionRepository(pst microservice.IPersister) IPurchaseOrderTransactionPGRepository {

	repo := &PurchaseOrderTransactionPGRepository{
		pst: pst,
	}

	repo.ITransactionConsumerRepository = repositories.NewTransactionConsumerRepository[models.PurchaseOrderTransactionPG](pst)
	return repo
}

func (repo PurchaseOrderTransactionPGRepository) Create(doc models.PurchaseOrderTransactionPG) error {

	if doc.DocNo == "" {
		return errors.New("DocNo cannot be empty")
	}

	err := repo.pst.Create(&doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo PurchaseOrderTransactionPGRepository) Update(shopID string, docNo string, doc models.PurchaseOrderTransactionPG) error {
	err := repo.pst.Update(&doc, map[string]interface{}{
		"shopid": shopID,
		"docno":  docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *PurchaseOrderTransactionPGRepository) DeleteData(shopID string, docNo string, doc models.PurchaseOrderTransactionPG) error {

	var details *[]models.PurchaseOrderDetailTransactionPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.PurchaseOrderDetailTransactionPG{}).Where(" shopid=? AND docno=?", shopID, docNo).Find(&details)
	for _, tmp := range *details {
		// mark delete
		err := tx.Delete(&models.PurchaseOrderDetailTransactionPG{}, tmp.ID).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err := tx.Delete(models.PurchaseOrderTransactionPG{}, map[string]interface{}{
		"shopid": shopID,
		"docno":  docNo,
	}).Error

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
