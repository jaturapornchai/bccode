package apadvancepayment

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IAPAdvancePaymentTransactionPGRepository interface {
	Get(shopID string, docNo string) (*models.APAdvancePaymentTransactionPG, error)
	Create(doc models.APAdvancePaymentTransactionPG) error
	Update(shopID string, docNo string, doc models.APAdvancePaymentTransactionPG) error
	Delete(shopID string, docNo string, doc models.APAdvancePaymentTransactionPG) error
	DeleteData(shopID string, docNo string, doc models.APAdvancePaymentTransactionPG) error
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

func (repo APAdvancePaymentTransactionPGRepository) Update(shopID string, docNo string, doc models.APAdvancePaymentTransactionPG) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"shopid": shopID,
		"docno":  docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *APAdvancePaymentTransactionPGRepository) DeleteData(shopID string, docNo string, doc models.APAdvancePaymentTransactionPG) error {

	var details *[]models.APAdvancePaymentTransactionDetailPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.APAdvancePaymentTransactionDetailPG{}).Where(" shopid=? AND docno=?", shopID, docNo).Find(&details)
	for _, tmp := range *details {
		// mark delete
		err := tx.Delete(&models.APAdvancePaymentTransactionDetailPG{}, tmp.ID).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err := tx.Delete(&models.APAdvancePaymentTransactionPG{}, "shopid=? AND docno=?", shopID, docNo).Error
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
