package aradvancepayment

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IARAdvancePaymentTransactionPGRepository interface {
	Get(shopID string, docNo string) (*models.ARAdvancePaymentTransactionPG, error)
	Create(doc models.ARAdvancePaymentTransactionPG) error
	Update(shopID string, docNo string, doc models.ARAdvancePaymentTransactionPG) error
	Delete(shopID string, docNo string, doc models.ARAdvancePaymentTransactionPG) error
	DeleteData(shopID string, docNo string, doc models.ARAdvancePaymentTransactionPG) error
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

func (repo ARAdvancePaymentTransactionPGRepository) Update(shopID string, docNo string, doc models.ARAdvancePaymentTransactionPG) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"shopid": shopID,
		"docno":  docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *ARAdvancePaymentTransactionPGRepository) DeleteData(shopID string, docNo string, doc models.ARAdvancePaymentTransactionPG) error {

	var details *[]models.ARAdvancePaymentTransactionDetailPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.ARAdvancePaymentTransactionDetailPG{}).Where(" shopid=? AND docno=?", shopID, docNo).Find(&details)
	for _, tmp := range *details {
		// mark delete
		err := tx.Delete(&models.ARAdvancePaymentTransactionDetailPG{}, tmp.ID).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err := tx.Delete(&models.ARAdvancePaymentTransactionPG{}, "shopid=? AND docno=?", shopID, docNo).Error
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
