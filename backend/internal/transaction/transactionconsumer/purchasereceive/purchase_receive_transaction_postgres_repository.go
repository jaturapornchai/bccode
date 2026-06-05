package purchasereceive

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IPurchaseReceiveTransactionPGRepository interface {
	Get(holdingCode string, docNo string) (*models.PurchaseReceiveTransactionPG, error)
	Create(doc models.PurchaseReceiveTransactionPG) error
	Update(holdingCode string, docNo string, doc models.PurchaseReceiveTransactionPG) error
	Delete(holdingCode string, docNo string, doc models.PurchaseReceiveTransactionPG) error
	DeleteData(holdingCode string, docNo string, doc models.PurchaseReceiveTransactionPG) error
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

func (repo PurchaseReceiveTransactionPGRepository) Update(holdingCode string, docNo string, doc models.PurchaseReceiveTransactionPG) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"holdingcode": holdingCode,
		"docno":       docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *PurchaseReceiveTransactionPGRepository) DeleteData(holdingCode string, docNo string, doc models.PurchaseReceiveTransactionPG) error {

	var details *[]models.PurchaseReceiveTransactionDetailPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.PurchaseReceiveTransactionDetailPG{}).Where(" holdingcode=? AND docno=?", holdingCode, docNo).Find(&details)
	for _, tmp := range *details {
		// mark delete
		err := tx.Delete(&models.PurchaseReceiveTransactionDetailPG{}, tmp.ID).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err := tx.Delete(&models.PurchaseReceiveTransactionPG{}, map[string]interface{}{
		"holdingcode": holdingCode,
		"docno":       docNo,
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
