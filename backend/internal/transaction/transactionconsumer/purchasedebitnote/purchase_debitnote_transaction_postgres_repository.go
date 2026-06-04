package purchasedebitnote

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IPurchaseDebitNoteTransactionPostgresRepository interface {
	Get(holdingCode string, docNo string) (*models.PurchaseDebitNoteTransactionPG, error)
	Create(doc models.PurchaseDebitNoteTransactionPG) error
	Update(holdingCode string, docNo string, doc models.PurchaseDebitNoteTransactionPG) error
	Delete(holdingCode string, docNo string, doc models.PurchaseDebitNoteTransactionPG) error
	DeleteData(holdingCode string, docNo string, doc models.PurchaseDebitNoteTransactionPG) error
}

type PurchaseDebitNoteTransactionRepository struct {
	pst microservice.IPersister
	repositories.ITransactionConsumerRepository[models.PurchaseDebitNoteTransactionPG]
}

func NewPurchaseDebitNoteTransactionPGRepository(pst microservice.IPersister) IPurchaseDebitNoteTransactionPostgresRepository {

	repo := &PurchaseDebitNoteTransactionRepository{
		pst: pst,
	}

	repo.ITransactionConsumerRepository = repositories.NewTransactionConsumerRepository[models.PurchaseDebitNoteTransactionPG](pst)
	return repo
}

func (repo PurchaseDebitNoteTransactionRepository) Create(doc models.PurchaseDebitNoteTransactionPG) error {
	err := repo.pst.Create(&doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo PurchaseDebitNoteTransactionRepository) Update(holdingCode string, docNo string, doc models.PurchaseDebitNoteTransactionPG) error {
	err := repo.pst.Update(&doc, map[string]interface{}{
		"holding_code": holdingCode,
		"docno":        docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *PurchaseDebitNoteTransactionRepository) DeleteData(holdingCode string, docNo string, doc models.PurchaseDebitNoteTransactionPG) error {
	var details *[]models.PurchaseDebitNoteTransactionDetailPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.PurchaseDebitNoteTransactionDetailPG{}).Where(" holding_code=? AND docno=?", holdingCode, docNo).Find(&details)
	for _, tmp := range *details {
		// mark delete
		err := tx.Delete(&models.PurchaseDebitNoteTransactionDetailPG{}, tmp.ID).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	// mark delete header
	err := tx.Delete(&models.PurchaseDebitNoteTransactionPG{}, "holding_code=? AND docno=?", holdingCode, docNo).Error
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
