package purchasereceive

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IPurchaseReceiveTransactionPGRepository interface {
	Get(shopID string, docNo string) (*models.PurchaseReceiveTransactionPG, error)
	Create(doc models.PurchaseReceiveTransactionPG) error
	Update(shopID string, docNo string, doc models.PurchaseReceiveTransactionPG) error
	Delete(shopID string, docNo string, doc models.PurchaseReceiveTransactionPG) error
	DeleteData(shopID string, docNo string, doc models.PurchaseReceiveTransactionPG) error
}

type PurchaseReceiveTransactionPGRepository struct {
	pst microservice.IPersister
	repositories.ITransactionConsumerRepository[models.PurchaseReceiveTransactionPG]
}

func NewPurchaseReceiveTransactionPGRepository(pst microservice.IPersister) IPurchaseReceiveTransactionPGRepository {

	repo := &PurchaseReceiveTransactionPGRepository{
		pst: pst,
	}

	repo.ITransactionConsumerRepository = repositories.NewTransactionConsumerRepository[models.PurchaseReceiveTransactionPG](pst)
	return repo
}

func (repo PurchaseReceiveTransactionPGRepository) Create(doc models.PurchaseReceiveTransactionPG) error {
	err := repo.pst.Create(&doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo PurchaseReceiveTransactionPGRepository) Update(shopID string, docNo string, doc models.PurchaseReceiveTransactionPG) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"shopid": shopID,
		"docno":  docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *PurchaseReceiveTransactionPGRepository) DeleteData(shopID string, docNo string, doc models.PurchaseReceiveTransactionPG) error {

	var details *[]models.PurchaseReceiveTransactionDetailPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.PurchaseReceiveTransactionDetailPG{}).Where(" shopid=? AND docno=?", shopID, docNo).Find(&details)
	for _, tmp := range *details {
		// mark delete
		err := tx.Delete(&models.PurchaseReceiveTransactionDetailPG{}, tmp.ID).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err := tx.Delete(&models.PurchaseReceiveTransactionPG{}, map[string]interface{}{
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
