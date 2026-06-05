package ardepositpayment

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IARDepositPaymentTransactionPGRepository interface {
	Get(holdingCode string, docNo string) (*models.ARDepositPaymentTransactionPG, error)
	Create(doc models.ARDepositPaymentTransactionPG) error
	Update(holdingCode string, docNo string, doc models.ARDepositPaymentTransactionPG) error
	Delete(holdingCode string, docNo string, doc models.ARDepositPaymentTransactionPG) error
	DeleteData(holdingCode string, docNo string, doc models.ARDepositPaymentTransactionPG) error
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

func (repo ARDepositPaymentTransactionPGRepository) Update(holdingCode string, docNo string, doc models.ARDepositPaymentTransactionPG) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"holdingcode": holdingCode,
		"docno":       docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *ARDepositPaymentTransactionPGRepository) DeleteData(holdingCode string, docNo string, doc models.ARDepositPaymentTransactionPG) error {

	var details *[]models.ARDepositPaymentTransactionDetailPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.ARDepositPaymentTransactionDetailPG{}).Where(" holdingcode=? AND docno=?", holdingCode, docNo).Find(&details)
	for _, tmp := range *details {
		err := tx.Delete(&models.ARDepositPaymentTransactionDetailPG{}, tmp.ID).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err := tx.Delete(&models.ARDepositPaymentTransactionPG{}, map[string]interface{}{
		"holdingcode": holdingCode,
		"docno":       docNo,
	}).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
