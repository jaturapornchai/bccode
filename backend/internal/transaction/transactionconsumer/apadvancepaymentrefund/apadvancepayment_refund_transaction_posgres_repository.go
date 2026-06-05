package apadvancepaymentrefund

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IAPAdvancePaymentRefundTransactionPGRepository interface {
	Get(holdingCode string, docNo string) (*models.APAdvancePaymentRefundTransactionPG, error)
	Create(doc models.APAdvancePaymentRefundTransactionPG) error
	Update(holdingCode string, docNo string, doc models.APAdvancePaymentRefundTransactionPG) error
	Delete(holdingCode string, docNo string, doc models.APAdvancePaymentRefundTransactionPG) error
	DeleteData(holdingCode string, docNo string, doc models.APAdvancePaymentRefundTransactionPG) error
}

type APAdvancePaymentRefundTransactionPGRepository struct {
	pst microservice.IPersister
	repositories.ITransactionConsumerRepository[models.APAdvancePaymentRefundTransactionPG]
}

func NewAPAdvancePaymentRefundTransactionPGRepository(pst microservice.IPersister) IAPAdvancePaymentRefundTransactionPGRepository {

	repo := &APAdvancePaymentRefundTransactionPGRepository{
		pst: pst,
	}

	repo.ITransactionConsumerRepository = repositories.NewTransactionConsumerRepository[models.APAdvancePaymentRefundTransactionPG](pst)
	return repo
}

func (repo APAdvancePaymentRefundTransactionPGRepository) Create(doc models.APAdvancePaymentRefundTransactionPG) error {
	err := repo.pst.Create(&doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo APAdvancePaymentRefundTransactionPGRepository) Update(holdingCode string, docNo string, doc models.APAdvancePaymentRefundTransactionPG) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"holdingcode": holdingCode,
		"docno":       docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *APAdvancePaymentRefundTransactionPGRepository) DeleteData(holdingCode string, docNo string, doc models.APAdvancePaymentRefundTransactionPG) error {

	var details *[]models.APAdvancePaymentRefundTransactionDetailPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.APAdvancePaymentRefundTransactionDetailPG{}).Where(" holdingcode=? AND docno=?", holdingCode, docNo).Find(&details)
	for _, tmp := range *details {
		err := tx.Delete(&models.PurchaseReceiveTransactionDetailPG{}, tmp.ID).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err := tx.Delete(&models.APAdvancePaymentRefundTransactionPG{}, map[string]interface{}{
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
