package appurchasereceive

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IAPPurchaseReceiveTransactionPostgresRepository interface {
	Get(shopID string, docNo string) (*models.APPurchaseReceivePG, error)
	Create(doc models.APPurchaseReceivePG) error
	Update(shopID string, docNo string, doc models.APPurchaseReceivePG) error
	Delete(shopID string, docNo string, doc models.APPurchaseReceivePG) error
	DeleteData(shopID string, docNo string, doc models.APPurchaseReceivePG) error
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

func (repo APPurchaseReceiveTransactionPostgresRepository) Update(shopID string, docNo string, doc models.APPurchaseReceivePG) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"shopid": shopID,
		"docno":  docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *APPurchaseReceiveTransactionPostgresRepository) DeleteData(shopID string, docNo string, doc models.APPurchaseReceivePG) error {

	var details *[]models.APPurchaseReceiveDetailPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.APPurchaseReceiveDetailPG{}).Where(" shopid=? AND docno=?", shopID, docNo).Find(&details)
	for _, tmp := range *details {
		// mark delete
		err := tx.Delete(&models.APPurchaseReceiveDetailPG{}, tmp.ID).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err := tx.Delete(&models.APPurchaseReceivePG{}, map[string]interface{}{
		"shopid": shopID,
		"docno":  docNo,
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
