package purchasereturn

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IPurchaseReturnTransactionPGRepository interface {
	Get(shopID string, docNo string) (*models.PurchaseReturnTransactionPG, error)
	Create(doc models.PurchaseReturnTransactionPG) error
	Update(shopID string, docNo string, doc models.PurchaseReturnTransactionPG) error
	Delete(shopID string, docNo string, doc models.PurchaseReturnTransactionPG) error
	DeleteData(shopID string, docNo string, doc models.PurchaseReturnTransactionPG) error
}

type PurchaseReturnTransactionPGRepository struct {
	pst microservice.IPersister
	repositories.ITransactionConsumerRepository[models.PurchaseReturnTransactionPG]
}

func NewPurchaseReturnTransactionPGRepository(pst microservice.IPersister) IPurchaseReturnTransactionPGRepository {

	repo := &PurchaseReturnTransactionPGRepository{
		pst: pst,
	}

	repo.ITransactionConsumerRepository = repositories.NewTransactionConsumerRepository[models.PurchaseReturnTransactionPG](pst)
	return repo
}

func (repo PurchaseReturnTransactionPGRepository) Create(doc models.PurchaseReturnTransactionPG) error {
	err := repo.pst.Create(&doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo PurchaseReturnTransactionPGRepository) Update(shopID string, docNo string, doc models.PurchaseReturnTransactionPG) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"shopid": shopID,
		"docno":  docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *PurchaseReturnTransactionPGRepository) DeleteData(shopID string, docNo string, doc models.PurchaseReturnTransactionPG) error {

	var details *[]models.PurchaseReturnTransactionDetailPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.PurchaseReturnTransactionDetailPG{}).Where(" shopid=? AND docno=?", shopID, docNo).Find(&details)
	for _, tmp := range *details {
		// mark delete
		tx.Delete(&models.PurchaseReturnTransactionDetailPG{}, tmp.ID)
	}

	err := tx.Delete(models.PurchaseReturnTransactionPG{}, map[string]interface{}{
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
