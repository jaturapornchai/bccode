package apadvancepayment

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IAPAdvancePaymentTransactionPGRepository interface {
	Get(holdingCode string, docNo string) (*models.APAdvancePaymentTransactionPG, error)
	Create(doc models.APAdvancePaymentTransactionPG) error
	Update(holdingCode string, docNo string, doc models.APAdvancePaymentTransactionPG) error
	Delete(holdingCode string, docNo string, doc models.APAdvancePaymentTransactionPG) error
	DeleteData(holdingCode string, docNo string, doc models.APAdvancePaymentTransactionPG) error
}

type APAdvancePaymentTransactionPGRepository struct {
	pst microservice.IPersister
	repositories.ITransactionConsumerRepository[models.APAdvancePaymentTransactionPG]
}

func NewAPAdvancePaymentTransactionPGRepository(pst microservice.IPersister) IAPAdvancePaymentTransactionPGRepository {

	repo := &APAdvancePaymentTransactionPGRepository{
		pst: pst,
	}

	repo.ITransactionConsumerRepository = repositories.NewTransactionConsumerRepository[models.APAdvancePaymentTransactionPG](pst)
	return repo
}

func (repo APAdvancePaymentTransactionPGRepository) Create(doc models.APAdvancePaymentTransactionPG) error {
	err := repo.pst.Create(&doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo APAdvancePaymentTransactionPGRepository) Update(holdingCode string, docNo string, doc models.APAdvancePaymentTransactionPG) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"holdingcode": holdingCode,
		"docno":       docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *APAdvancePaymentTransactionPGRepository) DeleteData(holdingCode string, docNo string, doc models.APAdvancePaymentTransactionPG) error {

	var details *[]models.APAdvancePaymentTransactionDetailPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.APAdvancePaymentTransactionDetailPG{}).Where(" holdingcode=? AND docno=?", holdingCode, docNo).Find(&details)
	for _, tmp := range *details {
		// mark delete
		err := tx.Delete(&models.APAdvancePaymentTransactionDetailPG{}, tmp.ID).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err := tx.Delete(&models.APAdvancePaymentTransactionPG{}, "holdingcode=? AND docno=?", holdingCode, docNo).Error
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
