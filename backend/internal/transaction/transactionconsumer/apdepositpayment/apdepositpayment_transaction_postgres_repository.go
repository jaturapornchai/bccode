package apdepositpayment

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IAPDepositPaymentTransactionPGRepository interface {
	Get(holdingCode string, docNo string) (*models.APDepositPaymentTransactionPG, error)
	Create(doc models.APDepositPaymentTransactionPG) error
	Update(holdingCode string, docNo string, doc models.APDepositPaymentTransactionPG) error
	Delete(holdingCode string, docNo string, doc models.APDepositPaymentTransactionPG) error
	DeleteData(holdingCode string, docNo string, doc models.APDepositPaymentTransactionPG) error
}

type APDepositPaymentTransactionPGRepository struct {
	pst microservice.IPersister
	repositories.ITransactionConsumerRepository[models.APDepositPaymentTransactionPG]
}

func NewAPDepositPaymentTransactionPGRepository(pst microservice.IPersister) IAPDepositPaymentTransactionPGRepository {

	repo := &APDepositPaymentTransactionPGRepository{
		pst: pst,
	}

	repo.ITransactionConsumerRepository = repositories.NewTransactionConsumerRepository[models.APDepositPaymentTransactionPG](pst)
	return repo
}

func (repo APDepositPaymentTransactionPGRepository) Create(doc models.APDepositPaymentTransactionPG) error {
	err := repo.pst.Create(&doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo APDepositPaymentTransactionPGRepository) Update(holdingCode string, docNo string, doc models.APDepositPaymentTransactionPG) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"holdingcode": holdingCode,
		"docno":       docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *APDepositPaymentTransactionPGRepository) DeleteData(holdingCode string, docNo string, doc models.APDepositPaymentTransactionPG) error {

	var details *[]models.APDepositPaymentTransactionDetailPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.APDepositPaymentTransactionDetailPG{}).Where(" holdingcode=? AND docno=?", holdingCode, docNo).Find(&details)
	for _, tmp := range *details {
		err := tx.Delete(&models.APDepositPaymentTransactionDetailPG{}, tmp.ID).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err := tx.Delete(&models.APDepositPaymentTransactionPG{}, map[string]interface{}{
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
