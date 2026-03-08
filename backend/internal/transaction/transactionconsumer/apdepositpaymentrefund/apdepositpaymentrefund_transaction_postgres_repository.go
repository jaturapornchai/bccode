package apdepositpaymentrefund

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IAPDepositPaymentRefundTransactionPGRepository interface {
	Get(shopID string, docNo string) (*models.APDepositPaymentRefundTransactionPG, error)
	Create(doc models.APDepositPaymentRefundTransactionPG) error
	Update(shopID string, docNo string, doc models.APDepositPaymentRefundTransactionPG) error
	Delete(shopID string, docNo string, doc models.APDepositPaymentRefundTransactionPG) error
	DeleteData(shopID string, docNo string, doc models.APDepositPaymentRefundTransactionPG) error
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

func (repo APDepositPaymentRefundTransactionPGRepository) Update(shopID string, docNo string, doc models.APDepositPaymentRefundTransactionPG) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"shopid": shopID,
		"docno":  docNo,
	})
	if err != nil {
		return err
	}
	return nil
}

func (repo *APDepositPaymentRefundTransactionPGRepository) DeleteData(shopID string, docNo string, doc models.APDepositPaymentRefundTransactionPG) error {

	var details *[]models.APDepositPaymentRefundTransactionDetailPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.APDepositPaymentRefundTransactionDetailPG{}).Where(" shopid=? AND docno=?", shopID, docNo).Find(&details)
	for _, tmp := range *details {
		err := tx.Delete(&models.APDepositPaymentRefundTransactionDetailPG{}, tmp.ID).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err := tx.Delete(&models.APDepositPaymentRefundTransactionPG{}, map[string]interface{}{
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
