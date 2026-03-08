package ardepositpayment

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IARDepositPaymentTransactionPGRepository interface {
	Get(shopID string, docNo string) (*models.ARDepositPaymentTransactionPG, error)
	Create(doc models.ARDepositPaymentTransactionPG) error
	Update(shopID string, docNo string, doc models.ARDepositPaymentTransactionPG) error
	Delete(shopID string, docNo string, doc models.ARDepositPaymentTransactionPG) error
	DeleteData(shopID string, docNo string, doc models.ARDepositPaymentTransactionPG) error
}

type ARDepositPaymentTransactionPGRepository struct {
	pst microservice.IPersister
	repositories.ITransactionConsumerRepository[models.ARDepositPaymentTransactionPG]
}

func NewARDepositPaymentTransactionPGRepository(pst microservice.IPersister) IARDepositPaymentTransactionPGRepository {

	repo := &ARDepositPaymentTransactionPGRepository{
		pst: pst,
	}

	repo.ITransactionConsumerRepository = repositories.NewTransactionConsumerRepository[models.ARDepositPaymentTransactionPG](pst)
	return repo
}

func (repo ARDepositPaymentTransactionPGRepository) Create(doc models.ARDepositPaymentTransactionPG) error {
	err := repo.pst.Create(&doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo ARDepositPaymentTransactionPGRepository) Update(shopID string, docNo string, doc models.ARDepositPaymentTransactionPG) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"shopid": shopID,
		"docno":  docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *ARDepositPaymentTransactionPGRepository) DeleteData(shopID string, docNo string, doc models.ARDepositPaymentTransactionPG) error {

	var details *[]models.ARDepositPaymentTransactionDetailPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.ARDepositPaymentTransactionDetailPG{}).Where(" shopid=? AND docno=?", shopID, docNo).Find(&details)
	for _, tmp := range *details {
		err := tx.Delete(&models.ARDepositPaymentTransactionDetailPG{}, tmp.ID).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err := tx.Delete(&models.ARDepositPaymentTransactionPG{}, map[string]interface{}{
		"shopid": shopID,
		"docno":  docNo,
	}).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
