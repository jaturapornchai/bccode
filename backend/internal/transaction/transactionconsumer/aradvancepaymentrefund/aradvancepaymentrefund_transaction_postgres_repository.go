package aradvancepaymentrefund

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"

	"gorm.io/gorm/clause"
)

type IARAdvancePaymentRefundTransactionPGRepository interface {
	Get(shopID string, docNo string) (*models.ARAdvancePaymentRefundTransactionPG, error)
	Create(doc models.ARAdvancePaymentRefundTransactionPG) error
	Update(shopID string, docNo string, doc models.ARAdvancePaymentRefundTransactionPG) error
	Delete(shopID string, docNo string, doc models.ARAdvancePaymentRefundTransactionPG) error
	DeleteData(shopID string, docNo string, doc models.ARAdvancePaymentRefundTransactionPG) error
}

type ARAdvancePaymentRefundTransactionPGRepository struct {
	pst microservice.IPersister
	repositories.ITransactionConsumerRepository[models.ARAdvancePaymentRefundTransactionPG]
}

func NewARAdvancePaymentRefundTransactionPGRepository(pst microservice.IPersister) IARAdvancePaymentRefundTransactionPGRepository {

	repo := &ARAdvancePaymentRefundTransactionPGRepository{
		pst: pst,
	}

	repo.ITransactionConsumerRepository = repositories.NewTransactionConsumerRepository[models.ARAdvancePaymentRefundTransactionPG](pst)
	return repo
}

func (repo ARAdvancePaymentRefundTransactionPGRepository) Create(doc models.ARAdvancePaymentRefundTransactionPG) error {
	err := repo.pst.Create(&doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo ARAdvancePaymentRefundTransactionPGRepository) Update(shopID string, docNo string, doc models.ARAdvancePaymentRefundTransactionPG) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"shopid": shopID,
		"docno":  docNo,
	})
	if err != nil {
		return err
	}
	return nil
}

func (repo *ARAdvancePaymentRefundTransactionPGRepository) DeleteData(shopID string, docNo string, doc models.ARAdvancePaymentRefundTransactionPG) error {

	var details *[]models.ARAdvancePaymentRefundTransactionDetailPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.ARAdvancePaymentRefundTransactionDetailPG{}).Where(" shopid=? AND docno=?", shopID, docNo).Find(&details)
	for _, tmp := range *details {
		err := tx.Delete(&models.ARAdvancePaymentRefundTransactionDetailPG{}, tmp.ID).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err := tx.Clauses(clause.Returning{}).Delete(&models.ARAdvancePaymentRefundTransactionPG{}, map[string]interface{}{
		"shopid": shopID,
		"docno":  docNo,
	}).Error

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
