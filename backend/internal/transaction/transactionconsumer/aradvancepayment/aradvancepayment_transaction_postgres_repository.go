package aradvancepayment

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IARAdvancePaymentTransactionPGRepository interface {
	Get(holdingCode string, docNo string) (*models.ARAdvancePaymentTransactionPG, error)
	Create(doc models.ARAdvancePaymentTransactionPG) error
	Update(holdingCode string, docNo string, doc models.ARAdvancePaymentTransactionPG) error
	Delete(holdingCode string, docNo string, doc models.ARAdvancePaymentTransactionPG) error
	DeleteData(holdingCode string, docNo string, doc models.ARAdvancePaymentTransactionPG) error
}

type ARAdvancePaymentTransactionPGRepository struct {
	pst microservice.IPersister
	repositories.ITransactionConsumerRepository[models.ARAdvancePaymentTransactionPG]
}

func NewARAdvancePaymentTransactionPGRepository(pst microservice.IPersister) IARAdvancePaymentTransactionPGRepository {

	repo := &ARAdvancePaymentTransactionPGRepository{
		pst: pst,
	}

	repo.ITransactionConsumerRepository = repositories.NewTransactionConsumerRepository[models.ARAdvancePaymentTransactionPG](pst)
	return repo
}

func (repo ARAdvancePaymentTransactionPGRepository) Create(doc models.ARAdvancePaymentTransactionPG) error {
	err := repo.pst.Create(&doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo ARAdvancePaymentTransactionPGRepository) Update(holdingCode string, docNo string, doc models.ARAdvancePaymentTransactionPG) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"holdingcode": holdingCode,
		"docno":       docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *ARAdvancePaymentTransactionPGRepository) DeleteData(holdingCode string, docNo string, doc models.ARAdvancePaymentTransactionPG) error {

	var details *[]models.ARAdvancePaymentTransactionDetailPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.ARAdvancePaymentTransactionDetailPG{}).Where(" holdingcode=? AND docno=?", holdingCode, docNo).Find(&details)
	for _, tmp := range *details {
		// mark delete
		err := tx.Delete(&models.ARAdvancePaymentTransactionDetailPG{}, tmp.ID).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err := tx.Delete(&models.ARAdvancePaymentTransactionPG{}, "holdingcode=? AND docno=?", holdingCode, docNo).Error
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
