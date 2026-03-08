package ardepositpaymentrefund

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IARDepositPaymentRefundTransactionPGRepository interface {
	Get(shopID string, docNo string) (*models.ARDepositPaymentRefundTransactionPG, error)
	Create(doc models.ARDepositPaymentRefundTransactionPG) error
	Update(shopID string, docNo string, doc models.ARDepositPaymentRefundTransactionPG) error
	Delete(shopID string, docNo string, doc models.ARDepositPaymentRefundTransactionPG) error
	DeleteData(shopID string, docNo string, doc models.ARDepositPaymentRefundTransactionPG) error
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

func (repo ARDepositPaymentRefundTransactionPGRepository) Update(shopID string, docNo string, doc models.ARDepositPaymentRefundTransactionPG) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"shopid": shopID,
		"docno":  docNo,
	})
	if err != nil {
		return err
	}
	return nil
}

func (repo *ARDepositPaymentRefundTransactionPGRepository) DeleteData(shopID string, docNo string, doc models.ARDepositPaymentRefundTransactionPG) error {

	var details *[]models.ARDepositPaymentRefundTransactionDetailPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.ARDepositPaymentRefundTransactionDetailPG{}).Where(" shopid=? AND docno=?", shopID, docNo).Find(&details)
	for _, tmp := range *details {
		err := tx.Delete(&models.ARDepositPaymentRefundTransactionDetailPG{}, tmp.ID).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err := tx.Delete(&models.ARDepositPaymentRefundTransactionPG{}, map[string]interface{}{
		"shopid": shopID,
		"docno":  docNo,
	}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
