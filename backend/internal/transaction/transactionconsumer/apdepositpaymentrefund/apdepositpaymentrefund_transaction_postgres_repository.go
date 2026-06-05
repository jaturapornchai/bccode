package apdepositpaymentrefund

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IAPDepositPaymentRefundTransactionPGRepository interface {
	Get(holdingCode string, docNo string) (*models.APDepositPaymentRefundTransactionPG, error)
	Create(doc models.APDepositPaymentRefundTransactionPG) error
	Update(holdingCode string, docNo string, doc models.APDepositPaymentRefundTransactionPG) error
	Delete(holdingCode string, docNo string, doc models.APDepositPaymentRefundTransactionPG) error
	DeleteData(holdingCode string, docNo string, doc models.APDepositPaymentRefundTransactionPG) error
}

type APDepositPaymentRefundTransactionPGRepository struct {
	pst microservice.IPersister
	repositories.ITransactionConsumerRepository[models.APDepositPaymentRefundTransactionPG]
}

func NewAPDepositPaymentRefundTransactionPGRepository(pst microservice.IPersister) IAPDepositPaymentRefundTransactionPGRepository {

	repo := &APDepositPaymentRefundTransactionPGRepository{
		pst: pst,
	}

	repo.ITransactionConsumerRepository = repositories.NewTransactionConsumerRepository[models.APDepositPaymentRefundTransactionPG](pst)
	return repo
}

func (repo APDepositPaymentRefundTransactionPGRepository) Create(doc models.APDepositPaymentRefundTransactionPG) error {
	err := repo.pst.Create(&doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo APDepositPaymentRefundTransactionPGRepository) Update(holdingCode string, docNo string, doc models.APDepositPaymentRefundTransactionPG) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"holdingcode": holdingCode,
		"docno":       docNo,
	})
	if err != nil {
		return err
	}
	return nil
}

func (repo *APDepositPaymentRefundTransactionPGRepository) DeleteData(holdingCode string, docNo string, doc models.APDepositPaymentRefundTransactionPG) error {

	var details *[]models.APDepositPaymentRefundTransactionDetailPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.APDepositPaymentRefundTransactionDetailPG{}).Where(" holdingcode=? AND docno=?", holdingCode, docNo).Find(&details)
	for _, tmp := range *details {
		err := tx.Delete(&models.APDepositPaymentRefundTransactionDetailPG{}, tmp.ID).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err := tx.Delete(&models.APDepositPaymentRefundTransactionPG{}, map[string]interface{}{
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
