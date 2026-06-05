package ardepositpaymentrefund

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IARDepositPaymentRefundTransactionPGRepository interface {
	Get(holdingCode string, docNo string) (*models.ARDepositPaymentRefundTransactionPG, error)
	Create(doc models.ARDepositPaymentRefundTransactionPG) error
	Update(holdingCode string, docNo string, doc models.ARDepositPaymentRefundTransactionPG) error
	Delete(holdingCode string, docNo string, doc models.ARDepositPaymentRefundTransactionPG) error
	DeleteData(holdingCode string, docNo string, doc models.ARDepositPaymentRefundTransactionPG) error
}

type ARDepositPaymentRefundTransactionPGRepository struct {
	pst microservice.IPersister
	repositories.ITransactionConsumerRepository[models.ARDepositPaymentRefundTransactionPG]
}

func NewARDepositPaymentRefundTransactionPGRepository(pst microservice.IPersister) IARDepositPaymentRefundTransactionPGRepository {

	repo := &ARDepositPaymentRefundTransactionPGRepository{
		pst: pst,
	}

	repo.ITransactionConsumerRepository = repositories.NewTransactionConsumerRepository[models.ARDepositPaymentRefundTransactionPG](pst)
	return repo
}

func (repo ARDepositPaymentRefundTransactionPGRepository) Create(doc models.ARDepositPaymentRefundTransactionPG) error {
	err := repo.pst.Create(&doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo ARDepositPaymentRefundTransactionPGRepository) Update(holdingCode string, docNo string, doc models.ARDepositPaymentRefundTransactionPG) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"holdingcode": holdingCode,
		"docno":       docNo,
	})
	if err != nil {
		return err
	}
	return nil
}

func (repo *ARDepositPaymentRefundTransactionPGRepository) DeleteData(holdingCode string, docNo string, doc models.ARDepositPaymentRefundTransactionPG) error {

	var details *[]models.ARDepositPaymentRefundTransactionDetailPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.ARDepositPaymentRefundTransactionDetailPG{}).Where(" holdingcode=? AND docno=?", holdingCode, docNo).Find(&details)
	for _, tmp := range *details {
		err := tx.Delete(&models.ARDepositPaymentRefundTransactionDetailPG{}, tmp.ID).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err := tx.Delete(&models.ARDepositPaymentRefundTransactionPG{}, map[string]interface{}{
		"holdingcode": holdingCode,
		"docno":       docNo,
	}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
