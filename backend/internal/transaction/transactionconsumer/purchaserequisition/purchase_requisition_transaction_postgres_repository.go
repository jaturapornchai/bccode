package purchaserequisition

import (
	"errors"
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IPurchaseRequisitionTransactionPGRepository interface {
	Get(holdingCode string, docNo string) (*models.PurchaseRequisitionTransactionPG, error)
	Create(doc models.PurchaseRequisitionTransactionPG) error
	Update(holdingCode string, docNo string, doc models.PurchaseRequisitionTransactionPG) error
	DeleteData(holdingCode string, docNo string, doc models.PurchaseRequisitionTransactionPG) error
}

type PurchaseRequisitionTransactionPGRepository struct {
	pst microservice.IPersister
	repositories.ITransactionConsumerRepository[models.PurchaseRequisitionTransactionPG]
}

func NewPurchaseRequisitionTransactionRepository(pst microservice.IPersister) IPurchaseRequisitionTransactionPGRepository {
	repo := &PurchaseRequisitionTransactionPGRepository{
		pst: pst,
	}
	repo.ITransactionConsumerRepository = repositories.NewTransactionConsumerRepository[models.PurchaseRequisitionTransactionPG](pst)
	return repo
}

func (repo PurchaseRequisitionTransactionPGRepository) Create(doc models.PurchaseRequisitionTransactionPG) error {
	if doc.DocNo == "" {
		return errors.New("DocNo cannot be empty")
	}
	err := repo.pst.Create(&doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo PurchaseRequisitionTransactionPGRepository) Update(holdingCode string, docNo string, doc models.PurchaseRequisitionTransactionPG) error {
	err := repo.pst.Update(&doc, map[string]interface{}{
		"holdingcode": holdingCode,
		"docno":       docNo,
	})
	if err != nil {
		return err
	}
	return nil
}

func (repo *PurchaseRequisitionTransactionPGRepository) DeleteData(holdingCode string, docNo string, doc models.PurchaseRequisitionTransactionPG) error {
	var details *[]models.PurchaseRequisitionDetailTransactionPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.PurchaseRequisitionDetailTransactionPG{}).Where(" holdingcode=? AND docno=?", holdingCode, docNo).Find(&details)
	for _, tmp := range *details {
		err := tx.Delete(&models.PurchaseRequisitionDetailTransactionPG{}, tmp.ID).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err := tx.Delete(models.PurchaseRequisitionTransactionPG{}, map[string]interface{}{
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
