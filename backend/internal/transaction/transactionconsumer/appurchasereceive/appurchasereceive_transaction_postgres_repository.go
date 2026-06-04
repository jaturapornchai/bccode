package appurchasereceive

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IAPPurchaseReceiveTransactionPostgresRepository interface {
	Get(holdingCode string, docNo string) (*models.APPurchaseReceivePG, error)
	Create(doc models.APPurchaseReceivePG) error
	Update(holdingCode string, docNo string, doc models.APPurchaseReceivePG) error
	Delete(holdingCode string, docNo string, doc models.APPurchaseReceivePG) error
	DeleteData(holdingCode string, docNo string, doc models.APPurchaseReceivePG) error
}

type APPurchaseReceiveTransactionPostgresRepository struct {
	pst microservice.IPersister
	repositories.ITransactionConsumerRepository[models.APPurchaseReceivePG]
}

func NewAPPurchaseReceiveTransactionPostgresRepository(pst microservice.IPersister) IAPPurchaseReceiveTransactionPostgresRepository {

	repo := &APPurchaseReceiveTransactionPostgresRepository{
		pst: pst,
	}

	repo.ITransactionConsumerRepository = repositories.NewTransactionConsumerRepository[models.APPurchaseReceivePG](pst)
	return repo
}

func (repo APPurchaseReceiveTransactionPostgresRepository) Create(doc models.APPurchaseReceivePG) error {
	err := repo.pst.Create(&doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo APPurchaseReceiveTransactionPostgresRepository) Update(holdingCode string, docNo string, doc models.APPurchaseReceivePG) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"holding_code": holdingCode,
		"docno":        docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *APPurchaseReceiveTransactionPostgresRepository) DeleteData(holdingCode string, docNo string, doc models.APPurchaseReceivePG) error {

	var details *[]models.APPurchaseReceiveDetailPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.APPurchaseReceiveDetailPG{}).Where(" holding_code=? AND docno=?", holdingCode, docNo).Find(&details)
	for _, tmp := range *details {
		// mark delete
		err := tx.Delete(&models.APPurchaseReceiveDetailPG{}, tmp.ID).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err := tx.Delete(&models.APPurchaseReceivePG{}, map[string]interface{}{
		"holding_code": holdingCode,
		"docno":        docNo,
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
