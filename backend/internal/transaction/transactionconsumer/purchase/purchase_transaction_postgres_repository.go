package purchase

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IPurchaseTransactionPGRepository interface {
	Get(holdingCode string, docNo string) (*models.PurchaseTransactionPG, error)
	Create(doc models.PurchaseTransactionPG) error
	Update(holdingCode string, docNo string, doc models.PurchaseTransactionPG) error
	Delete(holdingCode string, docNo string, doc models.PurchaseTransactionPG) error
	DeleteData(holdingCode string, docNo string, doc models.PurchaseTransactionPG) error
}

type PurchaseTransactionPGRepository struct {
	pst microservice.IPersister
	repositories.ITransactionConsumerRepository[models.PurchaseTransactionPG]
}

func NewPurchaseTransactionPGRepository(pst microservice.IPersister) IPurchaseTransactionPGRepository {

	repo := &PurchaseTransactionPGRepository{
		pst: pst,
	}

	repo.ITransactionConsumerRepository = repositories.NewTransactionConsumerRepository[models.PurchaseTransactionPG](pst)
	return repo
}

func (repo PurchaseTransactionPGRepository) Create(doc models.PurchaseTransactionPG) error {
	err := repo.pst.Create(&doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo PurchaseTransactionPGRepository) Update(holdingCode string, docNo string, doc models.PurchaseTransactionPG) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"holding_code": holdingCode,
		"docno":        docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *PurchaseTransactionPGRepository) DeleteData(holdingCode string, docNo string, doc models.PurchaseTransactionPG) error {

	var details *[]models.PurchaseTransactionDetailPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.PurchaseTransactionDetailPG{}).Where(" holding_code=? AND docno=?", holdingCode, docNo).Find(&details)
	for _, tmp := range *details {
		// mark delete
		tx.Delete(&models.PurchaseTransactionDetailPG{}, tmp.ID)
	}

	err := tx.Delete(models.PurchaseTransactionPG{}, map[string]interface{}{
		"holding_code": holdingCode,
		"docno":        docNo,
	}).Error

	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}
